package notify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/netsafe"
)

func TestDecideStatefulEscalation(t *testing.T) {
	prev := &AlertStateRow{
		State:          string(StateFiring),
		Level:          string(LevelWarning),
		LastNotifiedAt: time.Now().Add(-1 * time.Hour).Unix(),
		Occurrences:    2,
	}
	ev := Event{Type: TypeDiskHigh, Level: LevelError, State: StateFiring, DedupKey: "disk:/"}
	dec := DecideStateful(prev, ev, time.Now(), 24*time.Hour)
	if !dec.Notify {
		t.Fatal("expected notify on level escalation")
	}
}

func TestDecideStatefulSuppressReminder(t *testing.T) {
	prev := &AlertStateRow{
		State:          string(StateFiring),
		Level:          string(LevelWarning),
		LastNotifiedAt: time.Now().Add(-2 * time.Hour).Unix(),
	}
	ev := Event{Type: TypeDiskHigh, Level: LevelWarning, State: StateFiring, DedupKey: "disk:/"}
	dec := DecideStateful(prev, ev, time.Now(), 24*time.Hour)
	if dec.Notify {
		t.Fatal("expected suppress within reminder window")
	}
}

func TestEscapeSlackBlocksInjection(t *testing.T) {
	in := "<!channel> https://evil|click"
	out := escapeSlackMrkdwn(in)
	if out == in {
		t.Fatalf("expected escaped output, got %q", out)
	}
}

func TestValidateSlackURLRejectsHTTP(t *testing.T) {
	if err := validateSlackURL("http://hooks.slack.com/services/T/B/x"); err == nil {
		t.Fatal("expected error for http")
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"type":"test.ping"}`)
	ts := "1700000000"
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ts + "."))
	_, _ = mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	if !VerifyWebhookSignature(secret, ts, body, "sha256="+sig, time.Unix(1700000000, 0)) {
		t.Fatal("signature should verify")
	}
}

func TestRedactSlackError(t *testing.T) {
	url := "https://hooks.slack.com/services/AAA/BBB/secret-token"
	errText := "Post \"" + url + "\": dial failed"
	redacted := netsafe.RedactError(stringsErr(errText), url)
	if strings.Contains(redacted, "secret-token") {
		t.Fatalf("token leaked: %q", redacted)
	}
}

type stringsErr string

func (e stringsErr) Error() string { return string(e) }
