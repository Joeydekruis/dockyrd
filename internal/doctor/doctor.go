package doctor

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// Status is a single check outcome.
type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

// Check is one diagnostic line.
type Check struct {
	Name    string
	Status  Status
	Detail  string
	FixHint string
}

// Report aggregates checks for exit-code handling.
type Report struct {
	Version string
	Checks  []Check
}

// HasFailures is true when any check failed.
func (r Report) HasFailures() bool {
	for _, c := range r.Checks {
		if c.Status == StatusFail {
			return true
		}
	}
	return false
}

// Run executes host and dockyrd readiness checks.
func Run(ctx context.Context, cfg config.Config, version string) Report {
	rep := Report{Version: version}
	rep.Checks = append(rep.Checks, checkLayout(cfg)...)
	rep.Checks = append(rep.Checks, checkDatabase(cfg)...)
	rep.Checks = append(rep.Checks, checkDisk(cfg)...)
	rep.Checks = append(rep.Checks, checkDocker(ctx)...)
	rep.Checks = append(rep.Checks, checkDockerAddressPools()...)
	rep.Checks = append(rep.Checks, checkProxyContainer(ctx)...)
	rep.Checks = append(rep.Checks, checkProxySocket(cfg)...)
	rep.Checks = append(rep.Checks, checkHTTPSPorts(ctx)...)
	rep.Checks = append(rep.Checks, checkListen(cfg)...)
	rep.Checks = append(rep.Checks, checkClock()...)
	rep.Checks = append(rep.Checks, checkBuildResilience(ctx, cfg)...)
	return rep
}

// Print writes a human-readable report.
func Print(w io.Writer, rep Report) {
	fmt.Fprintf(w, "dockyrd doctor (v%s)\n\n", rep.Version)
	for _, c := range rep.Checks {
		icon := "✓"
		switch c.Status {
		case StatusWarn:
			icon = "!"
		case StatusFail:
			icon = "✗"
		}
		fmt.Fprintf(w, " [%s] %s", icon, c.Name)
		if c.Detail != "" {
			fmt.Fprintf(w, " — %s", c.Detail)
		}
		fmt.Fprintln(w)
		if c.FixHint != "" && c.Status != StatusOK {
			fmt.Fprintf(w, "     → %s\n", c.FixHint)
		}
	}
}

func checkLayout(cfg config.Config) []Check {
	var out []Check
	for _, item := range []struct {
		name string
		path string
		mode os.FileMode
	}{
		{"Data directory", cfg.DataDir, 0o750},
		{"Config directory", cfg.ConfigDir, 0o700},
	} {
		st, err := os.Stat(item.path)
		if err != nil {
			out = append(out, Check{
				Name:    item.name,
				Status:  StatusFail,
				Detail:  err.Error(),
				FixHint: fmt.Sprintf("Run `dockyrd init` or create %s", item.path),
			})
			continue
		}
		if !st.IsDir() {
			out = append(out, Check{
				Name:   item.name,
				Status: StatusFail,
				Detail: "not a directory",
			})
			continue
		}
		out = append(out, Check{Name: item.name, Status: StatusOK, Detail: item.path})
	}
	keyPath := cfg.SecretKeyPath()
	if _, err := os.Stat(keyPath); err != nil {
		out = append(out, Check{
			Name:    "Master key",
			Status:  StatusFail,
			Detail:  "missing",
			FixHint: "Run `dockyrd init` to create the encryption key",
		})
	} else {
		out = append(out, Check{Name: "Master key", Status: StatusOK, Detail: keyPath})
	}
	return out
}

func checkDatabase(cfg config.Config) []Check {
	path := cfg.DatabasePath()
	if _, err := os.Stat(path); err != nil {
		return []Check{{
			Name:    "Control-plane database",
			Status:  StatusWarn,
			Detail:  "not created yet",
			FixHint: "Complete setup in the UI or run `dockyrd init`",
		}}
	}
	conn, _, err := db.Open(path)
	if err != nil {
		return []Check{{
			Name:   "Control-plane database",
			Status: StatusFail,
			Detail: err.Error(),
		}}
	}
	defer conn.Close()
	return []Check{{Name: "Control-plane database", Status: StatusOK, Detail: path}}
}

func checkDisk(cfg config.Config) []Check {
	minPct := cfg.MinFreeDiskPercent()
	if err := build.CheckDiskGuard(cfg.DataDir, minPct); err != nil {
		return []Check{{
			Name:    "Disk space (data volume)",
			Status:  StatusFail,
			Detail:  err.Error(),
			FixHint: "Free disk space or expand the volume backing " + cfg.DataDir,
		}}
	}
	return []Check{{
		Name:   "Disk space (data volume)",
		Status: StatusOK,
		Detail: fmt.Sprintf("≥ %d%% free", minPct),
	}}
}

func checkDocker(ctx context.Context) []Check {
	cli, err := docker.NewClient()
	if err != nil {
		return []Check{{
			Name:    "Docker Engine",
			Status:  StatusFail,
			Detail:  err.Error(),
			FixHint: "Install Docker and ensure the daemon is running",
		}}
	}
	defer cli.Close()
	if err := cli.Ping(ctx); err != nil {
		return []Check{{
			Name:    "Docker Engine",
			Status:  StatusFail,
			Detail:  err.Error(),
			FixHint: "Start docker.service and verify access to /var/run/docker.sock",
		}}
	}
	return []Check{{Name: "Docker Engine", Status: StatusOK, Detail: "reachable"}}
}

func checkProxyContainer(ctx context.Context) []Check {
	cli, err := docker.NewClient()
	if err != nil {
		return nil
	}
	defer cli.Close()
	ok, detail := cli.ProxyRunning(ctx)
	if !ok {
		return []Check{{
			Name:    "dy-proxy container",
			Status:  StatusWarn,
			Detail:  detail,
			FixHint: "Restart `dockyrd.service`; check `journalctl -u dockyrd` for ensure dy-proxy errors",
		}}
	}
	return []Check{{Name: "dy-proxy container", Status: StatusOK, Detail: detail}}
}

func checkHTTPSPorts(ctx context.Context) []Check {
	d := net.Dialer{Timeout: 2 * time.Second}
	for _, addr := range []string{"127.0.0.1:443", "127.0.0.1:80"} {
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return []Check{{
				Name:    "HTTPS edge (443/80)",
				Status:  StatusWarn,
				Detail:  fmt.Sprintf("nothing listening on %s", addr),
				FixHint: "Set Public URL + ACME email in Instance settings and save; open firewall 80/443; restart dockyrd",
			}}
		}
		_ = conn.Close()
	}
	return []Check{{Name: "HTTPS edge (443/80)", Status: StatusOK, Detail: "listening on localhost"}}
}

func checkProxySocket(cfg config.Config) []Check {
	sock := cfg.CaddyAdminSocket()
	if _, err := os.Stat(sock); err != nil {
		return []Check{{
			Name:    "Caddy admin socket",
			Status:  StatusWarn,
			Detail:  "not present",
			FixHint: "Start dockyrd (`dockyrd serve`) so dy-proxy loads; live traffic uses the proxy container",
		}}
	}
	return []Check{{Name: "Caddy admin socket", Status: StatusOK, Detail: sock}}
}

func checkListen(cfg config.Config) []Check {
	listen := strings.TrimSpace(cfg.Listen)
	if listen == "" {
		listen = ":8000"
	}
	return []Check{{
		Name:   "Panel listen address",
		Status: StatusOK,
		Detail: listen,
	}}
}

func checkClock() []Check {
	// ACME needs reasonable system time; we only warn on large skew vs local monotonic sanity.
	now := time.Now()
	if now.Year() < 2024 {
		return []Check{{
			Name:    "System clock",
			Status:  StatusWarn,
			Detail:  now.Format(time.RFC3339),
			FixHint: "Enable NTP (timedatectl) — incorrect time breaks HTTPS certificates",
		}}
	}
	return []Check{{Name: "System clock", Status: StatusOK, Detail: now.UTC().Format(time.RFC3339)}}
}
