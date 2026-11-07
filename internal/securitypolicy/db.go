package securitypolicy

import (
	"database/sql"
	"encoding/json"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// InstanceFromRow maps sqlc security_policy to Instance.
func InstanceFromRow(row sqlc.SecurityPolicy) (Instance, error) {
	headers, err := ParseHeadersJSON(row.HeadersJson)
	if err != nil {
		return Instance{}, err
	}
	return Instance{
		DefaultProfile: Profile(row.DefaultProfile),
		Lock:           row.Lock != 0,
		Headers:        headers,
		HeadersJSON:    []byte(row.HeadersJson),
		CDNMode:        row.CdnMode,
		CDNRanges:      ParseCDNRangeJSON(row.CdnRangesJson),
		CDNHeader:      row.CdnHeader,
		DefaultMemMB:   int(row.DefaultMemMb),
		DefaultCPU:     row.DefaultCpu,
		DefaultPIDs:    int(row.DefaultPids),
	}, nil
}

// SiteFromRow maps sqlc service_security to Site.
func SiteFromRow(row sqlc.ServiceSecurity) (Site, error) {
	var allow []string
	if strings := row.IpAllowlistJson; strings != "" {
		_ = json.Unmarshal([]byte(strings), &allow)
	}
	site := Site{
		ServiceID:     row.ServiceID,
		Profile:       Profile(row.Profile),
		LockExempt:    row.LockExempt != 0,
		CSPReportOnly: row.CspReportOnly != 0,
		Maintenance:   row.Maintenance != 0,
		Allowlist:     allow,
	}
	if row.Csp.Valid {
		site.CSP = row.Csp.String
	}
	if row.BasicUser.Valid {
		site.BasicUser = row.BasicUser.String
	}
	if row.BasicHash.Valid {
		site.BasicHash = row.BasicHash.String
	}
	if row.MemMb.Valid {
		v := int(row.MemMb.Int64)
		site.MemMB = &v
	}
	if row.Cpu.Valid {
		v := row.Cpu.Float64
		site.CPU = &v
	}
	if row.Pids.Valid {
		v := int(row.Pids.Int64)
		site.PIDs = &v
	}
	return site, nil
}

// DefaultSite returns inherit defaults for lazy service_security rows.
func DefaultSite(serviceID string) Site {
	return Site{ServiceID: serviceID, Profile: ProfileInherit}
}

// ParseAllowlistJSON decodes stored JSON array.
func ParseAllowlistJSON(raw string) ([]string, error) {
	raw = stringsTrimJSON(raw)
	if raw == "" || raw == "[]" {
		return nil, nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return ValidateAllowlist(out)
}

func stringsTrimJSON(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n') {
		s = s[1:]
	}
	return s
}

// NullInt from optional int pointer.
func NullInt(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

// NullFloat from optional float pointer.
func NullFloat(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}
