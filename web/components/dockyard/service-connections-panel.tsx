"use client";

import { useCallback, useEffect, useState } from "react";
import { AddConnectionModal } from "@/components/dockyard/add-connection-modal";
import { Button } from "@/components/dockyard/ui/button";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import {
  deleteServiceLink,
  listServiceLinks,
  type Project,
  type ServiceLink,
} from "@/lib/api";
import { resourceStatusDisplay } from "@/lib/resource-dashboard";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type ServiceConnectionsPanelProps = {
  serviceId: string;
  project: Project;
  onChanged?: () => void;
};

function peerHref(kind: string, slug: string): string {
  if (kind === "postgres" || kind === "redis") return `/databases/${slug}`;
  if (kind === "bucket") return `/object-storage/${slug}`;
  if (kind === "static_site") return `/static-sites/${slug}`;
  return `/applications/${slug}`;
}

export function ServiceConnectionsPanel({
  serviceId,
  project,
  onChanged,
}: ServiceConnectionsPanelProps) {
  const [links, setLinks] = useState<ServiceLink[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listServiceLinks(serviceId);
      setLinks(res.outgoing);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load connections");
    } finally {
      setLoading(false);
    }
  }, [serviceId]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onRemove(link: ServiceLink) {
    if (
      !window.confirm(
        `Remove connection to ${link.peer.name}? Linked environment variables can be removed.`,
      )
    ) {
      return;
    }
    await deleteServiceLink(link.id, true);
    await refresh();
    onChanged?.();
  }

  const enforced = project.links_mode === "strict";

  return (
    <section className={RESOURCE_DETAIL_CARD} style={resourceDetailCardStyle()}>
      <div
        className="flex flex-wrap items-center justify-between gap-3 border-b pb-4"
        style={{ borderColor: "var(--border)" }}
      >
        <div>
          <h2
            className="font-mono text-[11px] font-medium uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Connections
          </h2>
          <p
            className="mt-1 text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            Only linked services are reachable for references.{" "}
            <span
              className="ml-1 rounded-full border px-2 py-0.5 text-[11px] font-mono uppercase"
              style={{ borderColor: "var(--border)" }}
            >
              {enforced ? "Enforced" : "Not enforced"}
            </span>
          </p>
        </div>
        <Button
          type="button"
          variant="secondary"
          onClick={() => setModalOpen(true)}
        >
          Add connection
        </Button>
      </div>

      {error ? (
        <p className="mt-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : loading ? (
        <p className="mt-4 text-[13px]" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : links.length === 0 ? (
        <EmptyState
          title="No connections yet"
          description="Link a database, cache, bucket, or private app so this application can reach it."
        />
      ) : (
        <ul className="mt-2">
          {links.map((link) => {
            const st = resourceStatusDisplay(
              link.peer.status === "healthy" ? "running" : link.peer.status,
              link.peer.kind === "postgres" || link.peer.kind === "redis"
                ? "database"
                : "application",
            );
            return (
              <li key={link.id}>
                <ResourceRow
                  name={link.peer.name}
                  href={peerHref(link.peer.kind, link.peer.slug)}
                  meta={`${link.peer.kind} · ${link.peer.slug}`}
                  statusLabel={st.label}
                  statusVariant={st.variant}
                  trailing={
                    <div className="flex items-center gap-2">
                      {link.variables_count > 0 ? (
                        <span
                          className="text-[12px] font-mono"
                          style={{ color: "var(--muted-text)" }}
                        >
                          {link.variables_count} vars
                        </span>
                      ) : null}
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() => void onRemove(link)}
                      >
                        Remove
                      </Button>
                    </div>
                  }
                />
              </li>
            );
          })}
        </ul>
      )}

      <AddConnectionModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        serviceId={serviceId}
        onCreated={() => {
          void refresh();
          onChanged?.();
        }}
      />
    </section>
  );
}
