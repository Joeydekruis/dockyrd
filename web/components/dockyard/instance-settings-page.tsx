"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { getGitHubStatus, getServerInfo, startGitHubAppLink } from "@/lib/api";

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

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [st, info] = await Promise.all([
        getGitHubStatus(),
        getServerInfo().catch(() => null),
      ]);
      setStatus(st);
      setServerInfo(info);
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
