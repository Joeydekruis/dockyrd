package accesslog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// Entry is one Caddy JSON access log line for a matched host.
type Entry struct {
	TS         time.Time
	Host       string
	Method     string
	URI        string
	Status     int
	DurationMs float64
	Bytes      int64
}

type rawAccess struct {
	TS      float64 `json:"ts"`
	Msg     string  `json:"msg"`
	Status  int     `json:"status"`
	Size    int64   `json:"size"`
	Dur     float64 `json:"duration"`
	Request struct {
		Host   string `json:"host"`
		Method string `json:"method"`
		URI    string `json:"uri"`
	} `json:"request"`
}

// HostSet normalizes hostnames for lookup (lowercase, no trailing dot).
func HostSet(hosts []string) map[string]struct{} {
	out := make(map[string]struct{}, len(hosts))
	for _, h := range hosts {
		n := normalizeHost(h)
		if n == "" {
			continue
		}
		out[n] = struct{}{}
	}
	return out
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimSuffix(host, ".")
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	return host
}

// ParseLine decodes one JSON access log line. Returns false when not a handled request.
func ParseLine(line string) (Entry, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return Entry{}, false
	}
	var raw rawAccess
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return Entry{}, false
	}
	if raw.Msg != "" && raw.Msg != "handled request" {
		return Entry{}, false
	}
	host := normalizeHost(raw.Request.Host)
	if host == "" {
		return Entry{}, false
	}
	ts := time.Unix(0, 0).UTC()
	if raw.TS > 0 {
		sec := int64(raw.TS)
		nsec := int64((raw.TS - float64(sec)) * 1e9)
		ts = time.Unix(sec, nsec).UTC()
	}
	return Entry{
		TS:         ts,
		Host:       host,
		Method:     raw.Request.Method,
		URI:        raw.Request.URI,
		Status:     raw.Status,
		DurationMs: raw.Dur * 1000,
		Bytes:      raw.Size,
	}, true
}

// TailText returns the last matching lines as plain text (newest last).
func TailText(path string, hosts map[string]struct{}, maxLines int) (string, error) {
	if maxLines <= 0 {
		maxLines = 200
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	// Allow long JSON lines.
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		ent, ok := ParseLine(sc.Text())
		if !ok || !hostMatches(hosts, ent.Host) {
			continue
		}
		lines = append(lines, formatEntry(ent))
		if len(lines) > maxLines {
			lines = lines[len(lines)-maxLines:]
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}

func formatEntry(e Entry) string {
	ts := e.TS.Format("2006-01-02 15:04:05")
	return fmt.Sprintf(
		"%s  %d  %s %s  %.0fms  %dB  %s",
		ts,
		e.Status,
		e.Method,
		e.URI,
		e.DurationMs,
		e.Bytes,
		e.Host,
	)
}

func hostMatches(hosts map[string]struct{}, host string) bool {
	if len(hosts) == 0 {
		return true
	}
	_, ok := hosts[normalizeHost(host)]
	return ok
}

// Bucket is one time bucket of aggregated traffic.
type Bucket struct {
	TS         int64   `json:"ts"`
	Requests   int64   `json:"requests"`
	Bytes      int64   `json:"bytes"`
	Status2xx  int64   `json:"status_2xx"`
	Status3xx  int64   `json:"status_3xx"`
	Status4xx  int64   `json:"status_4xx"`
	Status5xx  int64   `json:"status_5xx"`
	LatencyAvg float64 `json:"latency_avg_ms"`
	LatencyP95 float64 `json:"latency_p95_ms"`
}

// Summary is analytics for a time window.
type Summary struct {
	Range         string   `json:"range"`
	RequestsTotal int64    `json:"requests_total"`
	BytesTotal    int64    `json:"bytes_total"`
	Status2xx     int64    `json:"status_2xx"`
	Status3xx     int64    `json:"status_3xx"`
	Status4xx     int64    `json:"status_4xx"`
	Status5xx     int64    `json:"status_5xx"`
	Series        []Bucket `json:"series"`
}

type bucketAcc struct {
	requests  int64
	bytes     int64
	status2xx int64
	status3xx int64
	status4xx int64
	status5xx int64
	latencies []float64
}

// AggregateSince scans the log file and buckets entries per minute.
func AggregateSince(path string, hosts map[string]struct{}, since time.Time, rangeLabel string) (Summary, error) {
	out := Summary{Range: rangeLabel}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return out, err
	}
	defer f.Close()

	byMinute := make(map[int64]*bucketAcc)
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		ent, ok := ParseLine(sc.Text())
		if !ok || !hostMatches(hosts, ent.Host) {
			continue
		}
		if ent.TS.Before(since) {
			continue
		}
		minute := ent.TS.Truncate(time.Minute).Unix()
		acc := byMinute[minute]
		if acc == nil {
			acc = &bucketAcc{}
			byMinute[minute] = acc
		}
		acc.requests++
		acc.bytes += ent.Bytes
		switch {
		case ent.Status >= 500:
			acc.status5xx++
		case ent.Status >= 400:
			acc.status4xx++
		case ent.Status >= 300:
			acc.status3xx++
		default:
			acc.status2xx++
		}
		acc.latencies = append(acc.latencies, ent.DurationMs)
		out.RequestsTotal++
		out.BytesTotal += ent.Bytes
	}
	if err := sc.Err(); err != nil {
		return out, err
	}

	keys := make([]int64, 0, len(byMinute))
	for k := range byMinute {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	out.Series = make([]Bucket, 0, len(keys))
	for _, k := range keys {
		acc := byMinute[k]
		out.Status2xx += acc.status2xx
		out.Status3xx += acc.status3xx
		out.Status4xx += acc.status4xx
		out.Status5xx += acc.status5xx
		out.Series = append(out.Series, Bucket{
			TS:         k,
			Requests:   acc.requests,
			Bytes:      acc.bytes,
			Status2xx:  acc.status2xx,
			Status3xx:  acc.status3xx,
			Status4xx:  acc.status4xx,
			Status5xx:  acc.status5xx,
			LatencyAvg: avg(acc.latencies),
			LatencyP95: percentile(acc.latencies, 0.95),
		})
	}
	return out, nil
}

func avg(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	var sum float64
	for _, x := range v {
		sum += x
	}
	return sum / float64(len(v))
}

func percentile(v []float64, p float64) float64 {
	if len(v) == 0 {
		return 0
	}
	cp := append([]float64(nil), v...)
	sort.Float64s(cp)
	idx := int(float64(len(cp)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}
