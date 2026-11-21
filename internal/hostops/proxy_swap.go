package hostops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
	"github.com/Joeydekruis/dockyrd/internal/proxybuild"
)

// ProxySwap switches dy-proxy to the Dockyrd proxybuild image (blue/green with retained container).
type ProxySwap struct {
	Cfg       config.Config
	TargetRef string
}

func NewProxySwap(cfg config.Config) *ProxySwap {
	return &ProxySwap{Cfg: cfg}
}

func (p *ProxySwap) WithTargetRef(ref string) *ProxySwap {
	p.TargetRef = strings.TrimSpace(ref)
	return p
}

func (p *ProxySwap) Kind() string { return KindProxySwap }

func (p *ProxySwap) Plan(_ context.Context) (Plan, error) {
	target := p.TargetRef
	if target == "" {
		var err error
		target, err = proxybuild.ImageRefForVersion(proxybuild.PanelVersion)
		if err != nil {
			return nil, err
		}
	}
	current := docker.EffectiveProxyImage(p.Cfg)
	return Plan{
		"summary":        "Swap edge proxy to proxybuild image with config validation",
		"current_image":  current,
		"target_image":   target,
		"panel_version":  proxybuild.PanelVersion,
		"retain_until_ok": true,
	}, nil
}

func (p *ProxySwap) Apply(ctx context.Context) (RevertPayload, error) {
	target := p.TargetRef
	if target == "" {
		var err error
		target, err = proxybuild.ImageRefForVersion(proxybuild.PanelVersion)
		if err != nil {
			return nil, err
		}
	}
	conn, q, err := db.Open(p.Cfg.DatabasePath())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	cli, err := docker.NewClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	prev := docker.EffectiveProxyImage(p.Cfg)
	res, err := proxy.SwapEdgeImage(ctx, cli, p.Cfg, q, target)
	if err != nil {
		return nil, err
	}
	if res.PreviousImage == "" {
		res.PreviousImage = prev
	}
	raw, err := proxy.EncodeSwapRevert(res.PreviousImage, res.NewImage)
	if err != nil {
		return nil, err
	}
	return RevertPayload(raw), nil
}

func (p *ProxySwap) Verify(ctx context.Context) error {
	cli, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	ok, detail := cli.ProxyRunning(ctx)
	if !ok {
		return fmt.Errorf("dy-proxy not healthy: %s", detail)
	}
	return nil
}

func (p *ProxySwap) Revert(ctx context.Context, payload RevertPayload) error {
	meta, err := proxy.DecodeSwapRevert([]byte(payload))
	if err != nil {
		return err
	}
	if strings.TrimSpace(meta.PreviousImage) == "" {
		return errors.New("revert missing previous_image")
	}
	conn, q, err := db.Open(p.Cfg.DatabasePath())
	if err != nil {
		return err
	}
	defer conn.Close()
	cli, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	return proxy.RollbackEdgeImage(ctx, cli, p.Cfg, q, meta.PreviousImage)
}

func (p *ProxySwap) RequiresLockoutConfirm() bool { return false }

// ProxySwapFromPlan decodes an optional target_image from a stored plan JSON.
func ProxySwapFromPlan(cfg config.Config, planJSON string) *ProxySwap {
	p := NewProxySwap(cfg)
	if planJSON == "" {
		return p
	}
	var plan map[string]any
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return p
	}
	if v, ok := plan["target_image"].(string); ok {
		p.TargetRef = v
	}
	return p
}