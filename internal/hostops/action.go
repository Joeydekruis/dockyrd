package hostops

import (
	"context"
	"encoding/json"
	"errors"
)

var ErrNotImplemented = errors.New("host action kind not implemented")

// Plan is a serializable preview stored in plan_json.
type Plan map[string]any

// RevertPayload is opaque state needed to undo Apply.
type RevertPayload json.RawMessage

// Action is a typed host mutation (no shell strings).
type Action interface {
	Kind() string
	Plan(ctx context.Context) (Plan, error)
	Apply(ctx context.Context) (RevertPayload, error)
	Verify(ctx context.Context) error
	Revert(ctx context.Context, payload RevertPayload) error
	RequiresLockoutConfirm() bool
}

// Kind constants for v1 registered actions.
const (
	KindFirewallEnsure        = "firewall.ensure"
	KindSSHPasswordAuthDisable = "ssh.passwordauth.disable"
	KindUnattendedEnable      = "unattended.enable"
	KindCrowdsecBouncerInstall = "crowdsec.bouncer.install"
	KindProxySwap             = "proxy.swap"
	KindTestEcho              = "test.echo"
)

func encodePlan(p Plan) (string, error) {
	if p == nil {
		return "", nil
	}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func encodeRevert(p RevertPayload) (string, error) {
	if len(p) == 0 {
		return "", nil
	}
	return string(p), nil
}

func decodeRevert(s string) RevertPayload {
	if s == "" {
		return nil
	}
	return RevertPayload(s)
}
