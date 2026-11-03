"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { type ApplicationRow, listApplications } from "@/lib/api";

export function ApplicationsPage() {
  const [rows, setRows] = useState<ApplicationRow[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const { applications } = await listApplications();
      setRows(applications);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load applications");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return (
    <DashboardShell title="Applications">
      <section className="space-y-3">
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
        {loading ? (
          <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
        ) : rows.length === 0 ? (
          <p className="text-sm text-[var(--muted-foreground)]">
            No applications yet. Open a project to create an app service.
          </p>
        ) : (
          <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
            {rows.map((app) => (
              <li
                key={app.id}
                className="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-sm"
              >
                <div>
                  <p className="font-medium">{app.name}</p>
                  <p className="text-xs text-[var(--muted-foreground)]">
                    {app.project_name}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-[var(--muted-foreground)]">
                    {app.desired_state}
                  </span>
                  <Link
                    href={`/projects/${app.project_slug}`}
                    className="text-xs text-[var(--primary)] underline"
                  >
                    Open project
                  </Link>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </DashboardShell>
  );
}
