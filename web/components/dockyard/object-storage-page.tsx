"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { SelectField } from "@/components/dockyard/ui/select-field";
import {
  createService,
  getStorageStatus,
  installGarage,
  listProjects,
  type Project,
  type StorageStatus,
} from "@/lib/api";

export function ObjectStoragePage() {
  const [status, setStatus] = useState<StorageStatus | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [bucketName, setBucketName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [st, { projects: ps }] = await Promise.all([
        getStorageStatus(),
        listProjects(),
      ]);
      setStatus(st);
      setProjects(ps);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load storage");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onInstall() {
    try {
      await installGarage();
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Install failed");
    }
  }

  async function onCreateBucket(e: React.FormEvent) {
    e.preventDefault();
    if (!projectId || !bucketName.trim()) return;
    try {
      await createService({
        project_id: projectId,
        kind: "bucket",
        name: bucketName.trim(),
        config: {
          bucket_name: bucketName.trim().toLowerCase().replace(/\s+/g, "-"),
        },
      });
      setBucketName("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create bucket failed");
    }
  }

  return (
    <DashboardShell title="Object storage">
      <section className="space-y-6">
        <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4">
          <h2 className="text-sm font-semibold">Garage (recommended)</h2>
          <p className="mt-1 text-sm text-[var(--muted-foreground)]">
            Self-hosted S3-compatible storage on this server. MinIO is not
            offered in MVP.
          </p>
          <p className="mt-2 font-mono text-xs text-[var(--muted-foreground)]">
            {status?.installed
              ? `Installed · ${status.internal_endpoint}`
              : "Not installed"}
          </p>
          {!status?.installed ? (
            <Button
              type="button"
              className="mt-3"
              onClick={() => void onInstall()}
            >
              Install Garage
            </Button>
          ) : null}
        </div>

        {status?.installed ? (
          <form
            onSubmit={onCreateBucket}
            className="flex flex-col gap-3 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4 sm:flex-row sm:items-end"
          >
            <SelectField
              label="Project"
              placeholder="Select…"
              value={projectId}
              onChange={setProjectId}
              options={projects.map((p) => ({ value: p.id, label: p.name }))}
            />
            <label className="flex flex-1 flex-col gap-1 text-sm">
              <span className="text-[var(--muted-foreground)]">
                Bucket name
              </span>
              <input
                className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
                value={bucketName}
                onChange={(ev) => setBucketName(ev.target.value)}
                placeholder="uploads"
              />
            </label>
            <Button type="submit">Create bucket</Button>
          </form>
        ) : null}

        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}

        {loading ? (
          <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
        ) : (status?.buckets?.length ?? 0) === 0 ? (
          <p className="text-sm text-[var(--muted-foreground)]">
            No buckets yet.
          </p>
        ) : (
          <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
            {status?.buckets?.map((b) => (
              <li key={b.id} className="px-4 py-3 text-sm">
                <p className="font-medium">{b.name}</p>
                <p className="text-[var(--muted-foreground)]">
                  {b.project_name} ·{" "}
                  {String(
                    (b.config as { bucket_name?: string }).bucket_name ??
                      b.slug,
                  )}
                </p>
              </li>
            ))}
          </ul>
        )}
      </section>
    </DashboardShell>
  );
}
