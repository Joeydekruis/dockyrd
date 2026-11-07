"use client";

import { useCallback, useEffect, useState } from "react";
import { AddConnectionModal } from "@/components/dockyrd/add-connection-modal";
import { Button } from "@/components/dockyrd/ui/button";
import { EmptyState } from "@/components/dockyrd/ui/empty-state";
import { ResourceRow } from "@/components/dockyrd/ui/resource-row";
import {
  deleteServiceLink,
  listServiceLinks,
  type ServiceLink,
} from "@/lib/api";
import { resourceStatusDisplay } from "@/lib/resource-dashboard";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type ServiceConnectedApplicationsPanelProps = {
  providerServiceId: string;
  onChanged?: () => void;
};

export function ServiceConnectedApplicationsPanel({
  providerServiceId,
  onChanged,
}: ServiceConnectedApplicationsPanelProps) {
  const [links, setLinks] = useState<ServiceLink[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listServiceLinks(providerServiceId);
      setLinks(res.incoming);
    } finally {
      setLoading(false);
    }
  }, [providerServiceId]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onRemove(link: ServiceLink) {
    if (!window.confirm(`Disconnect ${link.peer.name} from this resource?`)) {
      return;
    }
    await deleteServiceLink(link.id, true);
    await refresh();
    onChanged?.();
  }

  return (
    <section className={RESOURCE_DETAIL_CARD} style={resourceDetailCardStyle()}>
      <div
        className="flex flex-wrap items-center justify-between gap-3 border-b pb-4"
        style={{ borderColor: "var(--border)" }}
      >
        <h2
          className="font-mono text-[11px] font-medium uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          Connected applications
        </h2>
        <Button
          type="button"
          variant="outline"
          onClick={() => setModalOpen(true)}
        >
          Connect application
        </Button>
      </div>
      {loading ? (
        <p className="mt-4 text-[13px]" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : links.length === 0 ? (
        <EmptyState
          title="No applications connected"
          description="Connect an app so it can use this resource over the internal network."
        />
      ) : (
        <ul className="mt-2">
          {links.map((link) => {
            const st = resourceStatusDisplay(
              link.peer.status === "healthy" ? "running" : link.peer.status,
              "application",
            );
            return (
              <li key={link.id}>
                <ResourceRow
                  name={link.peer.name}
                  href={`/applications/${link.peer.slug}`}
                  meta="application"
                  statusLabel={st.label}
                  statusVariant={st.variant}
                  trailing={
                    <Button
                      type="button"
                      variant="ghost"
                      onClick={() => void onRemove(link)}
                    >
                      Remove
                    </Button>
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
        serviceId={providerServiceId}
        providerMode
        onCreated={() => {
          void refresh();
          onChanged?.();
        }}
      />
    </section>
  );
}
