"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/dockyrd/ui/button";
import { EmptyState } from "@/components/dockyrd/ui/empty-state";
import { RadioGroup } from "@/components/dockyrd/ui/radio-group";
import { Switch } from "@/components/dockyrd/ui/switch";
import {
  clearServiceBasicAuth,
  getServiceSecurity,
  putServiceSecurity,
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
  const [busy, setBusy] = useState(false);
  const [redeployNote, setRedeployNote] = useState(false);
  const [basicUser, setBasicUser] = useState("");
  const [revealedPassword, setRevealedPassword] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      const payload = await getServiceSecurity(serviceId);
      setData(payload);
      setCsp(payload.stored.csp ?? "");
      setAllowlistText((payload.stored.ip_allowlist ?? []).join("\n"));
      setBasicUser(payload.stored.basic_user ?? "");
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load security");
    }
  }, [serviceId]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

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
      };
      if (kind === "app") {
        body.mem_mb = mem;
        body.cpu = cpu;
        body.pids = pids;
      }
      const next = await putServiceSecurity(serviceId, body);
      setData(next);
      setRedeployNote(kind === "app");
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

  const proxyReady =
    data.effective.proxy_rate_limit_installed ||
    data.effective.proxy_waf_installed;

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

      <section className="space-y-3 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6 opacity-90">
        <h2 className="text-lg font-medium">Traffic protection</h2>
        <p className="text-sm text-[var(--muted-text)]">
          {proxyReady
            ? "Rate limit and WAF handlers ship with the proxy build (P2)."
            : "Install the proxy build from Add-ons to enable rate limiting and WAF."}
        </p>
        {!proxyReady ? (
          <Link
            href="/server/security/addons"
            className="text-sm text-[var(--accent-blue)] underline"
          >
            View add-ons
          </Link>
        ) : null}
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
        <EmptyState
          title="No events yet"
          description="Events appear when rate limiting or WAF is enabled."
        />
      </section>

      <Button onClick={onSave} disabled={busy}>
        {busy ? "Saving…" : "Save security"}
      </Button>
    </div>
  );
}
