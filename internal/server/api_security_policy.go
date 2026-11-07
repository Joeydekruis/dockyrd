package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/security/cdn"
	"github.com/Joeydekruis/dockyrd/internal/securitypolicy"
)

func registerSecurityPolicyRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/security/policy", deps.handleGetSecurityPolicy)
	mux.HandleFunc("PUT /api/v1/security/policy", deps.handlePutSecurityPolicy)
}

type policyResponse struct {
	DefaultProfile string                         `json:"default_profile"`
	Lock           bool                           `json:"lock"`
	Headers        map[string]securitypolicy.HeaderEntry `json:"headers"`
	HeaderCatalog  []string                       `json:"header_catalog"`
	RlDefaultRpm   *int64                         `json:"rl_default_rpm,omitempty"`
	RlLoginRpm     *int64                         `json:"rl_login_rpm,omitempty"`
	CDNMode        string                         `json:"cdn_mode"`
	CDNRanges      []string                       `json:"cdn_ranges"`
	CDNHeader      string                         `json:"cdn_header"`
	DefaultMemMB   int                            `json:"default_mem_mb"`
	DefaultCPU     float64                        `json:"default_cpu"`
	DefaultPIDs    int                            `json:"default_pids"`
	ProxyRateLimit bool                           `json:"proxy_rate_limit_installed"`
	ProxyWAF       bool                           `json:"proxy_waf_installed"`
}

type putPolicyRequest struct {
	DefaultProfile string                                     `json:"default_profile"`
	Lock           bool                                       `json:"lock"`
	Headers        map[string]securitypolicy.HeaderEntry      `json:"headers"`
	RlDefaultRpm   *int64                                     `json:"rl_default_rpm"`
	RlLoginRpm     *int64                                     `json:"rl_login_rpm"`
	CDNMode        string                                     `json:"cdn_mode"`
	CDNRanges      []string                                   `json:"cdn_ranges"`
	CDNHeader      string                                     `json:"cdn_header"`
	DefaultMemMB   int                                        `json:"default_mem_mb"`
	DefaultCPU     float64                                    `json:"default_cpu"`
	DefaultPIDs    int                                        `json:"default_pids"`
}

func (d apiDeps) handleGetSecurityPolicy(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	row, err := d.q.GetSecurityPolicy(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	headers, err := securitypolicy.ParseHeadersJSON(row.HeadersJson)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "invalid headers_json")
		return
	}
	var ranges []string
	_ = json.Unmarshal([]byte(row.CdnRangesJson), &ranges)
	rlInstalled, wafInstalled := d.proxyBuildFlags(r.Context())
	writeJSON(w, http.StatusOK, policyResponse{
		DefaultProfile: row.DefaultProfile,
		Lock:           row.Lock != 0,
		Headers:        headers,
		HeaderCatalog:  securitypolicy.HeaderCatalog(),
		RlDefaultRpm:   nullInt64Ptr(row.RlDefaultRpm),
		RlLoginRpm:     nullInt64Ptr(row.RlLoginRpm),
		CDNMode:        row.CdnMode,
		CDNRanges:      ranges,
		CDNHeader:      row.CdnHeader,
		DefaultMemMB:   int(row.DefaultMemMb),
		DefaultCPU:     row.DefaultCpu,
		DefaultPIDs:    int(row.DefaultPids),
		ProxyRateLimit: rlInstalled,
		ProxyWAF:       wafInstalled,
	})
}

func (d apiDeps) handlePutSecurityPolicy(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	var body putPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.DefaultProfile = strings.TrimSpace(body.DefaultProfile)
	if body.DefaultProfile != "baseline" && body.DefaultProfile != "strict" {
		writeJSONError(w, http.StatusBadRequest, "default_profile must be baseline or strict")
		return
	}
	if body.Headers == nil {
		body.Headers = securitypolicy.DefaultBaselineHeaders()
	}
	headersRaw, err := json.Marshal(body.Headers)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid headers")
		return
	}
	inst := securitypolicy.Instance{
		DefaultProfile: securitypolicy.Profile(body.DefaultProfile),
		Lock:           body.Lock,
		Headers:        body.Headers,
	}
	if fe := securitypolicy.ValidatePolicyHeaders(inst); fe != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": fe.Message, "field": fe.Field})
		return
	}
	cdnMode := strings.TrimSpace(body.CDNMode)
	if cdnMode == "" {
		cdnMode = "direct"
	}
	ranges := body.CDNRanges
	if cdnMode == "cloudflare" {
		fetched, err := cdn.FetchCloudflareRanges(r.Context(), nil)
		if err != nil {
			writeJSONError(w, http.StatusBadGateway, "could not fetch Cloudflare IP ranges")
			return
		}
		ranges = fetched
	}
	if cdnMode == "custom" && len(ranges) == 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "cdn_ranges required", "field": "cdn_ranges"})
		return
	}
	rangesJSON, _ := json.Marshal(ranges)
	if err := d.q.UpdateSecurityPolicy(r.Context(), sqlc.UpdateSecurityPolicyParams{
		DefaultProfile: body.DefaultProfile,
		Lock:           boolToInt64(body.Lock),
		HeadersJson:    string(headersRaw),
		RlDefaultRpm:   ptrInt64ToNull(body.RlDefaultRpm),
		RlLoginRpm:     ptrInt64ToNull(body.RlLoginRpm),
		CdnMode:        cdnMode,
		CdnRangesJson:  string(rangesJSON),
		CdnHeader:      strings.TrimSpace(body.CDNHeader),
		DefaultMemMb:   int64(body.DefaultMemMB),
		DefaultCpu:     body.DefaultCPU,
		DefaultPids:    int64(body.DefaultPIDs),
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update error")
		return
	}
	_ = d.audit.Record(r.Context(), audit.Entry{
		UserID: user.ID, Action: "security.policy.update", TargetKind: "security_policy", TargetID: "1",
	})
	d.syncEdgeProxy(r.Context())
	d.handleGetSecurityPolicy(w, r)
}

func (d apiDeps) proxyBuildFlags(ctx context.Context) (rateLimit, waf bool) {
	for _, id := range []string{"proxy-waf", "proxy-rate-limit"} {
		row, err := d.q.GetAddon(ctx, id)
		if err != nil {
			continue
		}
		if row.State != "installed" {
			continue
		}
		if id == "proxy-waf" {
			waf = true
		}
		if id == "proxy-rate-limit" {
			rateLimit = true
		}
	}
	return rateLimit, waf
}

func ptrInt64ToNull(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}
