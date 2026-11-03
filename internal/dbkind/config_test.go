package dbkind

import (
	"encoding/json"
	"testing"
)

func TestNewPostgresConfigRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	raw, err := NewPostgresConfigJSON(key, 17)
	if err != nil {
		t.Fatal(err)
	}
	cfg, pw, err := ParsePostgres(raw, key)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 17 || pw == "" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	vars := Published("postgres", "db", pw, cfg)
	if vars["DATABASE_URL"] == "" || vars["PGHOST"] != "db" {
		t.Fatalf("vars: %+v", vars)
	}
	redacted := RedactConfig("postgres", raw)
	var check PostgresConfig
	if err := json.Unmarshal(redacted, &check); err != nil {
		t.Fatal(err)
	}
	if check.PasswordEnc != "" {
		t.Fatal("expected redacted password")
	}
}

func TestNewRedisConfigRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	raw, err := NewRedisConfigJSON(key, 8)
	if err != nil {
		t.Fatal(err)
	}
	_, pw, err := ParseRedis(raw, key)
	if err != nil {
		t.Fatal(err)
	}
	vars := Published("redis", "cache", pw, PostgresConfig{})
	if vars["REDIS_URL"] == "" || vars["REDIS_HOST"] != "cache" {
		t.Fatalf("vars: %+v", vars)
	}
}
