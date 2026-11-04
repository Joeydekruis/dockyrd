package dbkind

import (
	"encoding/json"
	"fmt"
)

// SetPostgresPublicPort updates public_port in postgres service config JSON.
func SetPostgresPublicPort(raw json.RawMessage, port int) (json.RawMessage, error) {
	var cfg PostgresConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	if port < 0 {
		return nil, fmt.Errorf("invalid public_port")
	}
	cfg.PublicPort = port
	return json.Marshal(cfg)
}
