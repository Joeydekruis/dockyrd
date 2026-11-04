"use client";

import { useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { getGitHubStatus, startGitHubAppLink } from "@/lib/api";

export function IntegrationsSettingsPage() {
  const search = useSearchParams();
  const linkedBanner = search.get("github") === "linked";
  const [status, setStatus] = useState<{
    linked: boolean;
    slug?: string;
    installation_id?: number;
  } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const st = await getGitHubStatus();
      setStatus(st);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load integrations");
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
    <DashboardShell title="Integrations" breadcrumb="Settings">
      <SettingsPageBody>
        <section
          aria-label="GitHub integration"
          className="space-y-4 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6"
        >
          <h2 className="text-lg font-medium">GitHub</h2>
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
            <Button type="button" onClick={() => void onConnect()}>
              Install GitHub App
            </Button>
          )}
        </section>
      </SettingsPageBody>
    </DashboardShell>
  );
}
