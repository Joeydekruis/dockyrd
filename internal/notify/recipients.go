package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type emailRecipientsConfig struct {
	Admins  bool     `json:"admins"`
	UserIDs []string `json:"user_ids"`
	Extra   []string `json:"extra"`
}

type emailChannelConfig struct {
	Recipients emailRecipientsConfig `json:"recipients"`
}

const maxExtraEmailRecipients = 20

// ValidateEmailChannelConfig checks recipient JSON on create/update.
func ValidateEmailChannelConfig(cfg ChannelConfig) error {
	if cfg == nil {
		return fmt.Errorf("config required")
	}
	var parsed emailChannelConfig
	raw, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("invalid config")
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("invalid config")
	}
	if len(parsed.Recipients.Extra) > maxExtraEmailRecipients {
		return fmt.Errorf("too many extra addresses")
	}
	if !parsed.Recipients.Admins && len(parsed.Recipients.UserIDs) == 0 && len(parsed.Recipients.Extra) == 0 {
		return fmt.Errorf("at least one recipient required")
	}
	for _, extra := range parsed.Recipients.Extra {
		if _, err := mail.ParseAddress(strings.TrimSpace(extra)); err != nil {
			return fmt.Errorf("invalid extra address")
		}
	}
	return nil
}

// ResolveEmailTargets expands config to unique lowercase addresses, skipping inbox dedup set.
func ResolveEmailTargets(ctx context.Context, q *sqlc.Queries, cfg ChannelConfig, skip map[string]struct{}) ([]string, error) {
	if q == nil {
		return nil, fmt.Errorf("database not ready")
	}
	var parsed emailChannelConfig
	raw, _ := json.Marshal(cfg)
	_ = json.Unmarshal(raw, &parsed)

	seen := map[string]struct{}{}
	var out []string
	add := func(email string) error {
		email = strings.TrimSpace(email)
		if email == "" {
			return nil
		}
		addr, err := mail.ParseAddress(email)
		if err != nil {
			return fmt.Errorf("invalid address %q", email)
		}
		key := strings.ToLower(addr.Address)
		if _, ok := skip[key]; ok {
			return nil
		}
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}
		out = append(out, key)
		return nil
	}

	if parsed.Recipients.Admins {
		users, err := q.ListUsers(ctx)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			if u.DisabledAt.Valid {
				continue
			}
			if u.Role != "owner" && u.Role != "admin" {
				continue
			}
			if err := add(u.Email); err != nil {
				return nil, err
			}
		}
	}
	for _, id := range parsed.Recipients.UserIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		u, err := q.GetUserByID(ctx, id)
		if err != nil {
			continue
		}
		if err := add(u.Email); err != nil {
			return nil, err
		}
	}
	for _, extra := range parsed.Recipients.Extra {
		if err := add(extra); err != nil {
			return nil, err
		}
	}
	return out, nil
}
