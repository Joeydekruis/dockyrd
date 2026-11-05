package notify

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

const (
	deliveryLeaseSec = 60
	deliveryTTL      = 24 * time.Hour
	maxConcurrent    = 4
	claimBatch       = 8
)

// Dispatcher sends outbox deliveries with retries and circuit-breaking.
type Dispatcher struct {
	Q        *sqlc.Queries
	Cfg      config.Config
	Log      *slog.Logger
	Registry Registry
	Notifier *Notifier
	Clock    Clock
	Version  string

	wake chan struct{}

	channelMu  sync.Map // channelID -> *sync.Mutex
	inflight   int
	inflightMu sync.Mutex
}

func NewDispatcher(q *sqlc.Queries, cfg config.Config, log *slog.Logger, reg Registry, notifier *Notifier, version string, wake chan struct{}) *Dispatcher {
	return &Dispatcher{
		Q:        q,
		Cfg:      cfg,
		Log:      log,
		Registry: reg,
		Notifier: notifier,
		Clock:    realClock{},
		Version:  version,
		wake:     wake,
	}
}

// Run processes the outbox until ctx is cancelled.
func (d *Dispatcher) Run(ctx context.Context, tick time.Duration) {
	if d == nil || d.Q == nil {
		return
	}
	if tick <= 0 {
		tick = 2 * time.Second
	}
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		d.recoverLeases(ctx)
		d.processDue(ctx)
		select {
		case <-ctx.Done():
			return
		case <-d.wake:
			d.processDue(ctx)
		case <-ticker.C:
		}
	}
}

func (d *Dispatcher) recoverLeases(ctx context.Context) {
	now := d.Clock.Now().Unix()
	_ = d.Q.RecoverExpiredDeliveryLeases(ctx, sql.NullInt64{Int64: now, Valid: true})
}

func (d *Dispatcher) processDue(ctx context.Context) {
	now := d.Clock.Now().Unix()
	rows, err := d.Q.ListDueDeliveries(ctx, sqlc.ListDueDeliveriesParams{
		NextAttemptAt: now,
		Limit:         claimBatch,
	})
	if err != nil {
		return
	}
	for _, row := range rows {
		if !d.acquireSlot() {
			break
		}
		go d.deliverOne(ctx, row)
	}
}

func (d *Dispatcher) acquireSlot() bool {
	d.inflightMu.Lock()
	defer d.inflightMu.Unlock()
	if d.inflight >= maxConcurrent {
		return false
	}
	d.inflight++
	return true
}

func (d *Dispatcher) releaseSlot() {
	d.inflightMu.Lock()
	d.inflight--
	d.inflightMu.Unlock()
}

func (d *Dispatcher) channelLock(id string) *sync.Mutex {
	v, _ := d.channelMu.LoadOrStore(id, &sync.Mutex{})
	return v.(*sync.Mutex)
}

func (d *Dispatcher) deliverOne(ctx context.Context, row sqlc.ListDueDeliveriesRow) {
	defer d.releaseSlot()
	mu := d.channelLock(row.ChannelID)
	mu.Lock()
	defer mu.Unlock()

	lease := d.Clock.Now().Unix() + deliveryLeaseSec
	if err := d.Q.MarkDeliverySending(ctx, sqlc.MarkDeliverySendingParams{
		LeaseUntil: sql.NullInt64{Int64: lease, Valid: true},
		ID:         row.ID,
	}); err != nil {
		return
	}

	ch, err := d.Q.GetNotificationChannel(ctx, row.ChannelID)
	if err != nil {
		return
	}
	sender := d.Registry[ch.Kind]
	if sender == nil {
		_ = d.Q.MarkDeliveryDead(ctx, sqlc.MarkDeliveryDeadParams{
			HttpStatus: sql.NullInt64{Int64: 0, Valid: true},
			LastError:  sql.NullString{String: "unknown channel kind", Valid: true},
			ID:         row.ID,
		})
		return
	}

	evRow, err := d.Q.GetEventForNotify(ctx, row.EventID)
	if err != nil {
		return
	}
	evView := eventViewFromRow(evRow)
	cfg := parseChannelConfig(ch.Config)
	secret, err := DecryptChannelSecret(d.Cfg, ch.SecretEnc)
	if err != nil {
		d.failDelivery(ctx, row.ID, ch, 0, "decrypt secret", false, false)
		return
	}
	allowPrivate := loadAllowPrivateTargets(ctx, d.Q)
	base := PublicBaseURL(ctx, d.Cfg, d.Q)
	rendered, err := sender.Render(evView, cfg, base)
	if err != nil {
		d.failDelivery(ctx, row.ID, ch, 400, err.Error(), true, false)
		return
	}
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	meta := DeliveryMeta{
		DeliveryID: row.ID,
		Target:     row.Target,
		EventType:  evView.Type,
		Instance:   instanceName(ctx, d.Q),
		Version:    d.Version,
	}
	res := sender.Send(sendCtx, cfg, secret, rendered, meta, allowPrivate)
	d.applyResult(ctx, row.ID, ch, res)
}

func (d *Dispatcher) applyResult(ctx context.Context, deliveryID string, ch sqlc.NotificationChannel, res Result) {
	now := d.Clock.Now().Unix()
	if res.OK {
		_ = d.Q.MarkDeliverySent(ctx, sqlc.MarkDeliverySentParams{
			SentAt:     sql.NullInt64{Int64: now, Valid: true},
			HttpStatus: sql.NullInt64{Int64: int64(res.HTTPStatus), Valid: res.HTTPStatus != 0},
			ID:         deliveryID,
		})
		d.updateChannelHealth(ctx, ch, true, "", res.HTTPStatus, false, "")
		return
	}
	if res.RetryAfter > 0 && res.HTTPStatus == 429 {
		next := now + int64(res.RetryAfter.Seconds())
		_ = d.Q.MarkDeliveryRetry(ctx, sqlc.MarkDeliveryRetryParams{
			NextAttemptAt: next,
			HttpStatus:    sql.NullInt64{Int64: int64(res.HTTPStatus), Valid: true},
			LastError:     sql.NullString{String: res.Err, Valid: res.Err != ""},
			ID:            deliveryID,
		})
		return
	}
	if res.Permanent || res.DisableChannel {
		_ = d.Q.MarkDeliveryDead(ctx, sqlc.MarkDeliveryDeadParams{
			HttpStatus: sql.NullInt64{Int64: int64(res.HTTPStatus), Valid: res.HTTPStatus != 0},
			LastError:  sql.NullString{String: res.Err, Valid: res.Err != ""},
			ID:         deliveryID,
		})
		reason := res.Err
		if res.DisableChannel {
			reason = "webhook unavailable"
		}
		d.updateChannelHealth(ctx, ch, false, res.Err, res.HTTPStatus, res.DisableChannel, reason)
		return
	}
	// retry with backoff
	del, err := d.Q.GetNotificationDelivery(ctx, deliveryID)
	if err != nil {
		return
	}
	if d.Clock.Now().Sub(time.Unix(del.CreatedAt, 0)) > deliveryTTL {
		_ = d.Q.MarkDeliveryDead(ctx, sqlc.MarkDeliveryDeadParams{
			HttpStatus: sql.NullInt64{Int64: int64(res.HTTPStatus), Valid: true},
			LastError:  sql.NullString{String: "expired", Valid: true},
			ID:         deliveryID,
		})
		return
	}
	next := now + int64(backoff(del.Attempts).Seconds())
	_ = d.Q.MarkDeliveryRetry(ctx, sqlc.MarkDeliveryRetryParams{
		NextAttemptAt: next,
		HttpStatus:    sql.NullInt64{Int64: int64(res.HTTPStatus), Valid: res.HTTPStatus != 0},
		LastError:     sql.NullString{String: res.Err, Valid: res.Err != ""},
		ID:            deliveryID,
	})
	d.updateChannelHealth(ctx, ch, false, res.Err, res.HTTPStatus, false, "")
}

func (d *Dispatcher) failDelivery(ctx context.Context, deliveryID string, ch sqlc.NotificationChannel, status int, msg string, permanent, disable bool) {
	res := Result{HTTPStatus: status, Err: msg, Permanent: permanent, DisableChannel: disable}
	d.applyResult(ctx, deliveryID, ch, res)
}

func (d *Dispatcher) updateChannelHealth(ctx context.Context, ch sqlc.NotificationChannel, ok bool, errMsg string, httpStatus int, disable bool, disableReason string) {
	now := d.Clock.Now().Unix()
	failures := ch.ConsecutiveFailures
	var lastSuccess, lastFailure sql.NullInt64
	status := ch.Status
	disabledReason := ch.DisabledReason
	if ok {
		failures = 0
		lastSuccess = sql.NullInt64{Int64: now, Valid: true}
		if status == "degraded" {
			status = "active"
			disabledReason = sql.NullString{}
		}
	} else {
		failures++
		lastFailure = sql.NullInt64{Int64: now, Valid: true}
		if failures >= 5 && status == "active" {
			status = "degraded"
		}
		if disable {
			status = "disabled"
			disabledReason = sql.NullString{String: disableReason, Valid: disableReason != ""}
			d.emitChannelFailed(ctx, ch)
		}
	}
	_ = d.Q.UpdateChannelDeliveryHealth(ctx, sqlc.UpdateChannelDeliveryHealthParams{
		ConsecutiveFailures: failures,
		LastSuccessAt:       lastSuccess,
		LastFailureAt:       lastFailure,
		LastError:           sql.NullString{String: errMsg, Valid: errMsg != "" && !ok},
		Status:              status,
		DisabledReason:      disabledReason,
		UpdatedAt:           now,
		ID:                  ch.ID,
	})
	_ = httpStatus
}

func (d *Dispatcher) emitChannelFailed(ctx context.Context, ch sqlc.NotificationChannel) {
	if d.Notifier == nil {
		return
	}
	ev := Event{
		Type:     TypeNotificationChannelFailed,
		Level:    LevelWarning,
		State:    StateFiring,
		DedupKey: "channel:" + ch.ID,
		Title:    "Alert channel failing · " + ch.Name,
		Detail:   "Deliveries to this channel failed repeatedly.",
		Href:     "/settings/server/channels",
		Data:     map[string]any{"channel_id": ch.ID, "kind": ch.Kind},
	}
	_ = d.Notifier.Emit(ctx, ev)
}

func backoff(attempts int64) time.Duration {
	switch attempts {
	case 1:
		return jitter(30 * time.Second)
	case 2:
		return jitter(2 * time.Minute)
	case 3:
		return jitter(10 * time.Minute)
	case 4:
		return jitter(time.Hour)
	default:
		return jitter(6 * time.Hour)
	}
}

func jitter(d time.Duration) time.Duration {
	f := 0.8 + rand.Float64()*0.4
	return time.Duration(float64(d) * f)
}

func eventViewFromRow(row sqlc.Event) EventView {
	ev := EventView{
		ID:     row.ID,
		Type:   row.Type.String,
		Level:  Level(row.Level.String),
		Title:  row.Title.String,
		Detail: row.Detail.String,
	}
	if row.State.Valid {
		ev.State = State(row.State.String)
	}
	if row.ProjectID.Valid {
		ev.ProjectID = row.ProjectID.String
	}
	if row.ServiceID.Valid {
		ev.ServiceID = row.ServiceID.String
	}
	if row.DeploymentID.Valid {
		ev.DeploymentID = row.DeploymentID.String
	}
	if row.CreatedAt.Valid {
		ev.OccurredAt = time.Unix(row.CreatedAt.Int64, 0)
	}
	if row.Data.Valid && row.Data.String != "" {
		_ = json.Unmarshal([]byte(row.Data.String), &ev.Data)
	}
	return ev
}

func parseChannelConfig(raw string) ChannelConfig {
	var cfg ChannelConfig
	_ = json.Unmarshal([]byte(raw), &cfg)
	if cfg == nil {
		cfg = ChannelConfig{}
	}
	return cfg
}

func loadAllowPrivateTargets(ctx context.Context, q *sqlc.Queries) bool {
	row, err := q.GetSetting(ctx, "notifications")
	if err != nil {
		return false
	}
	var settings struct {
		AllowPrivateTargets bool `json:"allow_private_targets"`
	}
	raw, _ := json.Marshal(row.Value)
	_ = json.Unmarshal(raw, &settings)
	return settings.AllowPrivateTargets
}

func instanceName(ctx context.Context, q *sqlc.Queries) string {
	row, err := q.GetSetting(ctx, "notifications")
	if err != nil {
		return "Dockyard"
	}
	var settings struct {
		InstanceName string `json:"instance_name"`
	}
	raw, _ := json.Marshal(row.Value)
	_ = json.Unmarshal(raw, &settings)
	if settings.InstanceName != "" {
		return settings.InstanceName
	}
	return "Dockyard"
}

// SendTest delivers synchronously through a sender (API test button).
func SendTest(ctx context.Context, cfg config.Config, q *sqlc.Queries, reg Registry, ch sqlc.NotificationChannel, version string) Result {
	sender := reg[ch.Kind]
	if sender == nil {
		return Result{Permanent: true, Err: "unsupported kind"}
	}
	secret, err := DecryptChannelSecret(cfg, ch.SecretEnc)
	if err != nil {
		return Result{Err: "decrypt failed"}
	}
	channelCfg := parseChannelConfig(ch.Config)
	if err := sender.Validate(channelCfg, secret, loadAllowPrivateTargets(ctx, q)); err != nil {
		return Result{Permanent: true, Err: err.Error()}
	}
	now := time.Now()
	ev := EventView{
		ID:         ulid.Make().String(),
		Type:       TypeTestPing,
		Level:      LevelInfo,
		Title:      "Test notification",
		Detail:     "This is a test from Dockyrd.",
		OccurredAt: now,
	}
	base := PublicBaseURL(ctx, cfg, q)
	rendered, err := sender.Render(ev, channelCfg, base)
	if err != nil {
		return Result{Permanent: true, Err: err.Error()}
	}
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	meta := DeliveryMeta{DeliveryID: ev.ID, EventType: TypeTestPing, Version: version}
	if ch.Kind == "email" {
		skip := map[string]struct{}{}
		targets, err := ResolveEmailTargets(ctx, q, channelCfg, skip)
		if err != nil || len(targets) == 0 {
			return Result{Permanent: true, Err: "no email recipients configured"}
		}
		meta.Target = targets[0]
	}
	return sender.Send(sendCtx, channelCfg, secret, rendered, meta, loadAllowPrivateTargets(ctx, q))
}
