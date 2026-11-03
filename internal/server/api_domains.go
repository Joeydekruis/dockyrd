package server

import (
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
	if svc.Kind != "app" {
		writeJSONError(w, http.StatusBadRequest, "domains only supported for app services")
		return
	}

	hostname := strings.TrimSpace(body.Hostname)
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

	now := time.Now().Unix()
	domainID := ulid.Make().String()
	isPrimary := int64(0)
	if body.Primary {
		isPrimary = 1
	}
	platformFlag := int64(0)
	if isPlatform {
		platformFlag = 1
	}
	if err := d.q.CreateDomain(ctx, sqlc.CreateDomainParams{
		ID:            domainID,
		ServiceID:     serviceID,
		Hostname:      hostname,
		IsPlatform:    platformFlag,
		IsPrimary:     isPrimary,
		DnsStatus:     sql.NullString{String: "pending", Valid: true},
		DnsCheckedAt:  sql.NullInt64{Int64: now, Valid: true},
		CertStatus:    sql.NullString{String: "pending", Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create domain error")
		return
	}
	row, err := d.q.GetDomainByID(ctx, domainID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, domainFromRow(row))
}

func (d apiDeps) handleDeleteDomain(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	if err := d.q.DeleteDomain(r.Context(), r.PathValue("id")); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "delete error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
