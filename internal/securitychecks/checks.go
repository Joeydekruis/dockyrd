package securitychecks

import (
	"context"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

// Deps bundles detection dependencies.
type Deps struct {
	Cfg    config.Config
	Q      *sqlc.Queries
	Docker *docker.Client
	Env    hostops.HostEnv
	Now    func() time.Time
}

func (d Deps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now()
}

// RunAll executes the eight baseline checks.
func RunAll(ctx context.Context, deps Deps) Report {
	ts := deps.now().Unix()
	out := make([]Result, 0, 8)
	out = append(out, checkFirewallActive(ctx, deps, ts))
	out = append(out, checkDockerPublishedPorts(ctx, deps, ts))
	out = append(out, checkSSHPasswordAuth(ctx, deps, ts))
	out = append(out, checkUnattended(ctx, deps, ts))
	out = append(out, checkRebootPending(deps, ts))
	out = append(out, checkPanel2FA(ctx, deps, ts))
	out = append(out, checkPanelHTTPS(ctx, deps, ts))
	out = append(out, checkProxyDefaults(deps, ts))
	return Report{RanAt: ts, Checks: out}
}

func checkFirewallActive(ctx context.Context, deps Deps, ts int64) Result {
	r := Result{
		ID: "firewall.active", Title: "Host firewall", Severity: "high",
		Fix: FixHostAction, FixKind: hostops.KindFirewallEnsure, LastRunAt: ts,
	}
	data, err := deps.Env.ReadNFTRules()
	if err != nil {
		if runtime.GOOS != "linux" {
			r.Status, r.Detail = StatusWarn, "Firewall state not inspectable on this dev host"
			r.Fix = FixNone
			return r
		}
		r.Status, r.Detail = StatusFail, "Firewall rules not installed"
		return r
	}
	if !strings.Contains(string(data), "table inet dockyrd") {
		r.Status, r.Detail = StatusFail, "dockyrd nftables table missing"
		return r
	}
	if deps.Env.HostRoot() == "" {
		out, err := deps.Env.Exec.Run(ctx, "nft", "list", "table", "inet", "dockyrd")
		if err != nil {
			r.Status, r.Detail = StatusWarn, "Rules file present but nft table inactive"
			return r
		}
		if !strings.Contains(string(out), "dockyrd") {
			r.Status, r.Detail = StatusWarn, "Rules file present but nft table inactive"
			return r
		}
	}
	r.Status, r.Detail = StatusOK, "dockyrd firewall table active"
	r.Fix = FixNone
	return r
}

func checkDockerPublishedPorts(ctx context.Context, deps Deps, ts int64) Result {
	r := Result{
		ID: "docker.published_ports", Title: "Docker published ports", Severity: "medium",
		Fix: FixNone, LastRunAt: ts,
	}
	if deps.Docker == nil {
		r.Status, r.Detail = StatusWarn, "Docker not connected"
		return r
	}
	exposures, err := deps.Docker.ListPublishedHostPorts(ctx)
	if err != nil {
		r.Status, r.Detail = StatusWarn, err.Error()
		return r
	}
	var bad []map[string]string
	allowed := map[string]struct{}{"80": {}, "443": {}}
	for _, e := range exposures {
		if _, ok := allowed[e.Port]; ok {
			continue
		}
		bad = append(bad, map[string]string{
			"container": e.Container,
			"port":      e.Port,
			"href":      "/applications/" + e.ServiceSlug,
		})
	}
	if len(bad) == 0 {
		r.Status, r.Detail = StatusOK, "No unexpected host-bound ports"
		return r
	}
	r.Status = StatusFail
	r.Detail = "Containers publish host ports outside 80/443"
	r.Extra = bad
	return r
}

func checkSSHPasswordAuth(ctx context.Context, deps Deps, ts int64) Result {
	r := Result{
		ID: "ssh.password_auth", Title: "SSH password authentication", Severity: "high",
		Fix: FixHostAction, FixKind: hostops.KindSSHPasswordAuthDisable, LastRunAt: ts,
	}
	if deps.Env.HostRoot() != "" {
		r.Status, r.Detail = StatusWarn, "sshd not inspected on dev host"
		return r
	}
	enabled, err := hostops.SSHPasswordAuthEnabled(ctx, deps.Env.Exec)
	if err != nil {
		r.Status, r.Detail = StatusWarn, "Cannot query sshd"
		r.Fix = FixNone
		return r
	}
	if enabled {
		r.Status, r.Detail = StatusFail, "Password authentication is enabled"
		return r
	}
	r.Status, r.Detail = StatusOK, "Password authentication disabled"
	r.Fix = FixNone
	return r
}

func checkUnattended(ctx context.Context, deps Deps, ts int64) Result {
	r := Result{
		ID: "updates.unattended", Title: "Unattended security updates", Severity: "medium",
		Fix: FixHostAction, FixKind: hostops.KindUnattendedEnable, LastRunAt: ts,
	}
	data, err := deps.Env.ReadFile(deps.Env.OSReleasePath())
	if err != nil && deps.Env.HostRoot() == "" {
		r.Status, r.Detail = StatusWarn, "Cannot read distro"
		r.Fix = FixNone
		return r
	}
	s := strings.ToLower(string(data))
	if deps.Env.HostRoot() == "" && !strings.Contains(s, "debian") && !strings.Contains(s, "ubuntu") {
		r.Status, r.Detail = StatusWarn, "Enable manually on non-Debian/Ubuntu hosts"
		r.Fix = FixNone
		return r
	}
	if _, err := os.Stat(deps.Env.UnattendedStampPath()); err != nil {
		r.Status, r.Detail = StatusFail, "Unattended upgrades not enabled"
		return r
	}
	r.Status, r.Detail = StatusOK, "Unattended upgrades enabled"
	r.Fix = FixNone
	_ = ctx
	return r
}

func checkRebootPending(deps Deps, ts int64) Result {
	r := Result{
		ID: "reboot.pending", Title: "Pending reboot", Severity: "low",
		Fix: FixNone, LastRunAt: ts,
	}
	if _, err := os.Stat(deps.Env.RebootRequiredPath()); err != nil {
		r.Status, r.Detail = StatusOK, "No reboot required"
		return r
	}
	r.Status, r.Detail = StatusWarn, "/var/run/reboot-required is present"
	return r
}

func checkPanel2FA(ctx context.Context, deps Deps, ts int64) Result {
	r := Result{
		ID: "panel.2fa", Title: "Require 2FA for all users", Severity: "medium",
		Fix: FixLink, FixHref: "/settings/server/team", LastRunAt: ts,
	}
	panel, err := instance.LoadPanel(ctx, deps.Q)
	if err != nil {
		r.Status, r.Detail = StatusWarn, "Cannot load instance settings"
		r.Fix = FixNone
		return r
	}
	if panel.Require2FAAll {
		r.Status, r.Detail = StatusOK, "Instance requires 2FA for all users"
		r.Fix = FixNone
		return r
	}
	r.Status, r.Detail = StatusWarn, "2FA not required for all users"
	return r
}

func checkPanelHTTPS(ctx context.Context, deps Deps, ts int64) Result {
	r := Result{
		ID: "panel.https", Title: "Panel served over HTTPS", Severity: "medium",
		Fix: FixLink, FixHref: "/settings/server/general", LastRunAt: ts,
	}
	panel, err := instance.LoadPanel(ctx, deps.Q)
	if err != nil {
		r.Status, r.Detail = StatusWarn, "Cannot load instance settings"
		r.Fix = FixNone
		return r
	}
	u := strings.TrimSpace(panel.PublicURL)
	if u == "" {
		u = deps.Cfg.PublicURL()
	}
	if strings.HasPrefix(strings.ToLower(u), "https://") {
		r.Status, r.Detail = StatusOK, "Public URL uses HTTPS"
		r.Fix = FixNone
		return r
	}
	r.Status, r.Detail = StatusWarn, "Public URL is not HTTPS"
	return r
}

func checkProxyDefaults(deps Deps, ts int64) Result {
	r := Result{
		ID: "proxy.defaults", Title: "Proxy baseline defaults", Severity: "medium",
		Fix: FixLink, FixHref: "/settings/server/updates", LastRunAt: ts,
	}
	raw, err := proxy.Render(proxy.RenderInput{
		ACMEEmail:     "checks@dockyrd.local",
		PanelHost:     "checks.dockyrd.local",
		PanelUpstream: "127.0.0.1:1",
	})
	if err != nil {
		r.Status, r.Detail = StatusFail, err.Error()
		return r
	}
	s := string(raw)
	for _, frag := range []string{`"handler": "encode"`, `Strict-Transport-Security`, `nosniff`, `read_header_timeout`} {
		if !strings.Contains(s, frag) {
			r.Status, r.Detail = StatusFail, "Rendered proxy missing baseline: "+frag
			return r
		}
	}
	r.Status, r.Detail = StatusOK, "Renderer includes encode, headers, and timeouts"
	r.Fix = FixNone
	return r
}
