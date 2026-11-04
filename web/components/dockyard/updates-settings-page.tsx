"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  applyDockyardUpdate,
  getUpdatesStatus,
  type UpdatesStatus,
} from "@/lib/api";

export function UpdatesSettingsPage() {
  const [status, setStatus] = useState<UpdatesStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [checking, setChecking] = useState(false);
  const [applying, setApplying] = useState(false);
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
      const data = await getUpdatesStatus();
      setStatus(data);
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
    setMessage(null);
    setError(null);
    try {
      const res = await applyDockyardUpdate();
      setMessage(res.message);
      if (res.restarting) {
        setMessage(
          `${res.message} Reload this page in a moment if the panel disconnects.`,
        );
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Update failed");
    } finally {
      setApplying(false);
    }
  }

  return (
    <DashboardShell title="Updates" breadcrumb="Settings">
      <section
        aria-label="Dockyard version and updates"
        className="max-w-xl space-y-4 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6"
      >
        <h2 className="text-lg font-medium">Dockyard version</h2>
        <p className="text-sm text-[var(--muted-foreground)]">
          Checks GitHub Releases when you open this page. Applying an update
          replaces the host binary and restarts the Dockyard service; your apps
          and proxy keep running.
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
            Checking for updates…
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
          <button
            type="button"
            disabled={loading || checking || applying}
            onClick={() => void load(false)}
            className="rounded-lg border border-[var(--border)] px-4 py-2 text-sm disabled:opacity-50"
          >
            {checking ? "Checking…" : "Check again"}
          </button>
          {status?.update_available ? (
            <button
              type="button"
              disabled={applying || !status.apply_supported || !status.check_ok}
              onClick={() => void onApply()}
              className="rounded-lg bg-[var(--primary)] px-4 py-2 text-sm font-medium text-[var(--primary-foreground)] disabled:opacity-50"
            >
              {applying ? "Updating…" : "Update now"}
            </button>
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
            on a Linux server, or develop locally without applying from the UI.
          </p>
        ) : null}
      </section>
    </DashboardShell>
  );
}
