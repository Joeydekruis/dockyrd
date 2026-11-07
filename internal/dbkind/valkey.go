package dbkind

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	RedisPresetCache = "cache"
	RedisPresetQueue = "queue"
)

// ResolveRedisPreset maps legacy persistence flag to a preset when preset is unset.
func ResolveRedisPreset(cfg RedisConfig) string {
	p := strings.TrimSpace(strings.ToLower(cfg.Preset))
	if p == RedisPresetCache || p == RedisPresetQueue {
		return p
	}
	if cfg.Persistence {
		return RedisPresetQueue
	}
	return RedisPresetCache
}

// ValkeyServerArgs builds valkey-server CLI args for container MemLimitMB.
func ValkeyServerArgs(cfg RedisConfig, password string, containerMemMB int) []string {
	if containerMemMB <= 0 {
		containerMemMB = 512
	}
	maxMemBytes := containerMemMB * 1024 * 1024 * 75 / 100
	args := []string{
		"valkey-server",
		"--requirepass", password,
		"--protected-mode", "yes",
		"--maxmemory", strconv.FormatInt(int64(maxMemBytes), 10),
	}
	switch ResolveRedisPreset(cfg) {
	case RedisPresetCache:
		args = append(args, "--maxmemory-policy", "allkeys-lru")
	default:
		args = append(args,
			"--maxmemory-policy", "noeviction",
			"--appendonly", "yes",
			"--appendfsync", "everysec",
		)
	}
	return args
}

// RedisPresetLabel returns a human label for API/docs.
func RedisPresetLabel(preset string) string {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case RedisPresetCache:
		return "Cache"
	case RedisPresetQueue:
		return "Queue / sessions"
	default:
		return fmt.Sprintf("%q", preset)
	}
}
