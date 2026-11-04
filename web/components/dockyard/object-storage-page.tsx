"use client";

import { HardDrive } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { AddCard } from "@/components/dockyard/ui/add-card";
import { Chip } from "@/components/dockyard/ui/chip";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { ResourceListSection } from "@/components/dockyard/ui/resource-list-section";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import { SelectField } from "@/components/dockyard/ui/select-field";
import { StatCard } from "@/components/dockyard/ui/stat-card";
import { TemplateCard } from "@/components/dockyard/ui/template-card";
import {
  createService,
  getStorageStatus,
  installGarage,
  listProjects,
  type Project,
  type StorageStatus,
} from "@/lib/api";
import { filterByQuery } from "@/lib/service-display";

export function ObjectStoragePage() {
  const [status, setStatus] = useState<StorageStatus | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [bucketName, setBucketName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);

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

  const buckets = status?.buckets ?? [];

  const filtered = useMemo(
    () =>
      filterByQuery(buckets, query, (b) => [
        b.name,
        b.slug,
        b.project_name,
        String(
          (b.config as { bucket_name?: string }).bucket_name ?? b.slug,
        ),
      ]),
    [buckets, query],
  );

  const projectCount = new Set(buckets.map((b) => b.project_id)).size;

  const countLabel =
    filtered.length === buckets.length
      ? `${buckets.length} total`
      : `${filtered.length} of ${buckets.length}`;

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
      setShowCreate(false);
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create bucket failed");
    }
  }

  const bucketForm =
    status?.installed && showCreate ? (
      <form
        onSubmit={onCreateBucket}
        className="flex flex-col gap-3 rounded-[20px] border p-4 sm:flex-row sm:items-end"
        style={{ borderColor: "var(--border)", background: "var(--surface)" }}
      >
        <SelectField
          label="Project"
          placeholder="Select…"
          value={projectId}
          onChange={setProjectId}
          options={projects.map((p) => ({ value: p.id, label: p.name }))}
        />
        <label className="flex flex-1 flex-col gap-1 text-sm">
          <span
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Bucket name
          </span>
          <input
            className="field-input"
            value={bucketName}
            onChange={(ev) => setBucketName(ev.target.value)}
            placeholder="uploads"
          />
        </label>
        <div className="flex flex-wrap gap-2">
          <Button type="submit">Create bucket</Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => setShowCreate(false)}
          >
            Cancel
          </Button>
        </div>
      </form>
    ) : null;

  return (
    <DashboardShell title="Object storage">
      <PageHeader
        title="Object storage"
        subtitle="S3-compatible buckets on this server via Garage."
      />

      {error ? (
        <p className="mb-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      <section aria-label="Summary" className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Buckets" value={loading ? "—" : buckets.length} />
        <StatCard label="Projects" value={loading ? "—" : projectCount} />
        <StatCard
          label="Garage"
          value={loading ? "—" : status?.installed ? "Installed" : "Not installed"}
          hint={
            status?.installed
              ? status.version ?? "Self-hosted S3"
              : "Install to create buckets"
          }
        />
        <StatCard
          label="Endpoint"
          value={loading ? "—" : status?.installed ? "Internal" : "—"}
          hint={
            status?.installed && status.internal_endpoint
              ? status.internal_endpoint
              : undefined
          }
        />
      </section>

      {!loading && !status?.installed ? (
        <section aria-label="Storage servers" className="mt-10">
          <h2 className="mb-4 text-xl font-semibold tracking-[-0.03em]">
            Storage servers
          </h2>
          <div className="grid gap-4 md:grid-cols-2">
            <TemplateCard
              name="Garage"
              description="Lightweight, self-hosted S3-compatible storage. Recommended for Dockyard."
              stack="S3 API"
              recommended
              onSelect={() => void onInstall()}
            />
            <article
              className="rounded-[20px] border p-5 opacity-60"
              style={{ borderColor: "var(--border)", background: "var(--surface)" }}
            >
              <Chip>Not in MVP</Chip>
              <p className="mt-3 font-semibold">MinIO</p>
              <p className="mt-2 text-[13px]" style={{ color: "var(--muted-text)" }}>
                Alternative S3 server — not offered in the first release.
              </p>
            </article>
          </div>
        </section>
      ) : null}

      {loading ? (
        <p className="mt-10 text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : !status?.installed ? (
        buckets.length === 0 ? (
          <div className="mt-10">
            <EmptyState
              title="No buckets yet"
              description="Install Garage on this server, then create project-scoped buckets for uploads and static assets."
            />
          </div>
        ) : null
      ) : buckets.length === 0 ? (
        <div className="mt-10 space-y-4">
          <EmptyState
            title="No buckets yet"
            description="Create a bucket and attach it to a project for app uploads or static assets."
          />
          {!showCreate ? (
            <AddCard
              title="Create bucket"
              description="Name a bucket under a project."
              onClick={() => setShowCreate(true)}
            />
          ) : (
            bucketForm
          )}
        </div>
      ) : (
        <ResourceListSection
          title="All buckets"
          countLabel={countLabel}
          filterValue={query}
          onFilterChange={setQuery}
          footer={
            showCreate ? (
              bucketForm
            ) : (
              <AddCard
                title="Create bucket"
                description="Add another S3 bucket."
                onClick={() => setShowCreate(true)}
              />
            )
          }
        >
          {filtered.length === 0 ? (
            <p className="py-6 text-center text-sm" style={{ color: "var(--muted-text)" }}>
              No buckets match your filter.
            </p>
          ) : (
            filtered.map((b) => {
              const bucketNameValue = String(
                (b.config as { bucket_name?: string }).bucket_name ?? b.slug,
              );
              const projectSlug =
                projects.find((p) => p.id === b.project_id)?.slug ??
                b.project_id;
              return (
                <ResourceRow
                  key={b.id}
                  name={b.name}
                  href={`/projects/${projectSlug}`}
                  meta={`${b.project_name} · ${bucketNameValue}`}
                  subMeta="S3 · Garage"
                  statusLabel="Ready"
                  statusVariant="ok"
                  icon={HardDrive}
                />
              );
            })
          )}
        </ResourceListSection>
      )}
    </DashboardShell>
  );
}
