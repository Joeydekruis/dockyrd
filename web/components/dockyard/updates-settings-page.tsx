"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import {
  UpdateApplyingOverlay,
  type UpdateApplyPhase,
} from "@/components/dockyard/update-applying-overlay";
import {
  applyDockyardUpdate,
  getUpdatesStatus,
  type UpdatesStatus,
} from "@/lib/api";
import { waitForPanelHealth } from "@/lib/wait-for-panel-health";

export function UpdatesSettingsPage() {
  const [status, setStatus] = useState<UpdatesStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [checking, setChecking] = useState(false);
  const [applying, setApplying] = useState(false);
  const [updatePhase, setUpdatePhase] = useState<UpdateApplyPhase | null>(null);
  const [updateDetail, setUpdateDetail] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async (initial: boolean) => {
    if (initial) {
      setLoading(true);
    } else {
      setChecking(true);
    }
    setError(null);
    try {
      const data = await getUpdatesStatus({ refresh: !initial });
      setStatus(data);
      if (data.check_error && !initial) {
        setError(data.check_error);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not check for updates");
      setStatus(null);
    } finally {
      setLoading(false);
      setChecking(false);
    }
  }, []);

  useEffect(() => {
    void load(true);
  }, [load]);

  async function onApply() {
    if (!status?.update_available) return;
    setApplying(true);
    setUpdatePhase("applying");
    setUpdateDetail(null);
    setMessage(null);
    setError(null);
    try {
      const res = await applyDockyardUpdate();
      if (res.restarting) {
        setUpdatePhase("restarting");
        setUpdateDetail(res.message);
        await waitForPanelHealth();
        setUpdatePhase("complete");
        setUpdateDetail(
          res.installed_tag
            ? `Dockyard ${res.installed_tag} is running.`
            : res.message,
        );
        await new Promise((resolve) => setTimeout(resolve, 900));
        window.location.reload();
        return;
      }
      setUpdatePhase("complete");
      setUpdateDetail(res.message);
      setMessage(res.message);
      await new Promise((resolve) => setTimeout(resolve, 600));
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Update failed";
      setUpdatePhase("error");
      setUpdateDetail(msg);
      setError(msg);
      await new Promise((resolve) => setTimeout(resolve, 1200));
    } finally {
      setApplying(false);
      setUpdatePhase(null);
      setUpdateDetail(null);
    }
  }

  return (
    <DashboardShell title="Updates" breadcrumb="Server settings">
      <UpdateApplyingOverlay
        open={updatePhase !== null}
        phase={updatePhase ?? "applying"}
        detail={updateDetail}
      />
      <SettingsPageBody>
        <section
          aria-label="Dockyard version and updates"
          className="space-y-4 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6"
        >
          <h2 className="text-lg font-medium">Dockyard version</h2>
          <p className="text-sm text-[var(--muted-foreground)]">
            Dockyard checks GitHub every 5 minutes in the background. This page
            shows the cached result; use Check again to refresh manually (up to
            40 times per hour without a GitHub token). Applying an update
            replaces the host binary and restarts the Dockyard service; your
            apps and proxy keep running.
          </p>

          {error ? (
            <p className="text-sm text-red-600" role="alert">
              {error}
            </p>
          ) : null}
          {message ? (
            <p className="text-sm text-green-700" role="status">
              {message}
            </p>
          ) : null}

          {loading ? (
            <p className="text-sm text-[var(--muted-foreground)]">
              Loading update status…
            </p>
          ) : status ? (
            <dl className="space-y-3 text-sm">
              <div>
                <dt className="text-[var(--muted-foreground)]">
                  Running version
                </dt>
                <dd className="font-mono">{status.version}</dd>
              </div>
              {!status.check_ok ? (
                <p className="text-[var(--muted-foreground)]">
                  Update check failed
                  {status.check_error ? `: ${status.check_error}` : "."}
                </p>
              ) : status.update_available ? (
                <>
                  <div>
                    <dt className="text-[var(--muted-foreground)]">
                      Latest release
                    </dt>
                    <dd className="font-mono">{status.latest_tag}</dd>
                  </div>
                  {status.release_notes ? (
                    <div>
                      <dt className="mb-1 text-[var(--muted-foreground)]">
                        Release notes
                      </dt>
                      <dd className="max-h-48 overflow-y-auto rounded-lg border border-[var(--border)] bg-[var(--background)] p-3 whitespace-pre-wrap">
                        {status.release_notes}
                      </dd>
                    </div>
                  ) : null}
                </>
              ) : (
                <p className="text-[var(--muted-foreground)]">
                  Dockyard is up to date.
                </p>
              )}
            </dl>
          ) : (
            <p className="text-sm text-[var(--muted-foreground)]">
              Could not load update status (admin only).
            </p>
          )}

          <div className="flex flex-wrap gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              disabled={loading || checking || applying}
              onClick={() => void load(false)}
            >
              {checking ? "Checking…" : "Check again"}
            </Button>
            {status?.update_available ? (
              <Button
                type="button"
                disabled={
                  applying || !status.apply_supported || !status.check_ok
                }
                onClick={() => void onApply()}
              >
                {applying ? "Updating…" : "Update now"}
              </Button>
            ) : null}
          </div>

          {status &&
          !status.apply_supported &&
          status.apply_unsupported_reason ? (
            <p className="text-xs text-[var(--muted-foreground)]">
              {status.apply_unsupported_reason} Use{" "}
              <code className="rounded bg-[var(--muted)]/40 px-1">
                dockyard update
              </code>{" "}
              on a Linux server, or develop locally without applying from the
              UI.
            </p>
          ) : null}
        </section>
      </SettingsPageBody>
    </DashboardShell>
  );
}
