package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func registerInboxRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/inbox/summary", deps.handleInboxSummary)
	mux.HandleFunc("GET /api/v1/inbox/notifications", deps.handleListInboxNotifications)
	mux.HandleFunc("POST /api/v1/inbox/notifications/{id}/read", deps.handleMarkInboxRead)
	mux.HandleFunc("POST /api/v1/inbox/notifications/read-all", deps.handleMarkAllInboxRead)
	mux.HandleFunc("GET /api/v1/inbox/open", deps.handleInboxOpenFromEmail)
}

type inboxNotificationResponse struct {
	ID           string  `json:"id"`
	Category     string  `json:"category"`
	Level        string  `json:"level"`
	Title        string  `json:"title"`
	Body         string  `json:"body"`
	SourceLabel  string  `json:"source_label"`
	CtaLabel     *string `json:"cta_label,omitempty"`
	CtaHref      *string `json:"cta_href,omitempty"`
	ReadAt       *int64  `json:"read_at,omitempty"`
	DeploymentID *string `json:"deployment_id,omitempty"`
	ServiceID    *string `json:"service_id,omitempty"`
	ProjectID    *string `json:"project_id,omitempty"`
	CreatedAt    int64   `json:"created_at"`
}

func inboxFromRow(row sqlc.ListInboxNotificationsForUserRow) inboxNotificationResponse {
	out := inboxNotificationResponse{
		ID:          row.ID,
		Category:    row.Category,
		Level:       row.Level,
		Title:       row.Title,
		Body:        row.Body,
		SourceLabel: row.SourceLabel,
		CreatedAt:   row.CreatedAt,
	}
	if row.CtaLabel.Valid {
		out.CtaLabel = &row.CtaLabel.String
	}
	if row.CtaHref.Valid {
		out.CtaHref = &row.CtaHref.String
	}
	if row.ReadAt.Valid {
		out.ReadAt = &row.ReadAt.Int64
	}
	if row.DeploymentID.Valid {
		out.DeploymentID = &row.DeploymentID.String
	}
	if row.ServiceID.Valid {
		out.ServiceID = &row.ServiceID.String
	}
	if row.ProjectID.Valid {
		out.ProjectID = &row.ProjectID.String
	}
	return out
}

func (d apiDeps) handleInboxSummary(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	unread, err := d.q.CountUnreadInboxForUser(ctx, user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	total, err := d.q.CountInboxForUser(ctx, user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"unread": unread,
		"total":  total,
	})
}

func (d apiDeps) handleListInboxNotifications(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	limit := int64(20)
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if category == "all" {
		category = ""
	}
	unreadOnly := int64(0)
	if strings.TrimSpace(r.URL.Query().Get("unread")) == "1" {
		unreadOnly = 1
	}
	var beforeCreated int64
	var cursorActive int64
	if v := strings.TrimSpace(r.URL.Query().Get("before")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			beforeCreated = n
			cursorActive = 1
		}
	}
	rows, err := d.q.ListInboxNotificationsForUser(r.Context(), sqlc.ListInboxNotificationsForUserParams{
		UserID:    user.ID,
		Column2:   category,
		Category:  category,
		Column4:   unreadOnly,
		Column5:   cursorActive,
		CreatedAt: beforeCreated,
		Limit:     limit,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]inboxNotificationResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, inboxFromRow(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"notifications": out})
}

func (d apiDeps) handleMarkInboxRead(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	now := time.Now().Unix()
	if err := d.q.MarkInboxNotificationRead(r.Context(), sqlc.MarkInboxNotificationReadParams{
		ReadAt: sql.NullInt64{Int64: now, Valid: true},
		ID:     id,
		UserID: user.ID,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleMarkAllInboxRead(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	now := time.Now().Unix()
	if err := d.q.MarkAllInboxNotificationsRead(r.Context(), sqlc.MarkAllInboxNotificationsReadParams{
		ReadAt: sql.NullInt64{Int64: now, Valid: true},
		UserID: user.ID,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleInboxOpenFromEmail(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Redirect(w, r, "/notifications", http.StatusFound)
		return
	}
	ctx := r.Context()
	row, err := d.q.GetInboxNotificationByEmailToken(ctx, auth.HashSessionToken(token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Redirect(w, r, "/notifications", http.StatusFound)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	now := time.Now().Unix()
	_ = d.q.MarkInboxNotificationReadByID(ctx, sqlc.MarkInboxNotificationReadByIDParams{
		ReadAt: sql.NullInt64{Int64: now, Valid: true},
		ID:     row.ID,
	})
	target := "/notifications"
	if row.CtaHref.Valid && strings.TrimSpace(row.CtaHref.String) != "" {
		target = row.CtaHref.String
	}
	http.Redirect(w, r, target, http.StatusFound)
}
