package hostops

import (
	"fmt"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

// Registry maps action kinds to factories.
type Registry struct {
	factories map[string]func() Action
}

func NewRegistry() *Registry {
	return &Registry{factories: map[string]func() Action{}}
}

func (r *Registry) Register(kind string, factory func() Action) {
	r.factories[kind] = factory
}

func (r *Registry) New(kind string) (Action, error) {
	f, ok := r.factories[kind]
	if !ok {
		return nil, fmt.Errorf("unknown host action kind %q", kind)
	}
	return f(), nil
}

// DefaultRegistry registers v1 stub kinds plus test.echo.
func DefaultRegistry(cfg config.Config) *Registry {
	env := NewHostEnv(cfg)
	r := NewRegistry()
	r.Register(KindFirewallEnsure, func() Action { return NewFirewallEnsure(env) })
	r.Register(KindSSHPasswordAuthDisable, func() Action { return NewSSHPasswordAuthDisable(env) })
	r.Register(KindUnattendedEnable, func() Action { return NewUnattendedEnable(env) })
	r.Register(KindCrowdsecBouncerInstall, NewCrowdsecBouncerInstallStub)
	r.Register(KindProxySwap, NewProxySwapStub)
	r.Register(KindTestEcho, func() Action { return NewTestEcho(cfg) })
	return r
}
