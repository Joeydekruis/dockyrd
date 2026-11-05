package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/notify"
)

func registerNotificationRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/settings/notifications", deps.handleGetNotificationSettings)
	mux.HandleFunc("PATCH /api/v1/settings/notifications", deps.handlePatchNotificationSettings)
	mux.HandleFunc("GET /api/v1/settings/notifications/channels", deps.handleListNotificationChannels)
	mux.HandleFunc("POST /api/v1/settings/notifications/channels", deps.handleCreateNotificationChannel)
	mux.HandleFunc("GET /api/v1/settings/notifications/channels/{id}", deps.handleGetNotificationChannel)
	mux.HandleFunc("PATCH /api/v1/settings/notifications/channels/{id}", deps.handlePatchNotificationChannel)
	mux.HandleFunc("DELETE /api/v1/settings/notifications/channels/{id}", deps.handleDeleteNotificationChannel)
	mux.HandleFunc("POST /api/v1/settings/notifications/channels/{id}/test", deps.handleTestNotificationChannel)
	mux.HandleFunc("GET /api/v1/settings/notifications/channels/{id}/deliveries", deps.handleListNotificationDeliveries)
	mux.HandleFunc("POST /api/v1/settings/notifications/deliveries/{id}/retry", deps.handleRetryNotificationDelivery)
}

type notificationSettings struct {
	DiskWarnPct         int    `json:"disk_warn_pct"`
	DiskCritPct         int    `json:"disk_crit_pct"`
	CertWarnPct         int    `json:"cert_warn_pct"`
	ReminderHours       int    `json:"reminder_hours"`
	InstanceName        string `json:"instance_name"`
	AllowPrivateTargets bool   `json:"allow_private_targets"`
	HeartbeatURL        string `json:"heartbeat_url"`
}

func defaultNotificationSettings() notificationSettings {
	return notificationSettings{
		DiskWarnPct:   85,
		DiskCritPct:   95,
		CertWarnPct:   15,
		ReminderHours: 24,
	}
}

func (d apiDeps) loadNotificationSettings(ctx context.Context) notificationSettings {
	out := defaultNotificationSettings()
	if d.q == nil {
		return out
	}
	row, err := d.q.GetSetting(ctx, "notifications")
	if err != nil {
		return out
	}
	raw, err := settingValueBytes(row.Value)
	if err != nil {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	if out.DiskWarnPct == 0 {
		out.DiskWarnPct = 85
	}
	if out.DiskCritPct == 0 {
		out.DiskCritPct = 95
	}
	if out.ReminderHours == 0 {
		out.ReminderHours = 24
	}
	return out
}

func (d apiDeps) saveNotificationSettings(ctx context.Context, s notificationSettings) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return d.q.UpsertSetting(ctx, sqlc.UpsertSettingParams{
		Key:   "notifications",
		Value: raw,
	})
}

func (d apiDeps) handleGetNotificationSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	ctx := r.Context()
	settings := d.loadNotificationSettings(ctx)
	smtp, _ := notify.SMTPConfigured(ctx, d.q)
	writeJSON(w, http.StatusOK, map[string]any{
		"settings":        settings,
		"event_types":     notify.AllEventTypes(),
		"smtp_configured": smtp,
	})
}

func (d apiDeps) handlePatchNotificationSettings(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	var body struct {
		DiskWarnPct         *int    `json:"disk_warn_pct"`
		DiskCritPct         *int    `json:"disk_crit_pct"`
		CertWarnPct         *int    `json:"cert_warn_pct"`
		ReminderHours       *int    `json:"reminder_hours"`
		InstanceName        *string `json:"instance_name"`
		AllowPrivateTargets *bool   `json:"allow_private_targets"`
		HeartbeatURL        *string `json:"heartbeat_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ctx := r.Context()
	cur := d.loadNotificationSettings(ctx)
	if body.DiskWarnPct != nil && *body.DiskWarnPct > 0 {
		cur.DiskWarnPct = *body.DiskWarnPct
	}
	if body.DiskCritPct != nil && *body.DiskCritPct > 0 {
		cur.DiskCritPct = *body.DiskCritPct
	}
	if body.CertWarnPct != nil && *body.CertWarnPct > 0 {
		cur.CertWarnPct = *body.CertWarnPct
	}
	if body.ReminderHours != nil && *body.ReminderHours > 0 {
		cur.ReminderHours = *body.ReminderHours
	}
	if body.InstanceName != nil {
		cur.InstanceName = strings.TrimSpace(*body.InstanceName)
	}
	if user.Role == "owner" && body.AllowPrivateTargets != nil {
		cur.AllowPrivateTargets = *body.AllowPrivateTargets
	}
	if body.HeartbeatURL != nil {
		cur.HeartbeatURL = strings.TrimSpace(*body.HeartbeatURL)
	}
	if err := d.saveNotificationSettings(ctx, cur); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	d.audit.Record(ctx, audit.Entry{
		Action: "notification_settings.update",
		UserID: user.ID,
	})
	d.handleGetNotificationSettings(w, r)
}

type channelResponse struct {
	ID                  string         `json:"id"`
	Kind                string         `json:"kind"`
	Name                string         `json:"name"`
	Enabled             bool           `json:"enabled"`
	Config              map[string]any `json:"config"`
	SecretSet           bool           `json:"secret_set"`
	SecretHint          string         `json:"secret_hint,omitempty"`
	MinLevel            string         `json:"min_level"`
	EventTypes          []string       `json:"event_types"`
	ProjectID           *string        `json:"project_id"`
	Status              string         `json:"status"`
	DisabledReason      string         `json:"disabled_reason,omitempty"`
	ConsecutiveFailures int64          `json:"consecutive_failures"`
	LastSuccessAt       *int64         `json:"last_success_at,omitempty"`
	LastFailureAt       *int64         `json:"last_failure_at,omitempty"`
	LastError           string         `json:"last_error,omitempty"`
	CreatedAt           int64          `json:"created_at"`
}

func channelToResponse(ch sqlc.NotificationChannel, hint string) channelResponse {
	var cfg map[string]any
	_ = json.Unmarshal([]byte(ch.Config), &cfg)
	types := notify.ParseEventTypesJSON(ch.EventTypes)
	out := channelResponse{
		ID:                  ch.ID,
		Kind:                ch.Kind,
		Name:                ch.Name,
		Enabled:             ch.Enabled != 0,
		Config:              cfg,
		SecretSet:           len(ch.SecretEnc) > 0,
		SecretHint:          hint,
		MinLevel:            ch.MinLevel,
		EventTypes:          types,
		Status:              ch.Status,
		ConsecutiveFailures: ch.ConsecutiveFailures,
		CreatedAt:           ch.CreatedAt,
	}
	if ch.ProjectID.Valid {
		out.ProjectID = &ch.ProjectID.String
	}
	if ch.DisabledReason.Valid {
		out.DisabledReason = ch.DisabledReason.String
	}
	if ch.LastSuccessAt.Valid {
		v := ch.LastSuccessAt.Int64
		out.LastSuccessAt = &v
	}
	if ch.LastFailureAt.Valid {
		v := ch.LastFailureAt.Int64
		out.LastFailureAt = &v
	}
	if ch.LastError.Valid {
		out.LastError = ch.LastError.String
	}
	return out
}

func (d apiDeps) handleListNotificationChannels(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	rows, err := d.q.ListNotificationChannels(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]channelResponse, 0, len(rows))
	for _, ch := range rows {
		hint, _ := notify.SecretHint(d.cfg, ch)
		out = append(out, channelToResponse(ch, hint))
	}
	writeJSON(w, http.StatusOK, map[string]any{"channels": out})
}

func (d apiDeps) handleCreateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	var body struct {
		Kind       string         `json:"kind"`
		Name       string         `json:"name"`
		Enabled    *bool          `json:"enabled"`
		Config     map[string]any `json:"config"`
		Secret     map[string]any `json:"secret"`
		MinLevel   string         `json:"min_level"`
		EventTypes []string       `json:"event_types"`
		ProjectID  string         `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Kind = strings.TrimSpace(body.Kind)
	if body.Kind != "slack" && body.Kind != "discord" && body.Kind != "webhook" && body.Kind != "email" {
		writeJSONError(w, http.StatusUnprocessableEntity, "unsupported kind")
		return
	}
	count, _ := d.q.CountNotificationChannels(r.Context())
	if count >= 20 {
		writeJSONError(w, http.StatusBadRequest, "channel limit reached")
		return
	}
	if body.Kind == "email" {
		if body.Config == nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "validation failed", "fields": map[string]string{"config": "required"}})
			return
		}
		if err := notify.ValidateEmailChannelConfig(notify.ChannelConfig(body.Config)); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "validation failed", "fields": map[string]string{"config": err.Error()}})
			return
		}
	} else if fields := validateChannelSecret(d, body.Kind, body.Secret); len(fields) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "validation failed", "fields": fields})
		return
	}
	now := time.Now().Unix()
	id := ulid.Make().String()
	cfgJSON, _ := json.Marshal(body.Config)
	if body.Config == nil {
		cfgJSON = []byte("{}")
	}
	types := body.EventTypes
	if len(types) == 0 {
		types = notify.DefaultExternalEventTypes()
	}
	typesJSON, _ := json.Marshal(types)
	minLevel := body.MinLevel
	if minLevel == "" {
		minLevel = "warning"
	}
	var enc []byte
	var encErr error
	if body.Kind != "email" {
		enc, encErr = notify.EncryptChannelSecret(d.cfg, notify.ChannelSecret(body.Secret))
		if encErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not encrypt secret")
			return
		}
	}
	enabled := int64(1)
	if body.Enabled != nil && !*body.Enabled {
		enabled = 0
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = body.Kind
	}
	err := d.q.InsertNotificationChannel(r.Context(), sqlc.InsertNotificationChannelParams{
		ID:                  id,
		Kind:                body.Kind,
		Name:                name,
		Enabled:             enabled,
		Config:              string(cfgJSON),
		SecretEnc:           enc,
		MinLevel:            minLevel,
		EventTypes:          string(typesJSON),
		ProjectID:           nullOptionalString(body.ProjectID),
		Status:              "active",
		DisabledReason:      sql.NullString{},
		ConsecutiveFailures: 0,
		LastSuccessAt:       sql.NullInt64{},
		LastFailureAt:       sql.NullInt64{},
		LastError:           sql.NullString{},
		CreatedBy:           sql.NullString{String: user.ID, Valid: true},
		CreatedAt:           now,
		UpdatedAt:           now,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	d.audit.Record(r.Context(), audit.Entry{Action: "notification_channel.create", UserID: user.ID, Meta: map[string]any{"id": id}})
	ch, _ := d.q.GetNotificationChannel(r.Context(), id)
	hint, _ := notify.SecretHint(d.cfg, ch)
	writeJSON(w, http.StatusCreated, channelToResponse(ch, hint))
}

func validateChannelSecret(d apiDeps, kind string, secret map[string]any) map[string]string {
	if kind == "email" {
		return nil
	}
	reg := d.notifyRegistry()
	s := reg[kind]
	if s == nil {
		return map[string]string{"kind": "unsupported"}
	}
	allow := d.loadNotificationSettings(context.Background()).AllowPrivateTargets
	if err := s.Validate(notify.ChannelConfig{}, notify.ChannelSecret(secret), allow); err != nil {
		return map[string]string{"url": err.Error()}
	}
	return nil
}

func (d apiDeps) handleGetNotificationChannel(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	ch, err := d.q.GetNotificationChannel(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	hint, _ := notify.SecretHint(d.cfg, ch)
	writeJSON(w, http.StatusOK, channelToResponse(ch, hint))
}

func (d apiDeps) handlePatchNotificationChannel(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	ctx := r.Context()
	ch, err := d.q.GetNotificationChannel(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	var body struct {
		Name       *string        `json:"name"`
		Enabled    *bool          `json:"enabled"`
		Config     map[string]any `json:"config"`
		Secret     map[string]any `json:"secret"`
		MinLevel   *string        `json:"min_level"`
		EventTypes []string       `json:"event_types"`
		ProjectID  *string        `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	name := ch.Name
	if body.Name != nil {
		name = strings.TrimSpace(*body.Name)
	}
	enabled := ch.Enabled
	if body.Enabled != nil {
		if *body.Enabled {
			enabled = 1
		} else {
			enabled = 0
		}
	}
	cfgStr := ch.Config
	if body.Config != nil {
		raw, _ := json.Marshal(body.Config)
		cfgStr = string(raw)
		if ch.Kind == "email" {
			if err := notify.ValidateEmailChannelConfig(notify.ChannelConfig(body.Config)); err != nil {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "validation failed", "fields": map[string]string{"config": err.Error()}})
				return
			}
		}
	}
	secretEnc := ch.SecretEnc
	if body.Secret != nil && len(body.Secret) > 0 {
		if fields := validateChannelSecret(d, ch.Kind, body.Secret); len(fields) > 0 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "validation failed", "fields": fields})
			return
		}
		enc, err := notify.EncryptChannelSecret(d.cfg, notify.ChannelSecret(body.Secret))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "encrypt failed")
			return
		}
		secretEnc = enc
	}
	minLevel := ch.MinLevel
	if body.MinLevel != nil {
		minLevel = strings.TrimSpace(*body.MinLevel)
	}
	typesStr := ch.EventTypes
	if body.EventTypes != nil {
		raw, _ := json.Marshal(body.EventTypes)
		typesStr = string(raw)
	}
	projectID := ch.ProjectID
	if body.ProjectID != nil {
		projectID = nullOptionalString(*body.ProjectID)
	}
	now := time.Now().Unix()
	err = d.q.UpdateNotificationChannel(ctx, sqlc.UpdateNotificationChannelParams{
		Name:                name,
		Enabled:             enabled,
		Config:              cfgStr,
		SecretEnc:           secretEnc,
		MinLevel:            minLevel,
		EventTypes:          typesStr,
		ProjectID:           projectID,
		Status:              ch.Status,
		DisabledReason:      ch.DisabledReason,
		ConsecutiveFailures: ch.ConsecutiveFailures,
		LastSuccessAt:       ch.LastSuccessAt,
		LastFailureAt:       ch.LastFailureAt,
		LastError:           ch.LastError,
		UpdatedAt:           now,
		ID:                  id,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	d.audit.Record(ctx, audit.Entry{Action: "notification_channel.update", UserID: user.ID, Meta: map[string]any{"id": id}})
	ch, _ = d.q.GetNotificationChannel(ctx, id)
	hint, _ := notify.SecretHint(d.cfg, ch)
	writeJSON(w, http.StatusOK, channelToResponse(ch, hint))
}

func (d apiDeps) handleDeleteNotificationChannel(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if err := d.q.DeleteNotificationChannel(r.Context(), id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	d.audit.Record(r.Context(), audit.Entry{Action: "notification_channel.delete", UserID: user.ID, Meta: map[string]any{"id": id}})
	w.WriteHeader(http.StatusNoContent)
}

var testRateMu sync.Mutex
var testRate = map[string]time.Time{}

func (d apiDeps) handleTestNotificationChannel(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	testRateMu.Lock()
	if t, ok := testRate[user.ID]; ok && time.Since(t) < 12*time.Second {
		testRateMu.Unlock()
		writeJSONError(w, http.StatusTooManyRequests, "rate limited")
		return
	}
	testRate[user.ID] = time.Now()
	testRateMu.Unlock()

	ctx := r.Context()
	ch, err := d.q.GetNotificationChannel(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	start := time.Now()
	res := notify.SendTest(ctx, d.cfg, d.q, d.notifyRegistry(), ch, BuildVersion())
	out := map[string]any{
		"ok":          res.OK,
		"http_status": res.HTTPStatus,
		"latency_ms":  time.Since(start).Milliseconds(),
	}
	if res.Err != "" {
		out["error"] = res.Err
	}
	d.audit.Record(ctx, audit.Entry{Action: "notification_channel.test", UserID: user.ID, Meta: map[string]any{"id": id, "ok": res.OK}})
	writeJSON(w, http.StatusOK, out)
}

func (d apiDeps) handleListNotificationDeliveries(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	limit := int64(50)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	rows, err := d.q.ListChannelDeliveries(r.Context(), sqlc.ListChannelDeliveriesParams{
		ChannelID: id,
		Limit:     limit,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	type rowOut struct {
		ID         string `json:"id"`
		EventID    string `json:"event_id"`
		Status     string `json:"status"`
		Attempts   int64  `json:"attempts"`
		HTTPStatus *int64 `json:"http_status,omitempty"`
		LastError  string `json:"last_error,omitempty"`
		CreatedAt  int64  `json:"created_at"`
		SentAt     *int64 `json:"sent_at,omitempty"`
	}
	out := make([]rowOut, 0, len(rows))
	for _, row := range rows {
		item := rowOut{
			ID:        row.ID,
			EventID:   row.EventID,
			Status:    row.Status,
			Attempts:  row.Attempts,
			CreatedAt: row.CreatedAt,
		}
		if row.HttpStatus.Valid {
			v := row.HttpStatus.Int64
			item.HTTPStatus = &v
		}
		if row.LastError.Valid {
			item.LastError = row.LastError.String
		}
		if row.SentAt.Valid {
			v := row.SentAt.Int64
			item.SentAt = &v
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"deliveries": out})
}

func (d apiDeps) handleRetryNotificationDelivery(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	now := time.Now().Unix()
	if err := d.q.RetryNotificationDelivery(r.Context(), sqlc.RetryNotificationDeliveryParams{
		NextAttemptAt: now,
		ID:            id,
	}); err != nil {
		writeJSONError(w, http.StatusBadRequest, "cannot retry")
		return
	}
	if d.notifier != nil {
		d.notifier.PingWake()
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func nullOptionalString(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func (d apiDeps) notifyRegistry() notify.Registry {
	if d.notifyReg != nil {
		return d.notifyReg
	}
	return notify.BuildRegistry(d.cfg, d.q, BuildVersion())
}
