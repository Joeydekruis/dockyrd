package dbkind

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
)

const (
	defaultPostgresVersion = 17
	defaultPostgresUser    = "dockyard"
	defaultPostgresDB      = "app"
	defaultMemLimitMB      = 512
	defaultRedisVersion    = 8
	passwordLength         = 32
)

// PostgresConfig is services.config for kind=postgres.
type PostgresConfig struct {
	Version     int    `json:"version"`
	Database    string `json:"database"`
	User        string `json:"user"`
	PasswordEnc string `json:"password_enc"`
	MemLimitMB  int    `json:"mem_limit_mb"`
	PublicPort  int    `json:"public_port"`
}

// RedisConfig is services.config for kind=redis (Valkey image).
type RedisConfig struct {
	Version     int    `json:"version"`
	PasswordEnc string `json:"password_enc"`
	Persistence bool   `json:"persistence"`
	MaxMemoryMB int    `json:"max_memory_mb"`
}

// NewPostgresConfigJSON creates config with a generated password (encrypted in JSON).
func NewPostgresConfigJSON(masterKey []byte, version int) (json.RawMessage, error) {
	if version <= 0 {
		version = defaultPostgresVersion
	}
	pw, err := randomPassword(passwordLength)
	if err != nil {
		return nil, err
	}
	enc, err := crypto.Encrypt(masterKey, []byte(pw))
	if err != nil {
		return nil, err
	}
	cfg := PostgresConfig{
		Version:     version,
		Database:    defaultPostgresDB,
		User:        defaultPostgresUser,
		PasswordEnc: base64.StdEncoding.EncodeToString(enc),
		MemLimitMB:  defaultMemLimitMB,
		PublicPort:  0,
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// NewRedisConfigJSON creates Valkey/redis config with generated password.
func NewRedisConfigJSON(masterKey []byte, version int) (json.RawMessage, error) {
	if version <= 0 {
		version = defaultRedisVersion
	}
	pw, err := randomPassword(passwordLength)
	if err != nil {
		return nil, err
	}
	enc, err := crypto.Encrypt(masterKey, []byte(pw))
	if err != nil {
		return nil, err
	}
	cfg := RedisConfig{
		Version:     version,
		PasswordEnc: base64.StdEncoding.EncodeToString(enc),
		Persistence: true,
		MaxMemoryMB: 256,
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// ParsePostgres validates config and returns plaintext password for container env.
func ParsePostgres(raw json.RawMessage, masterKey []byte) (PostgresConfig, string, error) {
	var cfg PostgresConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return PostgresConfig{}, "", err
	}
	if cfg.Version <= 0 {
		cfg.Version = defaultPostgresVersion
	}
	if strings.TrimSpace(cfg.User) == "" {
		cfg.User = defaultPostgresUser
	}
	if strings.TrimSpace(cfg.Database) == "" {
		cfg.Database = defaultPostgresDB
	}
	if cfg.MemLimitMB <= 0 {
		cfg.MemLimitMB = defaultMemLimitMB
	}
	pw, err := decryptField(masterKey, cfg.PasswordEnc)
	if err != nil {
		return PostgresConfig{}, "", err
	}
	return cfg, pw, nil
}

// ParseRedis validates redis config and returns plaintext password.
func ParseRedis(raw json.RawMessage, masterKey []byte) (RedisConfig, string, error) {
	var cfg RedisConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return RedisConfig{}, "", err
	}
	if cfg.Version <= 0 {
		cfg.Version = defaultRedisVersion
	}
	if cfg.MaxMemoryMB <= 0 {
		cfg.MaxMemoryMB = 256
	}
	pw, err := decryptField(masterKey, cfg.PasswordEnc)
	if err != nil {
		return RedisConfig{}, "", err
	}
	return cfg, pw, nil
}

// RedactConfig removes secret fields from config JSON for API responses.
func RedactConfig(kind string, raw json.RawMessage) json.RawMessage {
	switch kind {
	case "postgres":
		var cfg PostgresConfig
		if json.Unmarshal(raw, &cfg) != nil {
			return raw
		}
		cfg.PasswordEnc = ""
		out, _ := json.Marshal(cfg)
		return out
	case "redis":
		var cfg RedisConfig
		if json.Unmarshal(raw, &cfg) != nil {
			return raw
		}
		cfg.PasswordEnc = ""
		out, _ := json.Marshal(cfg)
		return out
	case "bucket":
		var cfg BucketConfig
		if json.Unmarshal(raw, &cfg) != nil {
			return raw
		}
		for i := range cfg.Keys {
			cfg.Keys[i].SecretEnc = ""
		}
		out, _ := json.Marshal(cfg)
		return out
	default:
		return raw
	}
}

func decryptField(masterKey []byte, b64 string) (string, error) {
	b64 = strings.TrimSpace(b64)
	if b64 == "" {
		return "", fmt.Errorf("missing password_enc")
	}
	enc, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	plain, err := crypto.Decrypt(masterKey, enc)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
