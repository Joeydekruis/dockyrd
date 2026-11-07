package dbkind

import (
	"strings"
	"testing"
)

func TestValkeyServerArgsQueuePreset(t *testing.T) {
	cfg := RedisConfig{Preset: RedisPresetQueue, Persistence: true}
	args := ValkeyServerArgs(cfg, "secret", 512)
	joined := strings.Join(args, " ")
	for _, want := range []string{"--appendonly", "yes", "--maxmemory-policy", "noeviction", "--protected-mode", "yes"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %v", want, args)
		}
	}
	// 75% of 512 MiB
	if !strings.Contains(joined, "--maxmemory 402653184") {
		t.Fatalf("maxmemory: %s", joined)
	}
}

func TestValkeyServerArgsCachePreset(t *testing.T) {
	cfg := RedisConfig{Preset: RedisPresetCache, Persistence: false}
	args := ValkeyServerArgs(cfg, "secret", 256)
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "--appendonly") {
		t.Fatalf("cache preset should not enable AOF: %v", args)
	}
	if !strings.Contains(joined, "allkeys-lru") {
		t.Fatalf("expected allkeys-lru: %v", args)
	}
}

func TestResolveRedisPresetLegacyPersistence(t *testing.T) {
	if ResolveRedisPreset(RedisConfig{Persistence: true}) != RedisPresetQueue {
		t.Fatal("expected queue")
	}
	if ResolveRedisPreset(RedisConfig{Persistence: false}) != RedisPresetCache {
		t.Fatal("expected cache")
	}
}
