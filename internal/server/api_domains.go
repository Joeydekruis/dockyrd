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

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
	"github.com/Joeydekruis/dockyrd/internal/watch"
)

type domainResponse struct {
	ID             string  `json:"id"`
	ServiceID      string  `json:"service_id"`
	Hostname       string  `json:"hostname"`
	IsPlatform     bool    `json:"is_platform"`
	IsPrimary      bool    `json:"is_primary"`
	RedirectTo     *string `json:"redirect_to,omitempty"`
	DNSStatus      *string `json:"dns_status,omitempty"`
	CertStatus     *string `json:"cert_status,omitempty"`
	TargetService  *string `json:"target_service,omitempty"`
	TargetPort     *int    `json:"target_port,omitempty"`
}

type createDomainRequest struct {
	Hostname      string `json:"hostname"`
	Platform      bool   `json:"platform"`
	Primary       bool   `json:"primary"`
	TargetService string `json:"target_service"`
	TargetPort    int    `json:"target_port"`
}

func registerDomainRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/domains", deps.handleListDomains)
	mux.HandleFunc("POST /api/v1/services/{id}/domains", deps.handleCreateDomain)
	mux.HandleFunc("DELETE /api/v1/domains/{id}", deps.handleDeleteDomain)
	mux.HandleFunc("PATCH /api/v1/domains/{id}", deps.handlePatchDomain)
	mux.HandleFunc("POST /api/v1/domains/{id}/check", deps.handleCheckDomain)
}

func (d apiDeps) handleListDomains(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	if _, err := d.q.GetServiceByID(r.Context(), serviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	rows, err := d.q.ListDomainsForService(r.Context(), serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]domainResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, domainFromRow(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"domains": out})
}

func (d apiDeps) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	var body createDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && r.ContentLength > 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	ctx := r.Context()
	svc, err := d.q.GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if svc.Kind != "app" && svc.Kind != "static_site" && svc.Kind != "compose" {
		writeJSONError(w, http.StatusBadRequest, "domains not supported for this service kind")
		return
	}
	if svc.Kind == "compose" {
		target := strings.TrimSpace(body.TargetService)
		if target == "" || body.TargetPort <= 0 {
			writeJSONError(w, http.StatusBadRequest, "target_service and target_port required for compose domains")
			return
		}
	}

	hostname := normalizeHostname(body.Hostname)
	isPlatform := body.Platform || hostname == ""
	if isPlatform {
		publicIP := d.cfg.EffectivePublicIPv4()
		hostname, err = proxy.PlatformHostname(svc.Slug, publicIP)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "platform hostname error")
			return
		}
	}

	row, err := d.insertDomain(ctx, serviceID, hostname, isPlatform, body.Primary, body.TargetService, body.TargetPort)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create domain error")
		return
	}
	d.syncEdgeProxy(ctx)
	writeJSON(w, http.StatusCreated, domainFromRow(row))
}

func (d apiDeps) handleDeleteDomain(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	ctx := r.Context()
	id := r.PathValue("id")
	dom, err := d.q.GetDomainByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "domain not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	wasPrimary := dom.IsPrimary != 0
	serviceID := dom.ServiceID
	if err := d.q.DeleteDomain(ctx, id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "delete error")
		return
	}
	if wasPrimary {
		rows, err := d.q.ListDomainsForService(ctx, serviceID)
		if err == nil && len(rows) > 0 {
			pick := rows[0].ID
			for _, row := range rows {
				if row.IsPlatform != 0 {
					pick = row.ID
					break
				}
			}
			_ = d.q.ClearPrimaryForService(ctx, serviceID)
			_ = d.q.SetDomainPrimary(ctx, pick)
		}
	}
	d.syncEdgeProxy(ctx)
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handlePatchDomain(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	var body struct {
		Primary       *bool  `json:"primary"`
		TargetService string `json:"target_service"`
		TargetPort    *int   `json:"target_port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	hasTarget := strings.TrimSpace(body.TargetService) != "" && body.TargetPort != nil
	if body.Primary == nil && !hasTarget {
		writeJSONError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	if body.Primary != nil && !*body.Primary && !hasTarget {
		writeJSONError(w, http.StatusBadRequest, "only {\"primary\": true} or compose target updates are supported")
		return
	}
	ctx := r.Context()
	id := r.PathValue("id")
	dom, err := d.q.GetDomainByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "domain not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if hasTarget {
		svc, err := d.q.GetServiceByID(ctx, dom.ServiceID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "database error")
			return
		}
		if svc.Kind != "compose" {
			writeJSONError(w, http.StatusBadRequest, "target_service applies to compose services only")
			return
		}
		if *body.TargetPort <= 0 {
			writeJSONError(w, http.StatusBadRequest, "target_port must be positive")
			return
		}
		if err := d.q.UpdateDomainComposeTarget(ctx, sqlc.UpdateDomainComposeTargetParams{
			TargetService: sql.NullString{String: strings.TrimSpace(body.TargetService), Valid: true},
			TargetPort:    sql.NullInt64{Int64: int64(*body.TargetPort), Valid: true},
			ID:            id,
		}); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "update error")
			return
		}
	}
	if body.Primary != nil && *body.Primary {
		if err := d.q.SetDomainPrimaryForService(ctx, sqlc.SetDomainPrimaryForServiceParams{
			ID:        id,
			ServiceID: dom.ServiceID,
		}); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "update error")
			return
		}
	}
	d.syncEdgeProxy(ctx)
	row, err := d.q.GetDomainByID(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, domainFromRow(row))
}

func (d apiDeps) handleCheckDomain(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()
	id := r.PathValue("id")
	if _, err := d.q.GetDomainByID(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "domain not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if _, err := watch.CheckDomainDNS(ctx, d.cfg, d.q, id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "dns check failed")
		return
	}
	row, err := d.q.GetDomainByID(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, domainFromRow(row))
}

func normalizeHostname(raw string) string {
	h := strings.TrimSpace(raw)
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	h = strings.TrimSuffix(h, "/")
	if i := strings.Index(h, "/"); i >= 0 {
		h = h[:i]
	}
	return strings.ToLower(h)
}

func (d apiDeps) insertDomain(
	ctx context.Context,
	serviceID, hostname string,
	isPlatform, isPrimary bool,
	targetService string,
	targetPort int,
) (sqlc.Domain, error) {
	now := time.Now().Unix()
	domainID := ulid.Make().String()
	isPrimaryFlag := int64(0)
	if isPrimary {
		isPrimaryFlag = 1
	}
	platformFlag := int64(0)
	if isPlatform {
		platformFlag = 1
	}
	dnsStatus := "pending"
	certStatus := "pending"
	if isPlatform {
		dnsStatus = "active"
		certStatus = "active"
	}
	if err := d.q.CreateDomain(ctx, sqlc.CreateDomainParams{
		ID:            domainID,
		ServiceID:     serviceID,
		Hostname:      hostname,
		IsPlatform:    platformFlag,
		IsPrimary:     isPrimaryFlag,
		DnsStatus:     sql.NullString{String: dnsStatus, Valid: true},
		DnsCheckedAt:  sql.NullInt64{Int64: now, Valid: true},
		CertStatus:    sql.NullString{String: certStatus, Valid: true},
		TargetService: sql.NullString{String: strings.TrimSpace(targetService), Valid: strings.TrimSpace(targetService) != ""},
		TargetPort:    sql.NullInt64{Int64: int64(targetPort), Valid: targetPort > 0},
	}); err != nil {
		return sqlc.Domain{}, err
	}
	return d.q.GetDomainByID(ctx, domainID)
}

func (d apiDeps) provisionServiceDomains(ctx context.Context, svc sqlc.Service, cfg json.RawMessage) {
	custom := strings.TrimSpace(staticSiteCustomDomain(cfg))
	platformPrimary := custom == ""
	if _, err := d.insertPlatformDomain(ctx, svc.ID, svc.Slug, platformPrimary); err != nil {
		return
	}
	if custom != "" {
		custom = normalizeHostname(custom)
		if custom != "" {
			_, _ = d.insertDomain(ctx, svc.ID, custom, false, true, "", 0)
		}
	}
	d.syncEdgeProxy(ctx)
}

func staticSiteCustomDomain(cfg json.RawMessage) string {
	if len(cfg) == 0 {
		return ""
	}
	var partial struct {
		CustomDomain string `json:"customDomain"`
	}
	_ = json.Unmarshal(cfg, &partial)
	return partial.CustomDomain
}

func (d apiDeps) insertPlatformDomain(ctx context.Context, serviceID, serviceSlug string, isPrimary bool) (sqlc.Domain, error) {
	publicIP := d.cfg.EffectivePublicIPv4()
	hostname, err := proxy.PlatformHostname(serviceSlug, publicIP)
	if err != nil {
		return sqlc.Domain{}, err
	}
	return d.insertDomain(ctx, serviceID, hostname, true, isPrimary, "", 0)
}

func domainFromRow(row sqlc.Domain) domainResponse {
	var redirect, dns, cert *string
	if row.RedirectTo.Valid {
		s := row.RedirectTo.String
		redirect = &s
	}
	if row.DnsStatus.Valid {
		s := row.DnsStatus.String
		dns = &s
	}
	if row.CertStatus.Valid {
		s := row.CertStatus.String
		cert = &s
	}
	var targetSvc *string
	if row.TargetService.Valid && strings.TrimSpace(row.TargetService.String) != "" {
		s := row.TargetService.String
		targetSvc = &s
	}
	var targetPort *int
	if row.TargetPort.Valid && row.TargetPort.Int64 > 0 {
		p := int(row.TargetPort.Int64)
		targetPort = &p
	}
	return domainResponse{
		ID:            row.ID,
		ServiceID:     row.ServiceID,
		Hostname:      row.Hostname,
		IsPlatform:    row.IsPlatform != 0,
		IsPrimary:     row.IsPrimary != 0,
		RedirectTo:    redirect,
		DNSStatus:     dns,
		CertStatus:    cert,
		TargetService: targetSvc,
		TargetPort:    targetPort,
	}
}
