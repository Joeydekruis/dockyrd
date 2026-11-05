package notify

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/instance"
)

const EmailLogKindNotification = "notification"

// RecordEmailLog persists an SMTP attempt for the admin email log UI.
func RecordEmailLog(ctx context.Context, q *sqlc.Queries, kind, from, to, subject, status, response string, cfg instance.SMTPConfig) {
	if q == nil {
		return
	}
	detailBytes, _ := json.Marshal(map[string]any{
		"smtp_host": cfg.Host,
		"smtp_port": cfg.Port,
		"use_tls":   cfg.UseTLS,
	})
	_ = q.InsertEmailLog(ctx, sqlc.InsertEmailLogParams{
		ID:           ulid.Make().String(),
		Kind:         kind,
		FromAddress:  sql.NullString{String: from, Valid: from != ""},
		ToAddress:    to,
		Subject:      sql.NullString{String: subject, Valid: subject != ""},
		Status:       status,
		ResponseText: sql.NullString{String: response, Valid: response != ""},
		DetailLog:    sql.NullString{String: string(detailBytes), Valid: len(detailBytes) > 0},
		CreatedAt:    time.Now().Unix(),
	})
}

func smtpFromAddress(cfg instance.SMTPConfig) string {
	if cfg.From != "" {
		return cfg.From
	}
	return cfg.Username
}
