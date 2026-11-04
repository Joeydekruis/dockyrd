"use client";

import { HardDrive, Plus } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CreateBucketModal } from "@/components/dockyard/create-bucket-modal";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  type ResourceProvisionContext,
  ResourceProvisioningModal,
} from "@/components/dockyard/resource-provisioning-modal";
import { Button } from "@/components/dockyard/ui/button";
import { HeroPromoCard } from "@/components/dockyard/ui/hero-promo-card";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { ResourceListAddLink } from "@/components/dockyard/ui/resource-list-add-link";
import { ResourceListSection } from "@/components/dockyard/ui/resource-list-section";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import { StatCard } from "@/components/dockyard/ui/stat-card";
import { TemplateCard } from "@/components/dockyard/ui/template-card";
import {
  getStorageStatus,
  installGarage,
  listProjects,
  type Project,
  type StorageStatus,
} from "@/lib/api";
import {
  formatResourceCount,
  projectPathLabel,
} from "@/lib/resource-dashboard";
import { filterByQuery } from "@/lib/service-display";

export function ObjectStoragePage() {
  const [status, setStatus] = useState<StorageStatus | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [provisionOpen, setProvisionOpen] = useState(false);
  const [provisionContext, setProvisionContext] =
    useState<ResourceProvisionContext | null>(null);

  const refresh = useCallback(async (opts?: { silent?: boolean }) => {
    if (!opts?.silent) setLoading(true);
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

  const onProvisionFinished = useCallback(() => {
    void refresh({ silent: true });
  }, [refresh]);

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
        String((b.config as { bucket_name?: string }).bucket_name ?? b.slug),
      ]),
    [buckets, query],
  );

  const dash = loading && buckets.length === 0 && !status ? "—" : undefined;
  const bucketCount = buckets.length;
  const showInitialLoading = loading && !status;
  const showGarageSetup =
    !showInitialLoading && bucketCount === 0 && !status?.installed;

  async function onInstall() {
    try {
      await installGarage();
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Install failed");
    }
  }

  function openCreate() {
    if (!status?.installed) {
      setError("Install Garage before creating buckets.");
      return;
    }
    setCreateOpen(true);
  }

  return (
    <DashboardShell title="Object storage" breadcrumb="Projects">
      <PageHeader
        eyebrow="All projects"
        title="Object storage"
        subtitle="S3-compatible buckets running on your server."
        action={
          <Button
            type="button"
            iconLeft={Plus}
            onClick={openCreate}
            disabled={!status?.installed}
          >
            Create bucket
          </Button>
        }
      />

      {error ? (
        <p className="mb-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      <section
        aria-label="Summary"
        className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4"
      >
        <StatCard
          label="Buckets"
          value={dash ?? bucketCount}
          hint={
            bucketCount === 0 && !showInitialLoading ? "none yet" : undefined
          }
        />
        <StatCard
          label="Storage used"
          value={dash ?? "0 B"}
          hint="of 214 GB free"
        />
        <StatCard
          label="Objects"
          value={dash ?? 0}
          hint={
            bucketCount === 0 && !showInitialLoading ? "no data" : undefined
          }
        />
        <StatCard label="Requests" value={dash ?? "—"} hint="last 24 hours" />
      </section>

      {showInitialLoading ? (
        <p className="mt-10 text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : showGarageSetup ? (
        <div className="mt-10 space-y-6">
          <HeroPromoCard
            badge="S3-compatible"
            title="No buckets yet"
            description="Store uploads, backups and assets on your own disk. Pick a bundled S3 server and Dockyard runs it for you."
          />
          <section
            aria-label="Storage servers"
            className="grid gap-4 md:grid-cols-2"
          >
            <TemplateCard
              name="Garage"
              description="Lightweight, built for small self-hosted clusters. Recommended for Dockyard."
              recommended
              actionLabel="Install Garage"
              onSelect={() => void onInstall()}
            />
            <TemplateCard
              name="MinIO"
              description="Full-featured S3 server with a familiar console."
              popular
              actionLabel="Install MinIO"
              actionDisabled
            />
          </section>
        </div>
      ) : bucketCount === 0 ? (
        <ResourceListSection
          title="All buckets"
          countLabel={formatResourceCount(0, "bucket")}
          footer={
            <ResourceListAddLink label="Create bucket" onClick={openCreate} />
          }
        >
          <p
            className="py-8 text-center text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            No buckets yet. Create one for uploads, backups, or static assets.
          </p>
        </ResourceListSection>
      ) : (
        <ResourceListSection
          title="All buckets"
          countLabel={formatResourceCount(bucketCount, "bucket")}
          filterValue={query}
          onFilterChange={setQuery}
          footer={
            <ResourceListAddLink label="Create bucket" onClick={openCreate} />
          }
        >
          {filtered.length === 0 ? (
            <p
              className="py-6 text-center text-sm"
              style={{ color: "var(--muted-text)" }}
            >
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
                  meta={projectPathLabel(b.project_name)}
                  subMeta={`S3 · ${bucketNameValue}`}
                  statusLabel="Ready"
                  statusVariant="ok"
                  icon={HardDrive}
                  internalOnly
                />
              );
            })
          )}
        </ResourceListSection>
      )}

      <CreateBucketModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={(ctx) => {
          setProvisionContext(ctx);
          setProvisionOpen(true);
        }}
      />

      <ResourceProvisioningModal
        open={provisionOpen}
        context={provisionContext}
        onClose={() => setProvisionOpen(false)}
        onFinished={onProvisionFinished}
      />
    </DashboardShell>
  );
}
