package securitypolicy

// Profile names match security_policy / service_security CHECK constraints.
type Profile string

const (
	ProfileInherit  Profile = "inherit"
	ProfileBaseline Profile = "baseline"
	ProfileStrict   Profile = "strict"
	ProfileCustom   Profile = "custom"
)

// HeaderEntry is one instance policy header toggle + value (no CSP at instance).
type HeaderEntry struct {
	Enabled bool   `json:"enabled"`
	Value   string `json:"value"`
}

// Instance is the singleton security_policy row (resolved fields).
type Instance struct {
	DefaultProfile Profile
	Lock           bool
	Headers        map[string]HeaderEntry
	HeadersJSON    []byte
	CDNMode        string
	CDNRanges      []string
	CDNHeader      string
	DefaultMemMB   int
	DefaultCPU     float64
	DefaultPIDs    int
}

// Site is per-service security settings used for resolution.
type Site struct {
	ServiceID    string
	Profile      Profile
	LockExempt   bool
	CSP          string
	CSPReportOnly bool
	BasicUser    string
	BasicHash    string
	Allowlist    []string
	Maintenance  bool
	MemMB        *int
	CPU          *float64
	PIDs         *int
}

// FloorInput captures a PUT body fragment for floor validation.
type FloorInput struct {
	Profile     Profile
	WAFMode     string
	RLMode      string
	Maintenance bool
}

// FloorError is returned when lock rejects a weaker value (HTTP 422).
type FloorError struct {
	Field   string
	Message string
}

func (e *FloorError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}
