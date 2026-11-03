package storage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

const (
	garageImage   = "dxflrs/garage:v1.2.0"
	garageVersion = "1.2.0"
)

// InstallGarage writes config, starts dy-garage, and records encrypted credentials.
func InstallGarage(ctx context.Context, cfg config.Config, q *sqlc.Queries, dc *docker.Client) error {
	if err := ValidateInstallRequest("garage"); err != nil {
		return err
	}
	if dc == nil {
		return fmt.Errorf("docker unavailable")
	}
	st, err := LoadGarage(ctx, q)
	if err != nil {
		return err
	}
	if st.Installed {
		return EnsureGarageRunning(ctx, cfg, dc)
	}

	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return err
	}
	rpcSecret, err := randomHex(32)
	if err != nil {
		return err
	}
	adminToken, err := randomHex(32)
	if err != nil {
		return err
	}
	if err := writeGarageConfig(cfg, rpcSecret); err != nil {
		return err
	}
	if err := dc.EnsureEdgeNetwork(ctx); err != nil {
		return err
	}
	if err := dc.RunGarage(ctx, cfg, rpcSecret); err != nil {
		return err
	}
	adminEnc, err := crypto.Encrypt(key, []byte(adminToken))
	if err != nil {
		return err
	}
	rpcEnc, err := crypto.Encrypt(key, []byte(rpcSecret))
	if err != nil {
		return err
	}
	return SaveGarage(ctx, q, GarageState{
		Installed:     true,
		Provider:      "garage",
		Version:       garageVersion,
		AdminTokenEnc: base64.StdEncoding.EncodeToString(adminEnc),
		RPCSecretEnc:  base64.StdEncoding.EncodeToString(rpcEnc),
	})
}

// EnsureGarageRunning starts dy-garage when settings say installed.
func EnsureGarageRunning(ctx context.Context, cfg config.Config, dc *docker.Client) error {
	if dc == nil {
		return nil
	}
	list, err := dc.ListManaged(ctx)
	if err != nil {
		return err
	}
	for _, c := range list {
		if c.Name == "dy-garage" && c.State == "running" {
			return nil
		}
	}
	rpcSecret, err := readRPCSecret(cfg)
	if err != nil {
		return err
	}
	return dc.RunGarage(ctx, cfg, rpcSecret)
}

func writeGarageConfig(cfg config.Config, rpcSecret string) error {
	root := filepath.Join(cfg.DataDir, "garage")
	meta := filepath.Join(root, "meta")
	data := filepath.Join(root, "data")
	for _, p := range []string{meta, data, filepath.Join(cfg.ConfigDir, "garage")} {
		if err := os.MkdirAll(p, 0o750); err != nil {
			return err
		}
	}
	toml := fmt.Sprintf(`metadata_dir = "/var/lib/garage/meta"
data_dir = "/var/lib/garage/data"
db_engine = "sqlite"
replication_factor = 1
rpc_bind_addr = "[::]:3901"
rpc_public_addr = "dy-garage:3901"
rpc_secret = "%s"

[s3_api]
s3_region = "garage"
api_bind_addr = "[::]:3900"
`, rpcSecret)
	path := filepath.Join(cfg.ConfigDir, "garage", "garage.toml")
	return os.WriteFile(path, []byte(toml), 0o600)
}

func readRPCSecret(cfg config.Config) (string, error) {
	raw, err := os.ReadFile(filepath.Join(cfg.ConfigDir, "garage", "garage.toml"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `rpc_secret = "`) {
			return strings.TrimSuffix(strings.TrimPrefix(line, `rpc_secret = "`), `"`), nil
		}
	}
	return "", fmt.Errorf("rpc_secret not found in garage.toml")
}

func randomHex(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
