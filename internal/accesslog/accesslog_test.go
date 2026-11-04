package accesslog

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseLine(t *testing.T) {
	line := `{"level":"info","ts":1700000000.5,"logger":"http.log.access","msg":"handled request","request":{"method":"GET","host":"Site.Example.COM","uri":"/index.html"},"status":200,"duration":0.012,"size":4096}`
	ent, ok := ParseLine(line)
	if !ok {
		t.Fatal("expected parse ok")
	}
	if ent.Host != "site.example.com" {
		t.Fatalf("host=%q", ent.Host)
	}
	if ent.Status != 200 || ent.Bytes != 4096 {
		t.Fatalf("status/bytes")
	}
	if ent.DurationMs < 11.9 || ent.DurationMs > 12.1 {
		t.Fatalf("duration ms=%v", ent.DurationMs)
	}
}

func TestTailTextAndAggregate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "access.log")
	lines := []string{
		`{"ts":1700000000,"msg":"handled request","request":{"method":"GET","host":"a.example","uri":"/"},"status":200,"duration":0.01,"size":100}`,
		`{"ts":1700000060,"msg":"handled request","request":{"method":"GET","host":"b.example","uri":"/"},"status":404,"duration":0.02,"size":50}`,
		`{"ts":1700000060,"msg":"handled request","request":{"method":"GET","host":"a.example","uri":"/x"},"status":200,"duration":0.03,"size":200}`,
	}
	if err := os.WriteFile(path, []byte(joinLines(lines)), 0o644); err != nil {
		t.Fatal(err)
	}
	hosts := HostSet([]string{"a.example"})
	text, err := TailText(path, hosts, 10)
	if err != nil {
		t.Fatal(err)
	}
	if text == "" || !contains(text, "a.example") {
		t.Fatalf("tail: %q", text)
	}
	if contains(text, "b.example") {
		t.Fatal("should filter host b")
	}
	since := time.Unix(1700000000, 0).UTC()
	sum, err := AggregateSince(path, hosts, since, "1h")
	if err != nil {
		t.Fatal(err)
	}
	if sum.RequestsTotal != 2 {
		t.Fatalf("requests=%d", sum.RequestsTotal)
	}
	if len(sum.Series) != 2 {
		t.Fatalf("series len=%d", len(sum.Series))
	}
}

func joinLines(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n"
		}
		out += l
	}
	return out
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
