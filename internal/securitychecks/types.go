package securitychecks

// Status is a check outcome for the dashboard.
type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

// FixKind describes how the UI can remediate a check.
type FixKind string

const (
	FixNone       FixKind = "none"
	FixHostAction FixKind = "host_action"
	FixLink       FixKind = "link"
)

// Result is one baseline check line.
type Result struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Status     Status  `json:"status"`
	Severity   string  `json:"severity"`
	Detail     string  `json:"detail"`
	Fix        FixKind `json:"fix"`
	FixKind    string  `json:"fix_kind,omitempty"`
	FixHref    string  `json:"fix_href,omitempty"`
	LastRunAt  int64   `json:"last_run_at"`
	Extra      any     `json:"extra,omitempty"`
}

// Report is returned from GET/POST checks APIs.
type Report struct {
	RanAt   int64    `json:"ran_at"`
	Checks  []Result `json:"checks"`
	Stale   bool     `json:"stale,omitempty"`
	Cached  bool     `json:"cached,omitempty"`
}
