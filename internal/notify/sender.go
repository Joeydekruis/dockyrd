package notify

import (
	"context"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// DeliveryMeta is passed to Send for logging and headers.
type DeliveryMeta struct {
	DeliveryID string
	Target     string
	EventType  string
	Instance   string
	Version    string
}

// Result describes an outbound attempt.
type Result struct {
	OK             bool
	HTTPStatus     int
	RetryAfter     time.Duration
	Permanent      bool
	DisableChannel bool
	Err            string
}

// Rendered holds channel-specific payload bytes or mail fields.
type Rendered struct {
	Body        []byte
	ContentType string
	Headers     map[string]string
}

// EventView is the read model for renderers.
type EventView struct {
	ID           string
	Type         string
	Level        Level
	State        State
	Title        string
	Detail       string
	Href         string
	ProjectID    string
	ServiceID    string
	DeploymentID string
	Data         map[string]any
	OccurredAt   time.Time
}

// ChannelConfig is non-secret JSON from notification_channels.config.
type ChannelConfig map[string]any

// ChannelSecret is decrypted secret_enc JSON.
type ChannelSecret map[string]any

// Sender delivers one notification channel kind.
type Sender interface {
	Kind() string
	Validate(cfg ChannelConfig, secret ChannelSecret, allowPrivate bool) error
	Render(ev EventView, cfg ChannelConfig, publicBase string) (Rendered, error)
	Send(ctx context.Context, cfg ChannelConfig, secret ChannelSecret, r Rendered, meta DeliveryMeta, allowPrivate bool) Result
}

// Registry maps channel kind to Sender.
type Registry map[string]Sender

func DefaultRegistry(version string) Registry {
	return BuildRegistry(config.Config{}, nil, version)
}

// BuildRegistry wires HTTP senders and optional email (requires DB for SMTP + recipients).
func BuildRegistry(cfg config.Config, q *sqlc.Queries, version string) Registry {
	httpFactory := func(allowPrivate bool) *HTTPDeps {
		return &HTTPDeps{Version: version, AllowPrivate: allowPrivate}
	}
	reg := Registry{
		"slack":   &SlackSender{HTTP: httpFactory},
		"discord": &DiscordSender{HTTP: httpFactory},
		"webhook": &WebhookSender{HTTP: httpFactory},
	}
	if q != nil {
		reg["email"] = &EmailSender{Cfg: cfg, Q: q}
	}
	return reg
}

// HTTPDeps holds shared outbound HTTP settings for senders.
type HTTPDeps struct {
	Version      string
	AllowPrivate bool
}
