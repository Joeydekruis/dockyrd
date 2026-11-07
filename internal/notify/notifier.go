package notify

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/events"
	"github.com/Joeydekruis/dockyrd/internal/inbox"
	"github.com/Joeydekruis/dockyrd/internal/instance"
)

// Notifier enqueues typed events and outbox deliveries (no network I/O in Emit).
type Notifier struct {
	DB       *sql.DB
	Q        *sqlc.Queries
	Cfg      config.Config
	Hub      *events.Hub
	Inbox    *inbox.Dispatcher
	Log      *slog.Logger
	Clock    Clock
	wake     chan struct{}
	reminder time.Duration
}

// NewNotifier wires dependencies; wake is shared with Dispatcher.
func NewNotifier(db *sql.DB, q *sqlc.Queries, cfg config.Config, hub *events.Hub, inboxDisp *inbox.Dispatcher, log *slog.Logger, wake chan struct{}) *Notifier {
	if wake == nil {
		wake = make(chan struct{}, 1)
	}
	return &Notifier{
		DB:       db,
		Q:        q,
		Cfg:      cfg,
		Hub:      hub,
		Inbox:    inboxDisp,
		Log:      log,
		Clock:    realClock{},
		wake:     wake,
		reminder: defaultReminderInterval,
	}
}

func (n *Notifier) Wake() chan struct{} { return n.wake }

func (n *Notifier) pingWake() {
	select {
	case n.wake <- struct{}{}:
	default:
	}
}

// PingWake nudges the outbox dispatcher.
func (n *Notifier) PingWake() {
	n.pingWake()
}

// Emit persists a typed event and matching deliveries in one transaction.
func (n *Notifier) Emit(ctx context.Context, ev Event) error {
	if n == nil || n.Q == nil || n.DB == nil {
		return nil
	}
	ev.Type = strings.TrimSpace(ev.Type)
	if ev.Type == "" {
		return errors.New("typed event required")
	}
	ev.Title = strings.TrimSpace(ev.Title)
	ev.Detail = scrubDetail(ev.Detail)
	if ev.Level == "" {
		ev.Level = LevelInfo
	}

	now := n.Clock.Now()
	nowUnix := now.Unix()

	tx, err := n.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := n.Q.WithTx(tx)

	if isStatefulType(ev.Type) {
		if strings.TrimSpace(ev.DedupKey) == "" {
			return errors.New("dedup key required")
		}
		var prev *AlertStateRow
		row, err := qtx.GetAlertState(ctx, ev.DedupKey)
		if err == nil {
			prev = &AlertStateRow{
				State:          row.State,
				Level:          row.Level,
				LastNotifiedAt: row.LastNotifiedAt,
				Occurrences:    row.Occurrences,
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		dec := DecideStateful(prev, ev, now, n.reminder)
		if dec.ResolvedDrop || !dec.Notify {
			if dec.UpdateState && ev.State == StateFiring {
				_ = upsertAlert(ctx, qtx, ev, nowUnix, dec.Occurrences, "")
			}
			if err := tx.Commit(); err != nil {
				return err
			}
			return nil
		}
	}

	eventID := ulid.Make().String()
	dataJSON, _ := json.Marshal(ev.Data)
	if err := qtx.InsertTypedEvent(ctx, sqlc.InsertTypedEventParams{
		ID:             eventID,
		Level:          sql.NullString{String: string(ev.Level), Valid: true},
		Category:       sql.NullString{String: categoryForType(ev.Type), Valid: true},
		Title:          sql.NullString{String: ev.Title, Valid: ev.Title != ""},
		Detail:         sql.NullString{String: ev.Detail, Valid: ev.Detail != ""},
		ProjectID:      nullStr(ev.ProjectID),
		ServiceID:      nullStr(ev.ServiceID),
		DeploymentID:   nullStr(ev.DeploymentID),
		AcknowledgedAt: sql.NullInt64{},
		CreatedAt:      sql.NullInt64{Int64: nowUnix, Valid: true},
		Type:           sql.NullString{String: ev.Type, Valid: true},
		State:          nullStr(string(ev.State)),
		DedupKey:       nullStr(ev.DedupKey),
		Data:           nullStr(string(dataJSON)),
	}); err != nil {
		return err
	}

	if isStatefulType(ev.Type) {
		if err := upsertAlert(ctx, qtx, ev, nowUnix, 1, eventID); err != nil {
			return err
		}
	}

	channels, err := qtx.ListNotificationChannels(ctx)
	if err != nil {
		return err
	}
	for _, ch := range channels {
		if ch.Enabled == 0 || ch.Status == "disabled" {
			continue
		}
		if ev.Type == TypeNotificationChannelFailed {
			if skipID, ok := ev.Data["channel_id"].(string); ok && skipID == ch.ID {
				continue
			}
		}
		if !channelMatchesEvent(ch, ev) {
			continue
		}
		skipInbox := n.inboxEmailSkipSet(ctx, ev)
		targets := n.deliveryTargets(ctx, ch, skipInbox)
		if len(targets) == 0 {
			continue
		}
		dedup := ev.DedupKey
		if dedup != "" {
			_ = qtx.SkipPendingDeliveriesByDedup(ctx, sqlc.SkipPendingDeliveriesByDedupParams{
				LastError: sql.NullString{String: "superseded", Valid: true},
				ChannelID: ch.ID,
				DedupKey:  nullStr(dedup),
			})
		}
		for _, target := range targets {
			deliveryID := ulid.Make().String()
			if err := qtx.InsertNotificationDelivery(ctx, sqlc.InsertNotificationDeliveryParams{
				ID:            deliveryID,
				EventID:       eventID,
				ChannelID:     ch.ID,
				Target:        target,
				DedupKey:      nullStr(dedup),
				Status:        "pending",
				Attempts:      0,
				NextAttemptAt: nowUnix,
				LeaseUntil:    sql.NullInt64{},
				HttpStatus:    sql.NullInt64{},
				LastError:     sql.NullString{},
				CreatedAt:     nowUnix,
				SentAt:        sql.NullInt64{},
			}); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	n.pingWake()
	if n.Hub != nil {
		n.Hub.Publish("events", map[string]any{
			"id":       eventID,
			"level":    ev.Level,
			"category": categoryForType(ev.Type),
			"title":    ev.Title,
			"type":     ev.Type,
		})
	}
	n.fanoutInbox(ctx, ev)
	return nil
}

func upsertAlert(ctx context.Context, q *sqlc.Queries, ev Event, now int64, occ int64, eventID string) error {
	var resolved sql.NullInt64
	if ev.State == StateResolved {
		resolved = sql.NullInt64{Int64: now, Valid: true}
	}
	return q.UpsertAlertState(ctx, sqlc.UpsertAlertStateParams{
		DedupKey:       ev.DedupKey,
		Type:           ev.Type,
		State:          string(ev.State),
		Level:          string(ev.Level),
		FirstFiredAt:   now,
		LastNotifiedAt: now,
		ResolvedAt:     resolved,
		Occurrences:    occ,
		EventID:        nullStr(eventID),
	})
}

func channelMatchesEvent(ch sqlc.NotificationChannel, ev Event) bool {
	minLevel := Level(ch.MinLevel)
	if !LevelMeetsMin(minLevel, ev.Level) {
		return false
	}
	if ch.ProjectID.Valid && ch.ProjectID.String != "" && ev.ProjectID != "" && ch.ProjectID.String != ev.ProjectID {
		return false
	}
	if ev.Type == TypeNotificationChannelFailed {
		if cid, ok := ev.Data["channel_id"].(string); ok && cid == ch.ID {
			return false
		}
	}
	types := parseEventTypes(ch.EventTypes)
	if len(types) == 0 {
		types = DefaultExternalEventTypes()
	}
	for _, t := range types {
		if t == ev.Type {
			return true
		}
	}
	return false
}

func parseEventTypes(raw string) []string {
	var out []string
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func (n *Notifier) fanoutInbox(ctx context.Context, ev Event) {
	if n.Inbox == nil {
		return
	}
	switch ev.Type {
	case TypeDeployFailed:
		n.Inbox.NotifyDeployFailed(ctx, ev.ServiceID, ev.DeploymentID, ev.Detail)
	case TypeDeploySucceeded:
		n.Inbox.NotifyDeployLive(ctx, ev.ServiceID, ev.DeploymentID)
	}
}

// EmitDeployFailed builds and emits deploy.failed.
func (n *Notifier) EmitDeployFailed(ctx context.Context, serviceID, deploymentID, detail string) {
	if n == nil {
		return
	}
	ctxRow, err := n.Q.GetServiceDeployContext(ctx, serviceID)
	if err != nil {
		return
	}
	title := "Deploy failed · " + strings.ToUpper(ctxRow.ProjectSlug+" / "+ctxRow.ServiceSlug)
	href := deployHref(ctxRow.ServiceKind, ctxRow.ServiceSlug, deploymentID)
	ev := Event{
		Type:         TypeDeployFailed,
		Level:        LevelError,
		Title:        title,
		Detail:       detail,
		ProjectID:    ctxRow.ProjectID,
		ServiceID:    serviceID,
		DeploymentID: deploymentID,
		Href:         href,
	}
	if err := n.Emit(ctx, ev); err != nil && n.Log != nil {
		n.Log.Warn("notify emit deploy.failed", "err", err)
	}
}

// EmitDeploySucceeded emits deploy.succeeded (usually filtered off channels).
func (n *Notifier) EmitDeploySucceeded(ctx context.Context, serviceID, deploymentID string) {
	if n == nil {
		return
	}
	ctxRow, err := n.Q.GetServiceDeployContext(ctx, serviceID)
	if err != nil {
		return
	}
	title := "Deploy succeeded · " + strings.ToUpper(ctxRow.ProjectSlug+" / "+ctxRow.ServiceSlug)
	ev := Event{
		Type:         TypeDeploySucceeded,
		Level:        LevelSuccess,
		Title:        title,
		ProjectID:    ctxRow.ProjectID,
		ServiceID:    serviceID,
		DeploymentID: deploymentID,
		Href:         deployHref(ctxRow.ServiceKind, ctxRow.ServiceSlug, deploymentID),
	}
	if err := n.Emit(ctx, ev); err != nil && n.Log != nil {
		n.Log.Warn("notify emit deploy.succeeded", "err", err)
	}
}

func deployHref(kind, slug, deploymentID string) string {
	switch kind {
	case "static_site":
		return "/static-sites/" + slug + "/deployments/" + deploymentID
	case "app":
		if deploymentID != "" {
			return "/applications/" + slug + "/deployments/" + deploymentID
		}
		return "/applications/" + slug
	default:
		return ""
	}
}

func nullStr(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func (n *Notifier) inboxEmailSkipSet(ctx context.Context, ev Event) map[string]struct{} {
	if ev.Type != TypeDeployFailed || n.Inbox == nil {
		return nil
	}
	return n.Inbox.InboxEmailRecipientsForDeployFailed(ctx)
}

func (n *Notifier) deliveryTargets(ctx context.Context, ch sqlc.NotificationChannel, skipInbox map[string]struct{}) []string {
	if ch.Kind != "email" {
		return []string{""}
	}
	cfg := parseChannelConfig(ch.Config)
	targets, err := ResolveEmailTargets(ctx, n.Q, cfg, skipInbox)
	if err != nil || len(targets) == 0 {
		return nil
	}
	return targets
}

// DecryptChannelSecret decrypts secret_enc for senders.
func DecryptChannelSecret(cfg config.Config, blob []byte) (ChannelSecret, error) {
	if len(blob) == 0 {
		return ChannelSecret{}, nil
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return ChannelSecret{}, err
	}
	plain, err := crypto.Decrypt(key, blob)
	if err != nil {
		return ChannelSecret{}, err
	}
	var out ChannelSecret
	if err := json.Unmarshal(plain, &out); err != nil {
		return ChannelSecret{}, err
	}
	return out, nil
}

// EncryptChannelSecret seals secret JSON for storage.
func EncryptChannelSecret(cfg config.Config, secret ChannelSecret) ([]byte, error) {
	raw, err := json.Marshal(secret)
	if err != nil {
		return nil, err
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return nil, err
	}
	return crypto.Encrypt(key, raw)
}

// PublicBaseURL loads panel public URL for renderers.
func PublicBaseURL(ctx context.Context, cfg config.Config, q *sqlc.Queries) string {
	panel, _ := instance.LoadPanel(ctx, q)
	return instance.PublicBaseURL(cfg, panel)
}
