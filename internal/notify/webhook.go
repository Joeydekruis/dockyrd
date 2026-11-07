package notify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/netsafe"
)

type WebhookSender struct {
	HTTP func(allowPrivate bool) *HTTPDeps
}

func (w *WebhookSender) Kind() string { return "webhook" }

func (w *WebhookSender) Validate(_ ChannelConfig, secret ChannelSecret, allowPrivate bool) error {
	raw, _ := secretString(secret, "url")
	return validateGenericWebhookURL(raw, allowPrivate)
}

func validateGenericWebhookURL(raw string, allowPrivate bool) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("url required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url")
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" && allowPrivate {
		return nil
	}
	return fmt.Errorf("url must use https")
}

func (w *WebhookSender) Render(ev EventView, cfg ChannelConfig, publicBase string) (Rendered, error) {
	includeData := true
	if v, ok := cfg["include_data"].(bool); ok {
		includeData = v
	}
	link := absoluteURL(publicBase, ev.Href)
	payload := map[string]any{
		"schema":      "dockyrd.notification.v1",
		"id":          ev.ID,
		"type":        ev.Type,
		"level":       string(ev.Level),
		"state":       string(ev.State),
		"occurred_at": ev.OccurredAt.UTC().Format(time.RFC3339),
		"title":       ev.Title,
		"message":     ev.Detail,
		"url":         link,
	}
	if includeData && len(ev.Data) > 0 {
		payload["data"] = ev.Data
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Rendered{}, err
	}
	return Rendered{Body: body, ContentType: "application/json"}, nil
}

func (w *WebhookSender) Send(ctx context.Context, cfg ChannelConfig, secret ChannelSecret, r Rendered, meta DeliveryMeta, allowPrivate bool) Result {
	targetURL, _ := secretString(secret, "url")
	signingSecret, _ := secretString(secret, "signing_secret")
	deps := w.HTTP(allowPrivate)
	client := netsafe.HTTPClient(netsafe.ClientOptions{AllowPrivateTargets: allowPrivate, Timeout: 10 * time.Second})
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	_, _ = mac.Write([]byte(ts + "."))
	_, _ = mac.Write(r.Body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(r.Body))
	if err != nil {
		return Result{Err: err.Error()}
	}
	req.Header.Set("Content-Type", r.ContentType)
	req.Header.Set("User-Agent", userAgent(deps.Version))
	req.Header.Set("X-dockyrd-Event", meta.EventType)
	req.Header.Set("X-dockyrd-Delivery", meta.DeliveryID)
	req.Header.Set("X-dockyrd-Timestamp", ts)
	if signingSecret != "" {
		req.Header.Set("X-dockyrd-Signature", "sha256="+sig)
	}
	if headers, ok := secret["headers"].(map[string]any); ok {
		for k, v := range headers {
			if s, ok := v.(string); ok && k != "" {
				req.Header.Set(k, s)
			}
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{Err: netsafe.RedactError(err, targetURL)}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return Result{OK: true, HTTPStatus: resp.StatusCode}
	}
	return classifyHTTPResult(resp, targetURL, meta)
}

// VerifyWebhookSignature validates dockyrd.notification.v1 HMAC (for tests/docs).
func VerifyWebhookSignature(secret string, ts string, body []byte, sigHex string, now time.Time) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ts + "."))
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)
	got, err := hex.DecodeString(strings.TrimPrefix(sigHex, "sha256="))
	if err != nil {
		return false
	}
	if !hmac.Equal(expected, got) {
		return false
	}
	tUnix, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}
	delta := now.Unix() - tUnix
	if delta < 0 {
		delta = -delta
	}
	return delta <= 300
}

func MaskWebhookURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return "••••"
	}
	return u.Scheme + "://" + u.Hostname() + "/••••"
}

// WebhookPayloadForTest builds a test.ping body.
func WebhookPayloadForTest(deliveryID string, now time.Time) ([]byte, error) {
	payload := map[string]any{
		"schema":      "dockyrd.notification.v1",
		"id":          deliveryID,
		"type":        TypeTestPing,
		"level":       string(LevelInfo),
		"title":       "Test notification",
		"message":     "This is a test from Dockyrd.",
		"occurred_at": now.UTC().Format(time.RFC3339),
	}
	return json.Marshal(payload)
}

func readWebhookRetryBody(r io.Reader) string {
	b, _ := io.ReadAll(io.LimitReader(r, 200))
	return strings.TrimSpace(string(b))
}
