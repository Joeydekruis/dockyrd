"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { RadioGroup } from "@/components/dockyard/ui/radio-group";
import { Segmented } from "@/components/dockyard/ui/segmented";
import { Switch } from "@/components/dockyard/ui/switch";
import {
  clearServiceBasicAuth,
  getServiceSecurity,
  listServiceSecurityEvents,
  putServiceSecurity,
  type SecurityEvent,
  type ServiceSecurityPayload,
  setServiceBasicAuth,
} from "@/lib/security-policy-api";

type Props = {
  serviceId: string;
  kind: "app" | "static_site";
};

export function ServiceSecurityTab({ serviceId, kind }: Props) {
  const [data, setData] = useState<ServiceSecurityPayload | null>(null);
  const [allowlistText, setAllowlistText] = useState("");
  const [csp, setCsp] = useState("");
  const [exceptionsText, setExceptionsText] = useState("");
  const [busy, setBusy] = useState(false);
  const [redeployNote, setRedeployNote] = useState(false);
  const [basicUser, setBasicUser] = useState("");
  const [revealedPassword, setRevealedPassword] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [events, setEvents] = useState<SecurityEvent[]>([]);

  const refreshEvents = useCallback(async () => {
    try {
      const res = await listServiceSecurityEvents(serviceId);
      setEvents(res.events ?? []);
    } catch {
      setEvents([]);
    }
  }, [serviceId]);

  const refresh = useCallback(async () => {
    try {
      const payload = await getServiceSecurity(serviceId);
      setData(payload);
      setCsp(payload.stored.csp ?? "");
      setAllowlistText((payload.stored.ip_allowlist ?? []).join("\n"));
      setExceptionsText((payload.stored.waf_path_exceptions ?? []).join("\n"));
      setBasicUser(payload.stored.basic_user ?? "");
      setError(null);
      await refreshEvents();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load security");
    }
  }, [serviceId, refreshEvents]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    function onFocus() {
      void refreshEvents();
    }
    window.addEventListener("focus", onFocus);
    return () => window.removeEventListener("focus", onFocus);
  }, [refreshEvents]);

  async function onSave() {
    if (!data) return;
    setBusy(true);
    try {
      const mem = data.stored.mem_mb;
      const cpu = data.stored.cpu;
      const pids = data.stored.pids;
      const body: Record<string, unknown> = {
        profile: data.stored.profile,
        csp,
        csp_report_only: data.stored.csp_report_only,
        ip_allowlist: allowlistText
          .split("\n")
          .map((l) => l.trim())
          .filter(Boolean),
        maintenance: data.stored.maintenance,
        rl_mode: data.stored.rl_mode,
        waf_mode: data.stored.waf_mode,
        waf_path_exceptions: exceptionsText
          .split("\n")
          .map((l) => l.trim())
          .filter(Boolean),
      };
      if (kind === "app") {
        body.mem_mb = mem;
        body.cpu = cpu;
        body.pids = pids;
      }
      const next = await putServiceSecurity(serviceId, body);
      setData(next);
      setExceptionsText((next.stored.waf_path_exceptions ?? []).join("\n"));
      setRedeployNote(kind === "app");
      await refreshEvents();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Save failed");
    } finally {
      setBusy(false);
    }
  }

  async function onSetBasicAuth() {
    setBusy(true);
    try {
      const res = await setServiceBasicAuth(serviceId, basicUser);
      setRevealedPassword(res.password);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Basic auth failed");
    } finally {
      setBusy(false);
    }
  }

  if (!data) {
    return (
      <p className="text-sm text-[var(--muted-text)]">{error ?? "Loading…"}</p>
    );
  }

  const wafInstalled = data.effective.proxy_waf_installed;
  const showWafControls = kind === "app";
  const wafMode = data.stored.waf_mode || "inherit";

  return (
    <div className="space-y-6">
      {error ? (
        <p className="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-700">
          {error}
        </p>
      ) : null}
      {redeployNote ? (
        <p
          className="rounded-lg border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm"
          role="status"
        >
          Resource limits changed — a redeploy was queued to apply container
          limits.
        </p>
      ) : null}

      <section className="space-y-3 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6">
        <RadioGroup
          name="site-profile"
          legend="Security profile"
          value={data.stored.profile}
          onChange={(profile) =>
            setData({ ...data, stored: { ...data.stored, profile } })
          }
          options={[
            {
              value: "inherit",
              label: "Inherit",
              description: `Instance default: ${data.effective.instance_default_profile}`,
            },
            { value: "baseline", label: "Baseline" },
            { value: "strict", label: "Strict" },
            { value: "custom", label: "Custom" },
          ]}
        />
        {data.effective.lock_floor ? (
          <p className="text-sm text-[var(--muted-text)]">
            Instance policy lock is on.
          </p>
        ) : null}
      </section>

      <section className="space-y-3 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6">
        <h2 className="text-lg font-medium">Content security</h2>
        <textarea
          className="min-h-[120px] w-full rounded-lg border border-[var(--border)] px-3 py-2 text-sm font-mono"
          value={csp}
          onChange={(e) => setCsp(e.target.value)}
          placeholder="Content-Security-Policy value"
        />
        <Switch
          label="Report only"
          checked={data.stored.csp_report_only}
          onChange={(csp_report_only) =>
            setData({ ...data, stored: { ...data.stored, csp_report_only } })
          }
        />
      </section>

      <section className="space-y-3 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6">
        <h2 className="text-lg font-medium">Access</h2>
        <label className="block text-sm">
          Basic auth username
          <input
            className="mt-1 w-full rounded-lg border border-[var(--border)] px-3 py-2"
            value={basicUser}
            onChange={(e) => setBasicUser(e.target.value)}
          />
        </label>
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            disabled={busy || !basicUser}
            onClick={onSetBasicAuth}
          >
            Set password
          </Button>
          {data.stored.basic_auth_configured ? (
            <Button
              type="button"
              variant="outline"
              disabled={busy}
              onClick={() => clearServiceBasicAuth(serviceId).then(refresh)}
            >
              Clear basic auth
            </Button>
          ) : null}
        </div>
        {revealedPassword ? (
          <p className="text-sm font-mono text-[var(--ink-navy)]">
            One-time password: {revealedPassword}
          </p>
        ) : null}
        <label className="block text-sm">
          IP allowlist (CIDR per line)
          <textarea
            className="mt-1 min-h-[100px] w-full rounded-lg border border-[var(--border)] px-3 py-2 font-mono text-sm"
            value={allowlistText}
            onChange={(e) => setAllowlistText(e.target.value)}
          />
        </label>
        <Switch
          label="Maintenance mode"
          checked={data.stored.maintenance}
          onChange={(maintenance) =>
            setData({ ...data, stored: { ...data.stored, maintenance } })
          }
        />
      </section>

      <section className="space-y-3 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6">
        <h2 className="text-lg font-medium">Traffic protection</h2>
        {showWafControls ? (
          wafInstalled ? (
            <div className="space-y-4">
              <Segmented
                legend="WAF"
                aria-label="WAF mode"
                value={
                  wafMode === "inherit"
                    ? data.effective.waf_mode || "off"
                    : wafMode
                }
                onChange={(waf_mode) =>
                  setData({ ...data, stored: { ...data.stored, waf_mode } })
                }
                options={[
                  { value: "off", label: "Off" },
                  { value: "detect", label: "Detect" },
                  { value: "block", label: "Block" },
                ]}
              />
              <p className="text-sm text-[var(--muted-text)]">
                Effective mode: {data.effective.waf_mode}. Strict profiles
                default to detect — Block is an explicit opt-in.
              </p>
              {wafMode === "block" ||
              (wafMode === "inherit" &&
                data.effective.waf_mode === "block") ? (
                <p
                  className="rounded-lg border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm"
                  role="status"
                >
                  Block can cause false positives and downtime. Prefer Detect
                  while reviewing path exceptions and security events.
                </p>
              ) : null}
              <label className="block text-sm">
                Path exceptions (prefix per line)
                <textarea
                  className="mt-1 min-h-[88px] w-full rounded-lg border border-[var(--border)] px-3 py-2 font-mono text-sm"
                  value={exceptionsText}
                  onChange={(e) => setExceptionsText(e.target.value)}
                  placeholder={"/api/\n/_next/"}
                />
              </label>
              <p className="text-sm text-[var(--muted-text)]">
                Rate limiting ships in a later proxy build — controls stay
                disabled until that add-on is installed.
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              <p className="text-sm text-[var(--muted-text)]">
                Install the proxy WAF add-on to enable Coraza detect/block and
                path exceptions.
              </p>
              <Link
                href="/server/security/addons"
                className="text-sm text-[var(--accent-blue)] underline"
              >
                View add-ons
              </Link>
            </div>
          )
        ) : (
          <p className="text-sm text-[var(--muted-text)]">
            WAF mode is not available for static sites. Security events still
            appear below when present.
          </p>
        )}
      </section>

      {kind === "app" ? (
        <section className="space-y-3 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6">
          <h2 className="text-lg font-medium">Resources</h2>
          <div className="grid gap-4 sm:grid-cols-3">
            <label className="text-sm">
              Memory (MiB)
              <input
                type="number"
                className="mt-1 w-full rounded-lg border border-[var(--border)] px-3 py-2"
                value={data.stored.mem_mb ?? data.effective.mem_mb}
                onChange={(e) =>
                  setData({
                    ...data,
                    stored: { ...data.stored, mem_mb: Number(e.target.value) },
                  })
                }
              />
            </label>
            <label className="text-sm">
              CPU
              <input
                type="number"
                step="0.1"
                className="mt-1 w-full rounded-lg border border-[var(--border)] px-3 py-2"
                value={data.stored.cpu ?? data.effective.cpu}
                onChange={(e) =>
                  setData({
                    ...data,
                    stored: { ...data.stored, cpu: Number(e.target.value) },
                  })
                }
              />
            </label>
            <label className="text-sm">
              PIDs
              <input
                type="number"
                className="mt-1 w-full rounded-lg border border-[var(--border)] px-3 py-2"
                value={data.stored.pids ?? data.effective.pids}
                onChange={(e) =>
                  setData({
                    ...data,
                    stored: { ...data.stored, pids: Number(e.target.value) },
                  })
                }
              />
            </label>
          </div>
        </section>
      ) : null}

      <section className="rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6">
        <h2 className="mb-2 text-lg font-medium">Security events</h2>
        {events.length === 0 ? (
          <EmptyState
            title="No events yet"
            description="Events appear when WAF (or rate limiting) records a hit. Client IPs are kept for 7 days for security only."
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[640px] text-left text-sm">
              <thead>
                <tr
                  className="border-b font-mono text-[11px] uppercase tracking-widest"
                  style={{
                    borderColor: "var(--border)",
                    color: "var(--muted-text)",
                  }}
                >
                  <th className="px-2 py-2 font-normal">Time</th>
                  <th className="px-2 py-2 font-normal">Source</th>
                  <th className="px-2 py-2 font-normal">Rule</th>
                  <th className="px-2 py-2 font-normal">Path</th>
                  <th className="px-2 py-2 font-normal">Action</th>
                  <th className="px-2 py-2 font-normal">IP</th>
                </tr>
              </thead>
              <tbody>
                {events.map((ev) => (
                  <tr
                    key={ev.id}
                    className="border-b"
                    style={{ borderColor: "var(--border)" }}
                    title={ev.detail}
                  >
                    <td className="px-2 py-2 font-mono text-[12px]">{ev.ts}</td>
                    <td className="px-2 py-2">{ev.source}</td>
                    <td className="px-2 py-2 font-mono text-[12px]">
                      {ev.rule || "—"}
                    </td>
                    <td className="max-w-[180px] truncate px-2 py-2 font-mono text-[12px]">
                      {ev.path || "—"}
                    </td>
                    <td className="px-2 py-2">{ev.action || "—"}</td>
                    <td className="px-2 py-2 font-mono text-[12px]">
                      {ev.ip || "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <Button onClick={onSave} disabled={busy}>
        {busy ? "Saving…" : "Save security"}
      </Button>
    </div>
  );
}
