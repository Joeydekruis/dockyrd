package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type Logger struct {
	Q *sqlc.Queries
}

type Entry struct {
	UserID     string
	Action     string
	TargetKind string
	TargetID   string
	Meta       map[string]any
	IP         string
}

func (l *Logger) Record(ctx context.Context, in Entry) error {
	if l == nil || l.Q == nil || in.Action == "" {
		return nil
	}
	var metaPayload interface{}
	if len(in.Meta) > 0 {
		b, err := json.Marshal(in.Meta)
		if err != nil {
			return err
		}
		metaPayload = string(b)
	}
	return l.Q.InsertAuditLog(ctx, sqlc.InsertAuditLogParams{
		ID:         ulid.Make().String(),
		UserID:     sql.NullString{String: in.UserID, Valid: in.UserID != ""},
		Action:     sql.NullString{String: in.Action, Valid: true},
		TargetKind: sql.NullString{String: in.TargetKind, Valid: in.TargetKind != ""},
		TargetID:   sql.NullString{String: in.TargetID, Valid: in.TargetID != ""},
		Meta:       metaPayload,
		Ip:         sql.NullString{String: in.IP, Valid: in.IP != ""},
		CreatedAt:  sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
	})
}
