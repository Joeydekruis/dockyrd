package preview

import (
	"encoding/json"
)

// Config holds preview deployment settings on a service.
type Config struct {
	Enabled bool `json:"enabled"`
}

// Enabled reports whether preview deployments are enabled in raw service config JSON.
func Enabled(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var root struct {
		Preview Config `json:"preview"`
	}
	if err := json.Unmarshal(raw, &root); err != nil {
		return false
	}
	return root.Preview.Enabled
}

// MergeConfig returns updated service config with preview settings applied.
func MergeConfig(raw json.RawMessage, cfg Config) (json.RawMessage, error) {
	root := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &root); err != nil {
			return nil, err
		}
	}
	root["preview"] = cfg
	out, err := json.Marshal(root)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EnabledRaw reports whether previews are enabled for a sqlc config value.
func EnabledRaw(raw any) bool {
	if raw == nil {
		return false
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return false
	}
	return Enabled(b)
}
