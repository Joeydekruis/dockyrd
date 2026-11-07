package hostops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// Runner orchestrates Plan → Apply → Verify → Confirm/Revert with DB transactions.
type Runner struct {
	DB       *sql.DB
	Q        *sqlc.Queries
	Registry *Registry
	Audit    audit.Logger
	Timer    LockoutTimer
	Now      func() time.Time
	Lockout  time.Duration
}

func NewRunner(db *sql.DB, q *sqlc.Queries, reg *Registry, auditLog audit.Logger, timer LockoutTimer) *Runner {
	if timer == nil {
		timer = NoopLockoutTimer{}
	}
	return &Runner{
		DB:       db,
		Q:        q,
		Registry: reg,
		Audit:    auditLog,
		Timer:    timer,
		Now:      time.Now,
		Lockout:  DefaultLockoutConfirmSeconds * time.Second,
	}
}

func (r *Runner) nowUnix() int64 { return r.Now().Unix() }

// Plan creates a row in planned state after calling Action.Plan.
func (r *Runner) Plan(ctx context.Context, kind, actorUserID string) (string, error) {
	act, err := r.Registry.New(kind)
	if err != nil {
		return "", err
	}
	plan, err := act.Plan(ctx)
	if err != nil {
		return "", err
	}
	planJSON, err := encodePlan(plan)
	if err != nil {
		return "", err
	}
	id := ulid.Make().String()
	now := r.nowUnix()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.Q.WithTx(tx)
	if err := qtx.InsertHostAction(ctx, sqlc.InsertHostActionParams{
		ID:          id,
		Kind:        kind,
		State:       string(StatePlanned),
		PlanJson:    sql.NullString{String: planJSON, Valid: planJSON != ""},
		ActorUserID: sql.NullString{String: actorUserID, Valid: actorUserID != ""},
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		return "", err
	}
	if err := r.auditTx(ctx, qtx, actorUserID, id, kind, "host_action.planned", nil); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}

// Apply transitions planned → applied.
func (r *Runner) Apply(ctx context.Context, id, actorUserID string) error {
	row, act, err := r.loadAction(ctx, id)
	if err != nil {
		return err
	}
	if row.State != string(StatePlanned) {
		return fmt.Errorf("%w: apply requires planned", ErrInvalidTransition)
	}
	payload, err := act.Apply(ctx)
	if err != nil {
		if len(payload) > 0 {
			_ = act.Revert(ctx, payload)
		}
		return err
	}
	revertJSON, err := encodeRevert(payload)
	if err != nil {
		return err
	}
	now := r.nowUnix()
	return r.transition(ctx, id, row, StateApplied, transitionFields{
		revertJSON: revertJSON,
		appliedAt:  now,
		actorID:    actorUserID,
		kind:       row.Kind,
		meta:       map[string]any{"kind": row.Kind},
	})
}

// Verify transitions applied → verified, then confirmed or awaiting_confirm.
func (r *Runner) Verify(ctx context.Context, id string) error {
	row, act, err := r.loadAction(ctx, id)
	if err != nil {
		return err
	}
	if row.State != string(StateApplied) {
		return fmt.Errorf("%w: verify requires applied", ErrInvalidTransition)
	}
	if err := act.Verify(ctx); err != nil {
		if revErr := r.failAndRevert(ctx, id, row, act, err.Error()); revErr != nil {
			return revErr
		}
		return err
	}
	if err := r.transition(ctx, id, row, StateVerified, transitionFields{
		kind: row.Kind,
		meta: map[string]any{"kind": row.Kind},
	}); err != nil {
		return err
	}
	row, act, err = r.loadAction(ctx, id)
	if err != nil {
		return err
	}
	if act.RequiresLockoutConfirm() {
		deadline := r.Now().Add(r.Lockout)
		if err := r.transition(ctx, id, row, StateAwaitingConfirm, transitionFields{
			confirmDeadline: deadline.Unix(),
			kind:            row.Kind,
			meta:            map[string]any{"kind": row.Kind},
		}); err != nil {
			return err
		}
		return r.Timer.Schedule(ctx, id, deadline)
	}
	row, _, err = r.loadAction(ctx, id)
	if err != nil {
		return err
	}
	return r.transition(ctx, id, row, StateConfirmed, transitionFields{
		confirmedAt: r.nowUnix(),
		kind:        row.Kind,
		meta:        map[string]any{"kind": row.Kind},
	})
}

// Confirm transitions awaiting_confirm → confirmed and cancels the lockout timer.
func (r *Runner) Confirm(ctx context.Context, id, actorUserID string) error {
	row, _, err := r.loadAction(ctx, id)
	if err != nil {
		return err
	}
	if row.State != string(StateAwaitingConfirm) {
		return fmt.Errorf("%w: confirm requires awaiting_confirm", ErrInvalidTransition)
	}
	_ = r.Timer.Cancel(ctx, id)
	return r.transition(ctx, id, row, StateConfirmed, transitionFields{
		confirmedAt: r.nowUnix(),
		actorID:     actorUserID,
		kind:        row.Kind,
		meta:        map[string]any{"kind": row.Kind},
	})
}

// RevertMeta annotates audit entries for revert sources.
type RevertMeta struct {
	Source string
	Actor  string
}

// Revert runs Action.Revert and moves to reverted (idempotent when already reverted).
func (r *Runner) Revert(ctx context.Context, id string, meta RevertMeta) error {
	row, act, err := r.loadAction(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return err
		}
		return err
	}
	if row.State == string(StateReverted) {
		return nil
	}
	if !Revertable(State(row.State)) {
		return fmt.Errorf("%w: revert not allowed from %s", ErrInvalidTransition, row.State)
	}
	payload := decodeRevert(row.RevertJson.String)
	if len(payload) > 0 {
		if err := act.Revert(ctx, payload); err != nil {
			return err
		}
	}
	_ = r.Timer.Cancel(ctx, id)
	auditMeta := map[string]any{"kind": row.Kind}
	if meta.Source != "" {
		auditMeta["source"] = meta.Source
	}
	actor := meta.Actor
	return r.transition(ctx, id, row, StateReverted, transitionFields{
		revertedAt: r.nowUnix(),
		actorID:    actor,
		kind:       row.Kind,
		meta:       auditMeta,
	})
}

// Cancel transitions planned → cancelled.
func (r *Runner) Cancel(ctx context.Context, id, actorUserID string) error {
	row, _, err := r.loadAction(ctx, id)
	if err != nil {
		return err
	}
	if row.State != string(StatePlanned) {
		return fmt.Errorf("%w: cancel requires planned", ErrInvalidTransition)
	}
	return r.transition(ctx, id, row, StateCancelled, transitionFields{
		actorID: actorUserID,
		kind:    row.Kind,
		meta:    map[string]any{"kind": row.Kind},
	})
}

func (r *Runner) failAndRevert(ctx context.Context, id string, row sqlc.HostAction, act Action, msg string) error {
	payload := decodeRevert(row.RevertJson.String)
	if len(payload) > 0 {
		if err := act.Revert(ctx, payload); err != nil {
			return err
		}
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := r.Q.WithTx(tx)
	now := r.nowUnix()
	if err := ValidateTransition(State(row.State), StateFailed); err != nil {
		return err
	}
	rowFailed := row
	rowFailed.State = string(StateFailed)
	rowFailed.LastError = sql.NullString{String: msg, Valid: true}
	rowFailed.UpdatedAt = now
	if err := updateHostActionCAS(ctx, qtx, State(row.State), sqlc.UpdateHostActionParams{
		ID:          id,
		State:       string(StateFailed),
		PlanJson:    row.PlanJson,
		RevertJson:  row.RevertJson,
		ActorUserID: row.ActorUserID,
		AppliedAt:   row.AppliedAt,
		LastError:   sql.NullString{String: msg, Valid: true},
		UpdatedAt:   now,
	}); err != nil {
		return err
	}
	if err := r.auditTx(ctx, qtx, row.ActorUserID.String, id, row.Kind, "host_action.failed", map[string]any{
		"kind":   row.Kind,
		"error":  msg,
		"source": "verify_failure",
	}); err != nil {
		return err
	}
	if err := ValidateTransition(StateFailed, StateReverted); err != nil {
		return err
	}
	revertedAt := r.nowUnix()
	if err := updateHostActionCAS(ctx, qtx, StateFailed, sqlc.UpdateHostActionParams{
		ID:          id,
		State:       string(StateReverted),
		PlanJson:    row.PlanJson,
		RevertJson:  row.RevertJson,
		ActorUserID: row.ActorUserID,
		AppliedAt:   row.AppliedAt,
		LastError:   sql.NullString{String: msg, Valid: true},
		RevertedAt:  sql.NullInt64{Int64: revertedAt, Valid: true},
		UpdatedAt:   revertedAt,
	}); err != nil {
		return err
	}
	if err := r.auditTx(ctx, qtx, "", id, row.Kind, "host_action.reverted", map[string]any{
		"kind":   row.Kind,
		"source": "verify_failure",
	}); err != nil {
		return err
	}
	return tx.Commit()
}

type transitionFields struct {
	planJSON        string
	revertJSON      string
	appliedAt       int64
	confirmDeadline int64
	confirmedAt     int64
	revertedAt      int64
	lastError       string
	actorID         string
	kind            string
	meta            map[string]any
}

func (r *Runner) transition(ctx context.Context, id string, row sqlc.HostAction, to State, f transitionFields) error {
	from := State(row.State)
	if err := ValidateTransition(from, to); err != nil {
		return err
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := r.Q.WithTx(tx)

	plan := row.PlanJson
	if f.planJSON != "" {
		plan = sql.NullString{String: f.planJSON, Valid: true}
	}
	revert := row.RevertJson
	if f.revertJSON != "" {
		revert = sql.NullString{String: f.revertJSON, Valid: true}
	}
	applied := row.AppliedAt
	if f.appliedAt != 0 {
		applied = sql.NullInt64{Int64: f.appliedAt, Valid: true}
	}
	confirm := row.ConfirmDeadline
	if f.confirmDeadline != 0 {
		confirm = sql.NullInt64{Int64: f.confirmDeadline, Valid: true}
	}
	confirmed := row.ConfirmedAt
	if f.confirmedAt != 0 {
		confirmed = sql.NullInt64{Int64: f.confirmedAt, Valid: true}
	}
	reverted := row.RevertedAt
	if f.revertedAt != 0 {
		reverted = sql.NullInt64{Int64: f.revertedAt, Valid: true}
	}
	lastErr := row.LastError
	if f.lastError != "" {
		lastErr = sql.NullString{String: f.lastError, Valid: true}
	}
	actor := row.ActorUserID
	if f.actorID != "" {
		actor = sql.NullString{String: f.actorID, Valid: true}
	}
	now := r.nowUnix()
	if err := updateHostActionCAS(ctx, qtx, from, sqlc.UpdateHostActionParams{
		ID:              id,
		State:           string(to),
		PlanJson:        plan,
		RevertJson:      revert,
		ActorUserID:     actor,
		AppliedAt:       applied,
		ConfirmDeadline: confirm,
		ConfirmedAt:     confirmed,
		RevertedAt:      reverted,
		LastError:       lastErr,
		UpdatedAt:       now,
	}); err != nil {
		return err
	}
	actionName := AuditActionForTransition(from, to)
	if err := r.auditTx(ctx, qtx, f.actorID, id, f.kind, actionName, f.meta); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Runner) loadAction(ctx context.Context, id string) (sqlc.HostAction, Action, error) {
	row, err := r.Q.GetHostAction(ctx, id)
	if err != nil {
		return sqlc.HostAction{}, nil, err
	}
	act, err := r.Registry.New(row.Kind)
	if err != nil {
		return row, nil, err
	}
	return row, act, nil
}

func (r *Runner) auditTx(ctx context.Context, qtx *sqlc.Queries, actor, targetID, kind, action string, meta map[string]any) error {
	if meta == nil {
		meta = map[string]any{}
	}
	if kind != "" {
		meta["kind"] = kind
	}
	return insertAuditInTx(ctx, qtx, audit.Entry{
		UserID:     actor,
		Action:     action,
		TargetKind: "host_action",
		TargetID:   targetID,
		Meta:       meta,
	})
}

func updateHostActionCAS(ctx context.Context, qtx *sqlc.Queries, from State, p sqlc.UpdateHostActionParams) error {
	p.ExpectedState = string(from)
	n, err := qtx.UpdateHostAction(ctx, p)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrInvalidTransition
	}
	return nil
}

func insertAuditInTx(ctx context.Context, qtx *sqlc.Queries, in audit.Entry) error {
	if in.Action == "" {
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
	return qtx.InsertAuditLog(ctx, sqlc.InsertAuditLogParams{
		ID:         ulid.Make().String(),
		UserID:     sql.NullString{String: in.UserID, Valid: in.UserID != ""},
		Action:     sql.NullString{String: in.Action, Valid: true},
		TargetKind: sql.NullString{String: in.TargetKind, Valid: in.TargetKind != ""},
		TargetID:   sql.NullString{String: in.TargetID, Valid: in.TargetID != ""},
		Meta:       metaPayload,
		CreatedAt:  sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
	})
}
