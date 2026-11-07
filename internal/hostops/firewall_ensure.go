package hostops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const nftRuleset = `table inet dockyrd {
	chain input {
		type filter hook input priority 0; policy drop;
		ct state established,related accept
		iif lo accept
		tcp dport { 22, 80, 443 } accept
	}
}
`

// FirewallEnsure applies a baseline nftables ruleset (idempotent).
type FirewallEnsure struct {
	Env HostEnv
}

func NewFirewallEnsure(env HostEnv) *FirewallEnsure {
	return &FirewallEnsure{Env: env}
}

func (f *FirewallEnsure) Kind() string { return KindFirewallEnsure }

func (f *FirewallEnsure) Plan(_ context.Context) (Plan, error) {
	return Plan{
		"summary":       "Allow SSH (22), HTTP (80), HTTPS (443); default drop on INPUT",
		"ruleset":       nftRuleset,
		"docker_caveat": "Published container ports may bypass INPUT; review DOCKER-USER if needed",
		"path":          f.Env.NFTRulesPath(),
	}, nil
}

func (f *FirewallEnsure) Apply(ctx context.Context) (RevertPayload, error) {
	path := f.Env.NFTRulesPath()
	prev := ""
	if data, err := f.Env.ReadFile(path); err == nil {
		prev = string(data)
	}
	if err := f.Env.WriteFile(path, []byte(nftRuleset), 0o644); err != nil {
		return nil, err
	}
	if f.Env.HostRoot() == "" {
		if out, err := f.Env.Exec.Run(ctx, "nft", "-f", path); err != nil {
			_ = restoreHostFile(f.Env, path, prev)
			return nil, fmt.Errorf("nft -f: %w (%s)", err, strings.TrimSpace(string(out)))
		}
	}
	b, err := json.Marshal(map[string]string{"path": path, "previous": prev})
	if err != nil {
		return nil, err
	}
	return RevertPayload(b), nil
}

func (f *FirewallEnsure) Verify(ctx context.Context) error {
	path := f.Env.NFTRulesPath()
	data, err := f.Env.ReadFile(path)
	if err != nil {
		return err
	}
	if !strings.Contains(string(data), "table inet dockyrd") {
		return fmt.Errorf("dockyrd table missing in %s", path)
	}
	if f.Env.HostRoot() == "" {
		out, err := f.Env.Exec.Run(ctx, "nft", "list", "table", "inet", "dockyrd")
		if err != nil {
			return fmt.Errorf("nft list: %w", err)
		}
		if !strings.Contains(string(out), "dockyrd") {
			return fmt.Errorf("active nft table not found")
		}
	}
	return nil
}

func (f *FirewallEnsure) Revert(_ context.Context, payload RevertPayload) error {
	var meta struct {
		Path     string `json:"path"`
		Previous string `json:"previous"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return err
	}
	if meta.Previous == "" {
		return os.Remove(meta.Path)
	}
	return f.Env.WriteFile(meta.Path, []byte(meta.Previous), 0o644)
}

func (f *FirewallEnsure) RequiresLockoutConfirm() bool { return true }
