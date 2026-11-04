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
)

type domainResponse struct {
	ID           string  `json:"id"`
	ServiceID    string  `json:"service_id"`
	Hostname     string  `json:"hostname"`
	IsPlatform   bool    `json:"is_platform"`
	IsPrimary    bool    `json:"is_primary"`
	RedirectTo   *string `json:"redirect_to,omitempty"`
	DNSStatus    *string `json:"dns_status,omitempty"`
	CertStatus   *string `json:"cert_status,omitempty"`
}

type createDomainRequest struct {
	Hostname string `json:"hostname"`
	Platform bool   `json:"platform"`
	Primary  bool   `json:"primary"`
}

func registerDomainRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/domains", deps.handleListDomains)
	mux.HandleFunc("POST /api/v1/services/{id}/domains", deps.handleCreateDomain)
	mux.HandleFunc("DELETE /api/v1/domains/{id}", deps.handleDeleteDomain)
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
	if svc.Kind != "app" && svc.Kind != "static_site" {
		writeJSONError(w, http.StatusBadRequest, "domains only supported for app and static site services")
		return
	}

	hostname := normalizeHostname(body.Hostname)
	isPlatform := body.Platform || hostname == ""
	if isPlatform {
		publicIP := d.cfg.PublicIPv4
		if publicIP == "" {
			publicIP = "127.0.0.1"
		}
		hostname, err = proxy.PlatformHostname(svc.Slug, publicIP)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "platform hostname error")
			return
		}
	}

	row, err := d.insertDomain(ctx, serviceID, hostname, isPlatform, body.Primary)
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
	if err := d.q.DeleteDomain(ctx, r.PathValue("id")); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "delete error")
		return
	}
	d.syncEdgeProxy(ctx)
	w.WriteHeader(http.StatusNoContent)
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
		ID:           domainID,
		ServiceID:    serviceID,
		Hostname:     hostname,
		IsPlatform:   platformFlag,
		IsPrimary:    isPrimaryFlag,
		DnsStatus:    sql.NullString{String: dnsStatus, Valid: true},
		DnsCheckedAt: sql.NullInt64{Int64: now, Valid: true},
		CertStatus:   sql.NullString{String: certStatus, Valid: true},
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
			_, _ = d.insertDomain(ctx, svc.ID, custom, false, true)
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
	publicIP := d.cfg.PublicIPv4
	if publicIP == "" {
		publicIP = "127.0.0.1"
	}
	hostname, err := proxy.PlatformHostname(serviceSlug, publicIP)
	if err != nil {
		return sqlc.Domain{}, err
	}
	return d.insertDomain(ctx, serviceID, hostname, true, isPrimary)
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
	return domainResponse{
		ID:         row.ID,
		ServiceID:  row.ServiceID,
		Hostname:   row.Hostname,
		IsPlatform: row.IsPlatform != 0,
		IsPrimary:  row.IsPrimary != 0,
		RedirectTo: redirect,
		DNSStatus:  dns,
		CertStatus: cert,
	}
}
