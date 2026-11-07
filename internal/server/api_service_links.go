package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/links"
	"github.com/Joeydekruis/dockyrd/internal/serviceops"
)

type linkPeerResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
}

type linkResponse struct {
	ID             string           `json:"id"`
	FromServiceID  string           `json:"from_service_id"`
	ToServiceID    string           `json:"to_service_id"`
	Peer           linkPeerResponse `json:"peer"`
	VariablesCount int              `json:"variables_count"`
	CreatedAt      int64            `json:"created_at"`
}

type createLinkRequest struct {
	ToServiceID   string `json:"to_service_id"`
	FromServiceID string `json:"from_service_id"`
	Variables     []struct {
		Key       string `json:"key"`
		Ref       string `json:"ref"`
		AtBuild   bool   `json:"at_build"`
		AtRuntime bool   `json:"at_runtime"`
		IsSecret  bool   `json:"is_secret"`
	} `json:"variables"`
}

func registerServiceLinkRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/links", deps.handleListServiceLinks)
	mux.HandleFunc("GET /api/v1/services/{id}/linkable", deps.handleListLinkable)
	mux.HandleFunc("POST /api/v1/services/{id}/links", deps.handleCreateServiceLink)
	mux.HandleFunc("DELETE /api/v1/links/{id}", deps.handleDeleteServiceLink)
	mux.HandleFunc("POST /api/v1/projects/{id}/links/backfill", deps.handleBackfillProjectLinks)
}

func (d apiDeps) handleListServiceLinks(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetServiceByID(ctx, serviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	outgoing, err := d.buildLinkList(ctx, serviceID, true)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	incoming, err := d.buildLinkList(ctx, serviceID, false)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"outgoing": outgoing,
		"incoming": incoming,
	})
}

func (d apiDeps) buildLinkList(ctx context.Context, serviceID string, outgoing bool) ([]linkResponse, error) {
	var rows []sqlc.ServiceLink
	var err error
	if outgoing {
		rows, err = d.q.ListOutgoingLinks(ctx, serviceID)
	} else {
		rows, err = d.q.ListIncomingLinks(ctx, serviceID)
	}
	if err != nil {
		return nil, err
	}
	out := make([]linkResponse, 0, len(rows))
	for _, row := range rows {
		peerID := row.ToServiceID
		if !outgoing {
			peerID = row.FromServiceID
		}
		peer, err := d.q.GetServiceByID(ctx, peerID)
		if err != nil {
			return nil, err
		}
		vars, err := d.q.ListVariablesByLink(ctx, sql.NullString{String: row.ID, Valid: true})
		if err != nil {
			return nil, err
		}
		out = append(out, linkResponse{
			ID:             row.ID,
			FromServiceID:  row.FromServiceID,
			ToServiceID:    row.ToServiceID,
			Peer:           servicePeer(peer),
			VariablesCount: len(vars),
			CreatedAt:      row.CreatedAt,
		})
	}
	return out, nil
}

func servicePeer(s sqlc.Service) linkPeerResponse {
	status := "stopped"
	if s.DesiredState == "running" {
		status = "running"
	}
	if s.LiveDeploymentID.Valid && s.LiveDeploymentID.String != "" {
		status = "healthy"
	}
	return linkPeerResponse{
		ID:     s.ID,
		Name:   s.Name,
		Kind:   s.Kind,
		Slug:   s.Slug,
		Status: status,
	}
}

func (d apiDeps) handleListLinkable(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	consumer, err := d.q.GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	linked := map[string]struct{}{consumer.ID: {}}
	outLinks, _ := d.q.ListOutgoingLinks(ctx, serviceID)
	for _, l := range outLinks {
		linked[l.ToServiceID] = struct{}{}
	}
	inLinks, _ := d.q.ListIncomingLinks(ctx, serviceID)
	for _, l := range inLinks {
		linked[l.FromServiceID] = struct{}{}
	}

	ctxRow, err := d.q.GetServiceDeployContext(ctx, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	all, err := d.q.ListServicesByProject(ctx, ctxRow.ProjectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	type candidate struct {
		Service             linkPeerResponse `json:"service"`
		SuggestedVariables  []links.SuggestedVariable `json:"suggested_variables"`
		InternalHost        string           `json:"internal_host"`
	}
	candidates := make([]candidate, 0)
	for _, s := range all {
		if s.EnvironmentID != consumer.EnvironmentID {
			continue
		}
		if _, skip := linked[s.ID]; skip {
			continue
		}
		fromKind, toKind := consumer.Kind, s.Kind
		if consumer.Kind != "app" {
			fromKind, toKind = "app", consumer.Kind
			if s.Kind != "app" {
				continue
			}
			if err := links.ValidatePair("app", consumer.Kind); err != nil {
				continue
			}
		} else {
			if err := links.ValidatePair(fromKind, toKind); err != nil {
				continue
			}
		}
		host := internalHostLabel(serviceops.ServiceFromListRow(s))
		candidates = append(candidates, candidate{
			Service:            servicePeer(serviceops.ServiceFromListRow(s)),
			SuggestedVariables: links.SuggestedVariables(s.Kind, s.Slug),
			InternalHost:       host,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"candidates": candidates})
}

func internalHostLabel(s sqlc.Service) string {
	switch s.Kind {
	case "postgres":
		return s.Slug + ":5432"
	case "redis":
		return s.Slug + ":6379"
	case "bucket":
		return "garage"
	default:
		return s.Slug
	}
}

func (d apiDeps) handleCreateServiceLink(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireMember(w, r)
	if !ok {
		return
	}
	pathID := r.PathValue("id")
	ctx := r.Context()
	var body createLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.ToServiceID = strings.TrimSpace(body.ToServiceID)
	body.FromServiceID = strings.TrimSpace(body.FromServiceID)

	fromID := pathID
	toID := body.ToServiceID
	anchor, err := d.q.GetServiceByID(ctx, pathID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if anchor.Kind != "app" {
		if body.FromServiceID == "" || body.ToServiceID != "" {
			writeJSONError(w, http.StatusBadRequest, "from_service_id required when linking from a provider")
			return
		}
		fromID = body.FromServiceID
		toID = pathID
	} else if toID == "" {
		writeJSONError(w, http.StatusBadRequest, "to_service_id required")
		return
	}
	if fromID == toID {
		writeJSONError(w, http.StatusUnprocessableEntity, "cannot link service to itself")
		return
	}
	fromSvc, err := d.q.GetServiceByID(ctx, fromID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid from service")
		return
	}
	toSvc, err := d.q.GetServiceByID(ctx, toID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid to service")
		return
	}
	if fromSvc.EnvironmentID != toSvc.EnvironmentID {
		writeJSONError(w, http.StatusUnprocessableEntity, "services must share an environment")
		return
	}
	if err := links.ValidatePair(fromSvc.Kind, toSvc.Kind); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if _, err := d.q.GetServiceLinkByPair(ctx, sqlc.GetServiceLinkByPairParams{
		FromServiceID: fromID,
		ToServiceID:   toID,
	}); err == nil {
		writeJSONError(w, http.StatusConflict, "link already exists")
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	masterKey, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "secrets not configured")
		return
	}
	existingKeys := map[string]struct{}{}
	varRows, _ := d.q.ListVariablesForService(ctx, fromID)
	for _, v := range varRows {
		existingKeys[v.Key] = struct{}{}
	}

	linkID := ulid.Make().String()
	now := time.Now().Unix()
	if err := d.q.CreateServiceLink(ctx, sqlc.CreateServiceLinkParams{
		ID:            linkID,
		FromServiceID: fromID,
		ToServiceID:   toID,
		CreatedBy:     sql.NullString{String: user.ID, Valid: true},
		CreatedAt:     now,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create link error")
		return
	}
	for _, v := range body.Variables {
		k := strings.TrimSpace(v.Key)
		refVal := strings.TrimSpace(v.Ref)
		if k == "" || refVal == "" {
			continue
		}
		if _, exists := existingKeys[k]; exists {
			writeJSONError(w, http.StatusConflict, "variable key already exists: "+k)
			return
		}
		enc, err := crypto.Encrypt(masterKey, []byte(refVal))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "encrypt error")
			return
		}
		atBuild, atRuntime, isSecret := int64(0), int64(1), int64(0)
		if v.AtBuild {
			atBuild = 1
		}
		if v.AtRuntime {
			atRuntime = 1
		}
		if v.IsSecret {
			isSecret = 1
		}
		if err := d.q.InsertVariableWithLink(ctx, sqlc.InsertVariableWithLinkParams{
			ID:        ulid.Make().String(),
			ServiceID: fromID,
			Key:       k,
			ValueEnc:  enc,
			IsSecret:  isSecret,
			AtBuild:   atBuild,
			AtRuntime: atRuntime,
			LinkID:    sql.NullString{String: linkID, Valid: true},
		}); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "insert variable error")
			return
		}
	}
	_ = d.q.TouchServiceUpdatedAt(ctx, sqlc.TouchServiceUpdatedAtParams{
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:        fromID,
	})
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "link.create",
		TargetKind: "service_link",
		TargetID:   linkID,
		Meta: map[string]any{
			"from_service_id": fromID,
			"to_service_id":   toID,
		},
	})
	writeJSON(w, http.StatusCreated, map[string]any{"id": linkID})
}

func (d apiDeps) handleDeleteServiceLink(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireMember(w, r)
	if !ok {
		return
	}
	linkID := r.PathValue("id")
	removeVars := r.URL.Query().Get("remove_variables") != "false"
	ctx := r.Context()
	link, err := d.q.GetServiceLinkByID(ctx, linkID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "link not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if removeVars {
		if err := d.q.DeleteVariablesByLink(ctx, sql.NullString{String: linkID, Valid: true}); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "delete variables error")
			return
		}
	}
	if err := d.q.DeleteServiceLink(ctx, linkID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "delete link error")
		return
	}
	now := time.Now().Unix()
	_ = d.q.TouchServiceUpdatedAt(ctx, sqlc.TouchServiceUpdatedAtParams{
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:        link.FromServiceID,
	})
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "link.delete",
		TargetKind: "service_link",
		TargetID:   linkID,
		Meta: map[string]any{
			"from_service_id": link.FromServiceID,
			"to_service_id":   link.ToServiceID,
		},
	})
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleBackfillProjectLinks(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	_ = user
	projectID := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetProjectByID(ctx, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "project not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	masterKey, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "secrets not configured")
		return
	}
	created, err := links.BackfillProject(ctx, d.q, masterKey, projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"links_created": created})
}
