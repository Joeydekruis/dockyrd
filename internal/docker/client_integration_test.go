//go:build integration

package docker

import (
	"context"
	"testing"
)

func TestEnsureNetworksIntegration(t *testing.T) {
	ctx := context.Background()
	cli, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Close() })
	if err := cli.Ping(ctx); err != nil {
		t.Skip("docker not available:", err)
	}
	if err := cli.EnsureEdgeNetwork(ctx); err != nil {
		t.Fatal(err)
	}
	if err := cli.EnsureEdgeNetwork(ctx); err != nil {
		t.Fatal("second ensure should be idempotent:", err)
	}
	if err := cli.EnsureProjectNetwork(ctx, "demo"); err != nil {
		t.Fatal(err)
	}
}
