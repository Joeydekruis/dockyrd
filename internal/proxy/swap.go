package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/proxybuild"
	"github.com/docker/docker/errdefs"
)

func isNotFound(err error) bool {
	return errdefs.IsNotFound(err)
}

// SwapResult captures proxy.swap outcome.
type SwapResult struct {
	PreviousImage string
	NewImage      string
}

// SwapRevertPayload is persisted by hostops proxy.swap.
type SwapRevertPayload struct {
	PreviousImage string `json:"previous_image"`
	NewImage      string `json:"new_image"`
}

// EncodeSwapRevert JSON-encodes revert metadata for proxy.swap.
func EncodeSwapRevert(previous, newImage string) ([]byte, error) {
	return json.Marshal(SwapRevertPayload{PreviousImage: previous, NewImage: newImage})
}

// DecodeSwapRevert parses proxy.swap revert metadata.
func DecodeSwapRevert(raw []byte) (SwapRevertPayload, error) {
	var p SwapRevertPayload
	if len(raw) == 0 {
		return p, errors.New("empty revert payload")
	}
	err := json.Unmarshal(raw, &p)
	return p, err
}

// SwapEdgeImage pulls targetImage, validates config, recreates dy-proxy, and reloads routes.
func SwapEdgeImage(ctx context.Context, c *docker.Client, cfg config.Config, q *sqlc.Queries, targetImage string) (SwapResult, error) {
	if c == nil {
		return SwapResult{}, errors.New("docker client required")
	}
	targetImage = strings.TrimSpace(targetImage)
	if targetImage == "" {
		return SwapResult{}, errors.New("target image required")
	}
	prev := docker.EffectiveProxyImage(cfg)
	if prev == targetImage {
		return SwapResult{PreviousImage: prev, NewImage: targetImage}, nil
	}

	if err := c.PullImage(ctx, targetImage); err != nil {
		return SwapResult{}, fmt.Errorf("pull %s: %w", targetImage, err)
	}
	if err := proxybuild.VerifyImageSignature(ctx, targetImage); err != nil {
		return SwapResult{}, err
	}

	configJSON, err := RenderFromDB(ctx, q, cfg, c)
	if err != nil {
		return SwapResult{}, fmt.Errorf("render proxy config: %w", err)
	}
	if err := c.ValidateCaddyJSON(ctx, targetImage, cfg, configJSON); err != nil {
		return SwapResult{}, fmt.Errorf("validate with new build: %w", err)
	}

	if err := c.ParkProxyForSwap(ctx); err != nil && !isNotFound(err) {
		return SwapResult{}, fmt.Errorf("park current proxy: %w", err)
	}

	if err := c.RecreateProxyWithImage(ctx, cfg, targetImage); err != nil {
		_ = c.RestoreParkedProxy(ctx, cfg)
		return SwapResult{}, err
	}

	if err := docker.WaitForCaddyAdmin(ctx, cfg.CaddyAdminSocket(), 45*time.Second); err != nil {
		_ = c.RemoveProxyContainer(ctx)
		_ = c.RestoreParkedProxy(ctx, cfg)
		return SwapResult{}, fmt.Errorf("new proxy admin socket: %w", err)
	}

	if err := Load(ctx, cfg.CaddyAdminSocket(), configJSON); err != nil {
		_ = c.RemoveProxyContainer(ctx)
		_ = c.RestoreParkedProxy(ctx, cfg)
		return SwapResult{}, fmt.Errorf("load config on new proxy: %w", err)
	}

	if err := c.RemoveParkedProxy(ctx); err != nil {
		return SwapResult{PreviousImage: prev, NewImage: targetImage}, err
	}

	st := docker.ProxyImageState{Image: targetImage, Proxybuild: strings.Contains(targetImage, "dockyrd-proxy")}
	_ = docker.WriteProxyImageState(cfg, st)

	return SwapResult{PreviousImage: prev, NewImage: targetImage}, nil
}

// RollbackEdgeImage recreates dy-proxy from a previous image ref and reloads routes.
func RollbackEdgeImage(ctx context.Context, c *docker.Client, cfg config.Config, q *sqlc.Queries, previousImage string) error {
	previousImage = strings.TrimSpace(previousImage)
	if previousImage == "" {
		previousImage = docker.ProxyImage
	}
	_, err := SwapEdgeImage(ctx, c, cfg, q, previousImage)
	return err
}