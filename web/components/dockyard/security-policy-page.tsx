"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyrd/dashboard-shell";
import { SecurityNav } from "@/components/dockyrd/security-nav";
import { Button } from "@/components/dockyrd/ui/button";
import { PageHeader } from "@/components/dockyrd/ui/page-header";
import { RadioGroup } from "@/components/dockyrd/ui/radio-group";
import { Switch } from "@/components/dockyrd/ui/switch";
import {
  getSecurityPolicy,
  type HeaderEntry,
  putSecurityPolicy,
  type SecurityPolicy,
} from "@/lib/security-policy-api";

const cardClass =
  "space-y-4 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6";

export function SecurityPolicyPage() {
  const [policy, setPolicy] = useState<SecurityPolicy | null>(null);
  const [headers, setHeaders] = useState<Record<string, HeaderEntry>>({});
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      const p = await getSecurityPolicy();
      setPolicy(p);
      setHeaders(p.headers ?? {});
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load policy");
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onSave() {
    if (!policy) return;
    setBusy(true);
    setMessage(null);
    try {
      const next = await putSecurityPolicy({
        ...policy,
        headers,
      });
      setPolicy(next);
      setHeaders(next.headers);
      setMessage("Policy saved.");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Save failed");
    } finally {
      setBusy(false);
    }
  }

  if (!policy) {
    return (
      <DashboardShell title="Security">
        <p className="text-sm text-[var(--muted-text)]">
          {error ?? "Loading policy…"}
        </p>
      </DashboardShell>
    );
  }

  const proxyReady =
    policy.proxy_rate_limit_installed || policy.proxy_waf_installed;

  return (
    <DashboardShell title="Security">
      <div className="flex flex-col gap-6 p-8">
        <PageHeader
          title="Policy"
          subtitle="Instance policy and defaults for every site."
        />
        <SecurityNav active="policy" />
        <div className="mt-8 space-y-6">
          {error ? (
            <p className="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-700">
              {error}
            </p>
          ) : null}
          {message ? (
            <p className="text-sm text-[var(--ink-navy)]" role="status">
              {message}
            </p>
          ) : null}

          <section className={cardClass}>
            <RadioGroup
              name="default-profile"
              legend="Default profile"
              value={policy.default_profile}
              onChange={(v) => setPolicy({ ...policy, default_profile: v })}
              options={[
                {
                  value: "baseline",
                  label: "Baseline",
                  description: "Standard headers and limits.",
                },
                {
                  value: "strict",
                  label: "Strict",
                  description: "Longer HSTS and Permissions-Policy.",
                },
              ]}
            />
            <Switch
              label="Policy lock"
              checked={policy.lock}
              onChange={(lock) => setPolicy({ ...policy, lock })}
            />
            <p className="text-sm text-[var(--muted-text)]">
              When lock is on, sites cannot be weaker than this instance policy
              unless an admin grants a per-site exemption.
            </p>
          </section>

          <section className={cardClass}>
            <h2 className="text-lg font-medium">Headers</h2>
            {policy.header_catalog
              .filter((h) => h !== "Content-Security-Policy")
              .map((name) => {
                const entry = headers[name] ?? { enabled: false, value: "" };
                return (
                  <div
                    key={name}
                    className="flex flex-col gap-2 border-b border-[var(--border)] pb-4"
                  >
                    <Switch
                      label={name}
                      checked={entry.enabled}
                      onChange={(enabled) =>
                        setHeaders({
                          ...headers,
                          [name]: { ...entry, enabled },
                        })
                      }
                    />
                    <input
                      className="rounded-lg border border-[var(--border)] bg-[var(--surface)] px-3 py-2 text-sm"
                      value={entry.value}
                      disabled={!entry.enabled}
                      onChange={(e) =>
                        setHeaders({
                          ...headers,
                          [name]: { ...entry, value: e.target.value },
                        })
                      }
                    />
                  </div>
                );
              })}
          </section>

          <section className={cardClass}>
            <h2 className="text-lg font-medium">Rate limits</h2>
            <p className="text-sm text-[var(--muted-text)]">
              {proxyReady
                ? "Stored for when the proxy build is active (P2)."
                : "Needs the proxy build — install rate limiting from Add-ons."}
            </p>
            <label className="text-sm">
              Default RPM
              <input
                type="number"
                disabled={!proxyReady}
                className="mt-1 block w-full rounded-lg border border-[var(--border)] px-3 py-2"
                value={policy.rl_default_rpm ?? ""}
                onChange={(e) =>
                  setPolicy({
                    ...policy,
                    rl_default_rpm: e.target.value
                      ? Number(e.target.value)
                      : undefined,
                  })
                }
              />
            </label>
          </section>

          <section className={cardClass}>
            <h2 className="text-lg font-medium">CDN / trusted proxies</h2>
            <p className="text-sm text-[var(--muted-text)]">
              Values are saved now; edge trusted-proxy render ships in a
              follow-up release.
            </p>
            <label className="text-sm">
              Mode
              <select
                className="mt-1 block w-full rounded-lg border border-[var(--border)] px-3 py-2"
                value={policy.cdn_mode}
                onChange={(e) =>
                  setPolicy({ ...policy, cdn_mode: e.target.value })
                }
              >
                <option value="direct">Direct</option>
                <option value="cloudflare">Cloudflare</option>
                <option value="custom">Custom</option>
              </select>
            </label>
          </section>

          <section className={cardClass}>
            <h2 className="text-lg font-medium">
              Default limits for new services
            </h2>
            <div className="grid gap-4 sm:grid-cols-3">
              {(
                [
                  ["default_mem_mb", "Memory (MiB)"],
                  ["default_cpu", "CPU"],
                  ["default_pids", "PIDs"],
                ] as const
              ).map(([key, label]) => (
                <label key={key} className="text-sm">
                  {label}
                  <input
                    type="number"
                    step={key === "default_cpu" ? "0.1" : "1"}
                    className="mt-1 block w-full rounded-lg border border-[var(--border)] px-3 py-2"
                    value={policy[key]}
                    onChange={(e) =>
                      setPolicy({
                        ...policy,
                        [key]:
                          key === "default_cpu"
                            ? Number(e.target.value)
                            : Number(e.target.value),
                      })
                    }
                  />
                </label>
              ))}
            </div>
          </section>

          <Button onClick={onSave} disabled={busy}>
            {busy ? "Saving…" : "Save policy"}
          </Button>
        </div>
      </div>
    </DashboardShell>
  );
}
