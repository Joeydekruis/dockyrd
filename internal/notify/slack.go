package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/netsafe"
)

type SlackSender struct {
	HTTP func(allowPrivate bool) *HTTPDeps
}

func (s *SlackSender) Kind() string { return "slack" }

func (s *SlackSender) Validate(cfg ChannelConfig, secret ChannelSecret, _ bool) error {
	raw, _ := secretString(secret, "url")
	return validateSlackURL(raw)
}

func validateSlackURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("url required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url")
	}
	if u.Scheme != "https" {
		return fmt.Errorf("url must use https")
	}
	host := strings.ToLower(u.Hostname())
	if host != "hooks.slack.com" && host != "hooks.slack-gov.com" {
		return fmt.Errorf("host must be hooks.slack.com")
	}
	if !strings.HasPrefix(u.Path, "/services/") {
		return fmt.Errorf("invalid slack webhook path")
	}
	return nil
}

func (s *SlackSender) Render(ev EventView, _ ChannelConfig, publicBase string) (Rendered, error) {
	title := truncateRunes(ev.Title, slackHeaderMax)
	section := escapeSlackMrkdwn(truncateRunes(ev.Detail, slackSectionMax))
	if section == "" {
		section = escapeSlackMrkdwn(title)
	}
	link := absoluteURL(publicBase, ev.Href)
	blocks := []any{
		map[string]any{
			"type": "header",
			"text": map[string]string{"type": "plain_text", "text": title},
		},
		map[string]any{
			"type": "section",
			"text": map[string]string{"type": "mrkdwn", "text": section},
		},
	}
	if link != "" {
		blocks = append(blocks, map[string]any{
			"type": "actions",
			"elements": []any{
				map[string]any{
					"type": "button",
					"text": map[string]string{"type": "plain_text", "text": "View in Dockyrd"},
					"url":  link,
				},
			},
		})
	}
	payload := map[string]any{
		"text": title,
		"attachments": []any{
			map[string]any{
				"color":  levelColorHex(ev.Level),
				"blocks": blocks,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Rendered{}, err
	}
	return Rendered{Body: body, ContentType: "application/json"}, nil
}

func (s *SlackSender) Send(ctx context.Context, _ ChannelConfig, secret ChannelSecret, r Rendered, meta DeliveryMeta, allowPrivate bool) Result {
	webhookURL, _ := secretString(secret, "url")
	deps := s.HTTP(allowPrivate)
	client := netsafe.HTTPClient(netsafe.ClientOptions{AllowPrivateTargets: allowPrivate, Timeout: 10 * time.Second})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(r.Body))
	if err != nil {
		return Result{Err: err.Error()}
	}
	req.Header.Set("Content-Type", r.ContentType)
	req.Header.Set("User-Agent", userAgent(deps.Version))
	resp, err := client.Do(req)
	if err != nil {
		return Result{Err: netsafe.RedactError(err, webhookURL)}
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(io.LimitReader(resp.Body, 512))
	return classifyHTTPResult(resp, webhookURL, meta)
}

func userAgent(version string) string {
	if version == "" {
		version = "dev"
	}
	return fmt.Sprintf("Dockyard/%s (+https://dockyrd.io)", version)
}

func secretString(secret ChannelSecret, key string) (string, bool) {
	if secret == nil {
		return "", false
	}
	v, ok := secret[key]
	if !ok {
		return "", false
	}
	s, _ := v.(string)
	return strings.TrimSpace(s), s != ""
}

func classifyHTTPResult(resp *http.Response, secretURL string, _ DeliveryMeta) Result {
	status := resp.StatusCode
	if status >= 200 && status < 300 {
		return Result{OK: true, HTTPStatus: status}
	}
	if status == 429 {
		ra := resp.Header.Get("Retry-After")
		d := parseRetryAfter(ra)
		return Result{HTTPStatus: status, RetryAfter: d, Err: "rate limited"}
	}
	if status == 400 {
		return Result{HTTPStatus: status, Permanent: true, Err: "invalid payload"}
	}
	if status == 401 || status == 403 {
		return Result{HTTPStatus: status, Permanent: true, DisableChannel: true, Err: "authentication rejected"}
	}
	if status == 404 || status == 410 {
		return Result{HTTPStatus: status, Permanent: true, DisableChannel: true, Err: "webhook not found"}
	}
	if status >= 300 && status < 400 {
		return Result{HTTPStatus: status, Permanent: true, Err: "redirect not followed"}
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", status)
	}
	if status >= 500 || status == 408 {
		return Result{HTTPStatus: status, Err: msg}
	}
	return Result{HTTPStatus: status, Permanent: true, Err: msg}
}

func parseRetryAfter(h string) time.Duration {
	h = strings.TrimSpace(h)
	if h == "" {
		return 30 * time.Second
	}
	if sec, err := time.ParseDuration(h + "s"); err == nil {
		if sec > time.Hour {
			return time.Hour
		}
		return sec
	}
	return 30 * time.Second
}

// MaskSlackURL returns a redacted hint for API responses.
func MaskSlackURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return "••••"
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 3 {
		return fmt.Sprintf("https://%s/services/%s/%s/••••", u.Hostname(), parts[1], parts[2])
	}
	return "https://" + u.Hostname() + "/••••"
}
