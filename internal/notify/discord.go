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

var discordHosts = map[string]struct{}{
	"discord.com":        {},
	"discordapp.com":     {},
	"ptb.discord.com":    {},
	"canary.discord.com": {},
}

type DiscordSender struct {
	HTTP func(allowPrivate bool) *HTTPDeps
}

func (d *DiscordSender) Kind() string { return "discord" }

func (d *DiscordSender) Validate(_ ChannelConfig, secret ChannelSecret, _ bool) error {
	raw, _ := secretString(secret, "url")
	return validateDiscordURL(raw)
}

func validateDiscordURL(raw string) error {
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
	if _, ok := discordHosts[strings.ToLower(u.Hostname())]; !ok {
		return fmt.Errorf("host must be discord.com")
	}
	if !strings.HasPrefix(u.Path, "/api/webhooks/") {
		return fmt.Errorf("invalid discord webhook path")
	}
	return nil
}

func (d *DiscordSender) Render(ev EventView, cfg ChannelConfig, publicBase string) (Rendered, error) {
	username := "Dockyrd"
	if v, ok := cfg["username"].(string); ok && strings.TrimSpace(v) != "" {
		username = strings.TrimSpace(v)
	}
	title := truncateRunes(ev.Title, 256)
	desc := escapeDiscordMarkdown(truncateRunes(ev.Detail, 1800))
	link := absoluteURL(publicBase, ev.Href)
	embed := map[string]any{
		"title":       title,
		"description": desc,
		"color":       levelColorDiscord(ev.Level),
		"timestamp":   ev.OccurredAt.UTC().Format(time.RFC3339),
		"footer":      map[string]string{"text": "Dockyrd"},
	}
	if link != "" {
		embed["url"] = link
	}
	payload := map[string]any{
		"username":         username,
		"allowed_mentions": map[string]any{"parse": []string{}},
		"embeds":           []any{embed},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Rendered{}, err
	}
	return Rendered{Body: body, ContentType: "application/json"}, nil
}

func (d *DiscordSender) Send(ctx context.Context, _ ChannelConfig, secret ChannelSecret, r Rendered, meta DeliveryMeta, allowPrivate bool) Result {
	webhookURL, _ := secretString(secret, "url")
	deps := d.HTTP(allowPrivate)
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
	if resp.StatusCode == 204 || (resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return Result{OK: true, HTTPStatus: resp.StatusCode}
	}
	if resp.StatusCode == 429 {
		var body struct {
			RetryAfter float64 `json:"retry_after"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&body)
		wait := time.Duration(body.RetryAfter * float64(time.Second))
		if wait <= 0 {
			wait = parseRetryAfter(resp.Header.Get("Retry-After"))
		}
		if wait > time.Hour {
			wait = time.Hour
		}
		return Result{HTTPStatus: 429, RetryAfter: wait, Err: "rate limited"}
	}
	return classifyHTTPResult(resp, webhookURL, meta)
}

func MaskDiscordURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return "••••"
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 4 {
		return fmt.Sprintf("https://%s/api/webhooks/%s/••••", u.Hostname(), parts[2])
	}
	return "https://" + u.Hostname() + "/••••"
}
