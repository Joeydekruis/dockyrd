package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
	"github.com/Joeydekruis/dockyrd/internal/securitypolicy"
)

func registerServiceSecurityRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/security", deps.handleGetServiceSecurity)
	mux.HandleFunc("PUT /api/v1/services/{id}/security", deps.handlePutServiceSecurity)
	mux.HandleFunc("POST /api/v1/services/{id}/security/basic-auth", deps.handleSetServiceBasicAuth)
	mux.HandleFunc("DELETE /api/v1/services/{id}/security/basic-auth", deps.handleClearServiceBasicAuth)
	mux.HandleFunc("GET /api/v1/services/{id}/security/events", deps.handleListServiceSecurityEvents)
}

type serviceSecurityStored struct {
	Profile       string   `json:"profile"`
	LockExempt    bool     `json:"lock_exempt"`
	CSP           string   `json:"csp,omitempty"`
	CSPReportOnly bool     `json:"csp_report_only"`
	BasicUser     string   `json:"basic_user,omitempty"`
	BasicAuthSet  bool     `json:"basic_auth_configured"`
	IPAllowlist   []string `json:"ip_allowlist"`
	Maintenance   bool     `json:"maintenance"`
	RLMode        string   `json:"rl_mode"`
	RlRpm         *int64   `json:"rl_rpm,omitempty"`
	RlLoginRpm    *int64   `json:"rl_login_rpm,omitempty"`
	WAFMode       string   `json:"waf_mode"`
	MemMB         *int     `json:"mem_mb,omitempty"`
	CPU           *float64 `json:"cpu,omitempty"`
	PIDs          *int     `json:"pids,omitempty"`
}

type serviceSecurityEffective struct {
	Profile       string            `json:"profile"`
	Headers       map[string]string `json:"headers"`
	LockFloor     bool              `json:"lock_floor"`
	InstanceProfile string          `json:"instance_default_profile"`
	MemMB         int               `json:"mem_mb"`
	CPU           float64           `json:"cpu"`
	PIDs          int               `json:"pids"`
	ProxyRateLimit bool             `json:"proxy_rate_limit_installed"`
	ProxyWAF       bool             `json:"proxy_waf_installed"`
}

type serviceSecurityResponse struct {
	Stored    serviceSecurityStored    `json:"stored"`
	Effective serviceSecurityEffective `json:"effective"`
}

type putServiceSecurityRequest struct {
	Profile       string   `json:"profile"`
	LockExempt    *bool    `json:"lock_exempt"`
	CSP           *string  `json:"csp"`
	CSPReportOnly *bool    `json:"csp_report_only"`
	IPAllowlist   []string `json:"ip_allowlist"`
	Maintenance   *bool    `json:"maintenance"`
	RLMode        *string  `json:"rl_mode"`
	RlRpm         *int64   `json:"rl_rpm"`
	RlLoginRpm    *int64   `json:"rl_login_rpm"`
	WAFMode       *string  `json:"waf_mode"`
	MemMB         *int     `json:"mem_mb"`
	CPU           *float64 `json:"cpu"`
	PIDs          *int     `json:"pids"`
}

func (d apiDeps) handleGetServiceSecurity(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")
	ctx := r.Context()
	projectID, err := d.q.GetServiceProjectID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if _, ok := d.requireProjectScope(w, r, apitoken.ScopeRead, projectID); !ok {
		return
	}
	row, lazy, err := d.ensureServiceSecurity(ctx, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	_ = lazy
	writeJSON(w, http.StatusOK, d.serviceSecurityPayload(ctx, row))
}

func (d apiDeps) handlePutServiceSecurity(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	serviceID := r.PathValue("id")
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
	prev, _, err := d.ensureServiceSecurity(ctx, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	var body putServiceSecurityRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	inst, err := securitypolicy.LoadInstance(ctx, d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "policy missing")
		return
	}
	beforeSite, _ := securitypolicy.SiteFromRow(prev)
	next := beforeSite
	if p := strings.TrimSpace(body.Profile); p != "" {
		next.Profile = securitypolicy.Profile(p)
	}
	if body.LockExempt != nil {
		if *body.LockExempt && !beforeSite.LockExempt {
			_ = d.audit.Record(ctx, audit.Entry{
				UserID: user.ID, Action: "security.lock_exempt", TargetKind: "service", TargetID: serviceID,
			})
		}
		if !*body.LockExempt && user.Role != "owner" && user.Role != "admin" {
			writeJSONError(w, http.StatusForbidden, "admin required")
			return
		}
		ex := *body.LockExempt
		next.LockExempt = ex
	}
	if body.CSP != nil {
		if err := securitypolicy.ValidateCSP(*body.CSP); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		next.CSP = strings.TrimSpace(*body.CSP)
	}
	if body.CSPReportOnly != nil {
		next.CSPReportOnly = *body.CSPReportOnly
	}
	if body.IPAllowlist != nil {
		list, err := securitypolicy.ValidateAllowlist(body.IPAllowlist)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		next.Allowlist = list
	}
	if body.Maintenance != nil {
		next.Maintenance = *body.Maintenance
	}
	if err := securitypolicy.ValidateResources(body.MemMB, body.CPU, body.PIDs); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if body.MemMB != nil {
		next.MemMB = body.MemMB
	}
	if body.CPU != nil {
		next.CPU = body.CPU
	}
	if body.PIDs != nil {
		next.PIDs = body.PIDs
	}
	lockExempt := next.LockExempt || beforeSite.LockExempt
	if fe := securitypolicy.CheckFloor(inst, beforeSite, securitypolicy.FloorInput{Profile: next.Profile}, lockExempt); fe != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": fe.Message, "field": fe.Field})
		return
	}
	allowJSON, _ := json.Marshal(next.Allowlist)
	params := sqlc.UpdateServiceSecurityParams{
		Profile:         string(next.Profile),
		LockExempt:      boolToInt64(next.LockExempt),
		Csp:             nullString(strings.TrimSpace(next.CSP)),
		CspReportOnly:   boolToInt64(next.CSPReportOnly),
		BasicUser:       prev.BasicUser,
		BasicHash:       prev.BasicHash,
		IpAllowlistJson: string(allowJSON),
		Maintenance:     boolToInt64(next.Maintenance),
		RlMode:          prev.RlMode,
		RlRpm:           prev.RlRpm,
		RlLoginRpm:      prev.RlLoginRpm,
		WafMode:         prev.WafMode,
		MemMb:           securitypolicy.NullInt(next.MemMB),
		Cpu:             securitypolicy.NullFloat(next.CPU),
		Pids:            securitypolicy.NullInt(next.PIDs),
		ServiceID:       serviceID,
	}
	if body.RLMode != nil {
		params.RlMode = strings.TrimSpace(*body.RLMode)
	}
	if body.RlRpm != nil {
		params.RlRpm = ptrInt64ToNull(body.RlRpm)
	}
	if body.RlLoginRpm != nil {
		params.RlLoginRpm = ptrInt64ToNull(body.RlLoginRpm)
	}
	if body.WAFMode != nil {
		params.WafMode = strings.TrimSpace(*body.WAFMode)
	}
	if err := d.q.UpdateServiceSecurity(ctx, params); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID: user.ID, Action: "service.security.update", TargetKind: "service", TargetID: serviceID,
	})
	d.syncEdgeProxy(ctx)
	redeploy := svc.Kind == "app" && securitypolicy.ResourcesChanged(beforeSite, next)
	if redeploy && svc.LiveDeploymentID.Valid {
		if err := d.syncAppMemFromSecurity(ctx, serviceID); err == nil {
			_, _ = jobs.EnqueueServiceDeploy(ctx, d.q, serviceID, "redeploy", user.ID, jobs.GitMeta{})
		}
	}
	updated, _, _ := d.ensureServiceSecurity(ctx, serviceID)
	writeJSON(w, http.StatusOK, d.serviceSecurityPayload(ctx, updated))
}

func (d apiDeps) handleSetServiceBasicAuth(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	row, _, err := d.ensureServiceSecurity(ctx, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := securitypolicy.ValidateBasicUsername(body.Username); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	password, hash, err := securitypolicy.GenerateBasicAuthPassword(body.Username)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "hash error")
		return
	}
	params := sqlc.UpdateServiceSecurityParams{
		Profile:         row.Profile,
		LockExempt:      row.LockExempt,
		Csp:             row.Csp,
		CspReportOnly:   row.CspReportOnly,
		BasicUser:       nullString(body.Username),
		BasicHash:       nullString(hash),
		IpAllowlistJson: row.IpAllowlistJson,
		Maintenance:     row.Maintenance,
		RlMode:          row.RlMode,
		RlRpm:           row.RlRpm,
		RlLoginRpm:      row.RlLoginRpm,
		WafMode:         row.WafMode,
		MemMb:           row.MemMb,
		Cpu:             row.Cpu,
		Pids:            row.Pids,
		ServiceID:       serviceID,
	}
	if err := d.q.UpdateServiceSecurity(ctx, params); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID: user.ID, Action: "service.security.basic_auth.set", TargetKind: "service", TargetID: serviceID,
	})
	d.syncEdgeProxy(ctx)
	writeJSON(w, http.StatusOK, map[string]any{"password": password})
}

func (d apiDeps) handleClearServiceBasicAuth(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	if err := d.q.ClearServiceBasicAuth(ctx, serviceID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID: user.ID, Action: "service.security.basic_auth.clear", TargetKind: "service", TargetID: serviceID,
	})
	d.syncEdgeProxy(ctx)
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleListServiceSecurityEvents(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")
	ctx := r.Context()
	projectID, err := d.q.GetServiceProjectID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if _, ok := d.requireProjectScope(w, r, apitoken.ScopeRead, projectID); !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": []any{}})
}

func (d apiDeps) ensureServiceSecurity(ctx context.Context, serviceID string) (sqlc.ServiceSecurity, bool, error) {
	row, err := d.q.GetServiceSecurity(ctx, serviceID)
	if err == nil {
		return row, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return sqlc.ServiceSecurity{}, false, err
	}
	def := securitypolicy.DefaultSite(serviceID)
	allowJSON, _ := json.Marshal(def.Allowlist)
	params := sqlc.InsertServiceSecurityParams{
		ServiceID:       serviceID,
		Profile:         string(def.Profile),
		LockExempt:      0,
		Csp:             sql.NullString{},
		CspReportOnly:   0,
		BasicUser:       sql.NullString{},
		BasicHash:       sql.NullString{},
		IpAllowlistJson: string(allowJSON),
		Maintenance:     0,
		RlMode:          "inherit",
		RlRpm:           sql.NullInt64{},
		RlLoginRpm:      sql.NullInt64{},
		WafMode:         "inherit",
		MemMb:           sql.NullInt64{},
		Cpu:             sql.NullFloat64{},
		Pids:            sql.NullInt64{},
	}
	if err := d.q.InsertServiceSecurity(ctx, params); err != nil {
		return sqlc.ServiceSecurity{}, false, err
	}
	row, err = d.q.GetServiceSecurity(ctx, serviceID)
	return row, true, err
}

func (d apiDeps) serviceSecurityPayload(ctx context.Context, row sqlc.ServiceSecurity) serviceSecurityResponse {
	inst, _ := securitypolicy.LoadInstance(ctx, d.q)
	site, _ := securitypolicy.SiteFromRow(row)
	var allow []string
	_ = json.Unmarshal([]byte(row.IpAllowlistJson), &allow)
	rl, waf := d.proxyBuildFlags(ctx)
	stored := serviceSecurityStored{
		Profile:       row.Profile,
		LockExempt:    row.LockExempt != 0,
		CSPReportOnly: row.CspReportOnly != 0,
		BasicAuthSet:  row.BasicUser.Valid && row.BasicHash.Valid,
		IPAllowlist:   allow,
		Maintenance:   row.Maintenance != 0,
		RLMode:        row.RlMode,
		WAFMode:       row.WafMode,
	}
	if row.Csp.Valid {
		stored.CSP = row.Csp.String
	}
	if row.BasicUser.Valid {
		stored.BasicUser = row.BasicUser.String
	}
	if row.MemMb.Valid {
		v := int(row.MemMb.Int64)
		stored.MemMB = &v
	}
	if row.Cpu.Valid {
		v := row.Cpu.Float64
		stored.CPU = &v
	}
	if row.Pids.Valid {
		v := int(row.Pids.Int64)
		stored.PIDs = &v
	}
	if row.RlRpm.Valid {
		v := row.RlRpm.Int64
		stored.RlRpm = &v
	}
	if row.RlLoginRpm.Valid {
		v := row.RlLoginRpm.Int64
		stored.RlLoginRpm = &v
	}
	return serviceSecurityResponse{
		Stored: stored,
		Effective: serviceSecurityEffective{
			Profile:         string(securitypolicy.EffectiveProfile(inst, site)),
			Headers:         securitypolicy.ResolvedHeaders(inst, site),
			LockFloor:       inst.Lock,
			InstanceProfile: string(inst.DefaultProfile),
			MemMB:           securitypolicy.EffectiveMemMB(inst, site),
			CPU:             securitypolicy.EffectiveCPU(inst, site),
			PIDs:            securitypolicy.EffectivePIDs(inst, site),
			ProxyRateLimit:  rl,
			ProxyWAF:        waf,
		},
	}
}

func (d apiDeps) syncAppMemFromSecurity(ctx context.Context, serviceID string) error {
	svc, err := d.q.GetServiceByID(ctx, serviceID)
	if err != nil || svc.Kind != "app" {
		return err
	}
	inst, err := securitypolicy.LoadInstance(ctx, d.q)
	if err != nil {
		return err
	}
	site, err := securitypolicy.LoadSite(ctx, d.q, serviceID)
	if err != nil {
		return err
	}
	mem := securitypolicy.EffectiveMemMB(inst, site)
	cfg, err := appconfig.Parse(appconfig.RawConfig(svc.Config))
	if err != nil {
		return err
	}
	cfg.MemLimitMB = mem
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return d.q.UpdateServiceConfig(ctx, sqlc.UpdateServiceConfigParams{
		Config:    raw,
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:        serviceID,
	})
}

func nullString(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
