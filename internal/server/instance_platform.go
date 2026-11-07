package server

import (
	"context"

	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

func (d apiDeps) loadPanel(ctx context.Context) (instance.Panel, error) {
	return instance.LoadPanel(ctx, d.q)
}

func (d apiDeps) platformHostInput(ctx context.Context, serviceSlug string) (proxy.PlatformHostInput, error) {
	panel, err := d.loadPanel(ctx)
	if err != nil {
		return proxy.PlatformHostInput{}, err
	}
	return proxy.PlatformHostInput{
		ServiceSlug:      serviceSlug,
		PublicIPv4:       d.cfg.EffectivePublicIPv4(),
		WildcardPattern:  panel.Wildcard,
		WildcardVerified: instance.WildcardIsVerified(panel),
	}, nil
}

func (d apiDeps) platformHostname(ctx context.Context, serviceSlug string) (string, error) {
	in, err := d.platformHostInput(ctx, serviceSlug)
	if err != nil {
		return "", err
	}
	return proxy.PlatformHostnameFor(in)
}
