package notify

import (
	"context"
	"encoding/json"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/instance"
)

// AllEventTypes lists selectable event types for channel configuration.
func AllEventTypes() []string {
	return append(DefaultExternalEventTypes(), TypeDeploySucceeded, TypeBackupSucceeded, TypeTestPing)
}

// ParseEventTypesJSON decodes event_types column JSON.
func ParseEventTypesJSON(raw string) []string {
	var out []string
	_ = json.Unmarshal([]byte(raw), &out)
	if len(out) == 0 {
		return DefaultExternalEventTypes()
	}
	return out
}

// SecretHint returns a masked URL hint for API responses.
func SecretHint(cfg config.Config, ch sqlc.NotificationChannel) (string, error) {
	secret, err := DecryptChannelSecret(cfg, ch.SecretEnc)
	if err != nil || secret == nil {
		return "", err
	}
	url, _ := secretString(secret, "url")
	switch ch.Kind {
	case "slack":
		return MaskSlackURL(url), nil
	case "discord":
		return MaskDiscordURL(url), nil
	case "webhook":
		return MaskWebhookURL(url), nil
	case "email":
		return "SMTP recipients", nil
	default:
		return "", nil
	}
}

// SMTPConfigured reports whether instance SMTP is ready (for UI hints).
func SMTPConfigured(ctx context.Context, q *sqlc.Queries) (bool, error) {
	if q == nil {
		return false, nil
	}
	pub, err := instance.LoadSMTPPublic(ctx, q)
	if err != nil {
		return false, err
	}
	return pub.Configured, nil
}
