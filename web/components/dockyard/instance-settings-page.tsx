"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  getGitHubStatus,
  getInstanceSettings,
  getServerInfo,
  patchInstanceSettings,
  startGitHubAppLink,
  testSMTP,
} from "@/lib/api";

export function InstanceSettingsPage() {
  const search = useSearchParams();
  const linkedBanner = search.get("github") === "linked";
  const [status, setStatus] = useState<{
    linked: boolean;
    slug?: string;
    installation_id?: number;
  } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [serverInfo, setServerInfo] = useState<{
    version: string;
    update_available: boolean;
    latest_tag?: string;
  } | null>(null);
  const [instance, setInstance] = useState<{
    public_url: string;
    acme_email: string;
    wildcard_domain: string;
    smtp: {
      host: string;
      port: number;
      username: string;
      from: string;
      use_tls: boolean;
      configured: boolean;
    };
  } | null>(null);
  const [smtpPassword, setSmtpPassword] = useState("");
  const [testTo, setTestTo] = useState("");
  const [saveMsg, setSaveMsg] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [st, info, inst] = await Promise.all([
        getGitHubStatus(),
        getServerInfo().catch(() => null),
        getInstanceSettings().catch(() => null),
      ]);
      setStatus(st);
      setServerInfo(info);
      setInstance(inst);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load settings");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onConnect() {
    try {
      const { url } = await startGitHubAppLink();
      window.location.href = url;
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not start GitHub setup");
    }
  }

  async function onSaveInstance(e: React.FormEvent) {
    e.preventDefault();
    if (!instance) return;
    setSaveMsg(null);
    try {
      const updated = await patchInstanceSettings({
        public_url: instance.public_url,
        acme_email: instance.acme_email,
        wildcard_domain: instance.wildcard_domain,
        smtp: {
          host: instance.smtp.host,
          port: instance.smtp.port || 587,
          username: instance.smtp.username,
          from: instance.smtp.from,
          use_tls: instance.smtp.use_tls,
          password: smtpPassword || undefined,
        },
      });
      setInstance(updated);
      setSmtpPassword("");
      setSaveMsg("Settings saved.");
    } catch (e) {
      setSaveMsg(e instanceof Error ? e.message : "Could not save");
    }
  }

  async function onTestSMTP() {
    setSaveMsg(null);
    try {
      await testSMTP(testTo.trim());
      setSaveMsg("Test email sent.");
    } catch (e) {
      setSaveMsg(e instanceof Error ? e.message : "SMTP test failed");
    }
  }

  return (
    <DashboardShell title="Instance settings">
      <div className="mb-4 text-sm">
        <Link
          href="/settings/team"
          className="text-[var(--muted-foreground)] underline"
        >
          Team
        </Link>
        {" · "}
        <Link
          href="/settings/profile"
          className="text-[var(--muted-foreground)] underline"
        >
          Profile & security
        </Link>
      </div>
      <section
        aria-label="Dockyard version"
        className="mb-6 max-w-xl space-y-2 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6"
      >
        <h2 className="text-lg font-medium">Updates</h2>
        {serverInfo ? (
          <>
            <p className="text-sm text-[var(--muted-foreground)]">
              Running version{" "}
              <span className="font-mono text-[var(--foreground)]">
                {serverInfo.version}
              </span>
            </p>
            {serverInfo.update_available ? (
              <p className="text-sm">
                Update available:{" "}
                <span className="font-mono">{serverInfo.latest_tag}</span>. On
                the server run{" "}
                <code className="rounded bg-[var(--muted)]/40 px-1">
                  sudo dockyard update
                </code>
                .
              </p>
            ) : (
              <p className="text-sm text-[var(--muted-foreground)]">
                Up to date (or update check unavailable).
              </p>
            )}
          </>
        ) : loading ? (
          <p className="text-sm text-[var(--muted-foreground)]">
            Loading version…
          </p>
        ) : (
          <p className="text-sm text-[var(--muted-foreground)]">
            Could not load version info.
          </p>
        )}
      </section>
      <section
        aria-label="Panel URL and email"
        className="mb-6 max-w-xl space-y-4 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6"
      >
        <h2 className="text-lg font-medium">Panel &amp; SMTP</h2>
        <p className="text-sm text-[var(--muted-foreground)]">
          Public URL is used in password-reset and invite emails. SMTP password
          is stored encrypted on the server.
        </p>
        {saveMsg ? (
          <p className="text-sm text-green-700" role="status">
            {saveMsg}
          </p>
        ) : null}
        {instance ? (
          <form onSubmit={(e) => void onSaveInstance(e)} className="space-y-3">
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">Public URL</span>
              <input
                className="field-input mt-1 w-full"
                value={instance.public_url}
                onChange={(e) =>
                  setInstance({ ...instance, public_url: e.target.value })
                }
                placeholder="https://panel.example.com"
              />
            </label>
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">ACME email</span>
              <input
                className="field-input mt-1 w-full"
                type="email"
                value={instance.acme_email}
                onChange={(e) =>
                  setInstance({ ...instance, acme_email: e.target.value })
                }
              />
            </label>
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">
                Wildcard domain
              </span>
              <input
                className="field-input mt-1 w-full"
                value={instance.wildcard_domain}
                onChange={(e) =>
                  setInstance({ ...instance, wildcard_domain: e.target.value })
                }
                placeholder="*.apps.example.com"
              />
            </label>
            <hr className="border-[var(--border)]" />
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">SMTP host</span>
              <input
                className="field-input mt-1 w-full"
                value={instance.smtp.host}
                onChange={(e) =>
                  setInstance({
                    ...instance,
                    smtp: { ...instance.smtp, host: e.target.value },
                  })
                }
              />
            </label>
            <div className="grid grid-cols-2 gap-3">
              <label className="block text-sm">
                <span className="text-[var(--muted-foreground)]">Port</span>
                <input
                  className="field-input mt-1 w-full"
                  type="number"
                  value={instance.smtp.port || 587}
                  onChange={(e) =>
                    setInstance({
                      ...instance,
                      smtp: {
                        ...instance.smtp,
                        port: Number(e.target.value) || 587,
                      },
                    })
                  }
                />
              </label>
              <label className="flex items-end gap-2 pb-2 text-sm">
                <input
                  type="checkbox"
                  checked={instance.smtp.use_tls}
                  onChange={(e) =>
                    setInstance({
                      ...instance,
                      smtp: { ...instance.smtp, use_tls: e.target.checked },
                    })
                  }
                />
                TLS (port 465)
              </label>
            </div>
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">Username</span>
              <input
                className="field-input mt-1 w-full"
                value={instance.smtp.username}
                onChange={(e) =>
                  setInstance({
                    ...instance,
                    smtp: { ...instance.smtp, username: e.target.value },
                  })
                }
              />
            </label>
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">
                Password{" "}
                {instance.smtp.configured ? "(leave blank to keep)" : ""}
              </span>
              <input
                className="field-input mt-1 w-full"
                type="password"
                autoComplete="new-password"
                value={smtpPassword}
                onChange={(e) => setSmtpPassword(e.target.value)}
              />
            </label>
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">From</span>
              <input
                className="field-input mt-1 w-full"
                value={instance.smtp.from}
                onChange={(e) =>
                  setInstance({
                    ...instance,
                    smtp: { ...instance.smtp, from: e.target.value },
                  })
                }
                placeholder="Dockyard <noreply@example.com>"
              />
            </label>
            <button
              type="submit"
              className="rounded-lg bg-[var(--primary)] px-4 py-2 text-sm font-medium text-[var(--primary-foreground)]"
            >
              Save instance settings
            </button>
          </form>
        ) : loading ? (
          <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
        ) : (
          <p className="text-sm text-[var(--muted-foreground)]">
            Could not load instance settings (admin only).
          </p>
        )}
        {instance?.smtp.configured ? (
          <div className="flex flex-wrap items-end gap-2 border-t border-[var(--border)] pt-4">
            <label className="block flex-1 text-sm">
              <span className="text-[var(--muted-foreground)]">
                Send test email to
              </span>
              <input
                className="field-input mt-1 w-full"
                type="email"
                value={testTo}
                onChange={(e) => setTestTo(e.target.value)}
              />
            </label>
            <button
              type="button"
              onClick={() => void onTestSMTP()}
              className="rounded-lg border border-[var(--border)] px-4 py-2 text-sm"
            >
              Test SMTP
            </button>
          </div>
        ) : null}
      </section>
      <section
        aria-label="GitHub integration"
        className="max-w-xl space-y-4 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6"
      >
        <h2 className="text-lg font-medium">GitHub App</h2>
        <p className="text-sm text-[var(--muted-foreground)]">
          Connect a GitHub App to deploy from private repositories and receive
          push webhooks.
        </p>
        {linkedBanner ? (
          <p className="text-sm text-green-700" role="status">
            GitHub App linked successfully.
          </p>
        ) : null}
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
        {loading ? (
          <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
        ) : status?.linked ? (
          <dl className="space-y-2 text-sm">
            <div>
              <dt className="text-[var(--muted-foreground)]">App slug</dt>
              <dd className="font-mono">{status.slug ?? "—"}</dd>
            </div>
            {status.installation_id ? (
              <div>
                <dt className="text-[var(--muted-foreground)]">
                  Installation ID
                </dt>
                <dd className="font-mono">{status.installation_id}</dd>
              </div>
            ) : null}
            <p className="text-[var(--muted-foreground)]">
              Connected. Push deploys use this installation.
            </p>
          </dl>
        ) : (
          <button
            type="button"
            onClick={() => void onConnect()}
            className="rounded-lg bg-[var(--primary)] px-4 py-2 text-sm font-medium text-[var(--primary-foreground)]"
          >
            Install GitHub App
          </button>
        )}
      </section>
    </DashboardShell>
  );
}
