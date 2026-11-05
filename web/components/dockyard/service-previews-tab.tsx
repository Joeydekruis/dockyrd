"use client";

import { ExternalLink, Trash2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  deleteServicePreview,
  listServicePreviews,
  type ServicePreview,
} from "@/lib/api";
import { formatTimeAgo } from "@/lib/resource-dashboard";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type Props = {
  serviceId: string;
  onError: (message: string | null) => void;
};

export function ServicePreviewsTab({ serviceId, onError }: Props) {
  const [rows, setRows] = useState<ServicePreview[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const { previews } = await listServicePreviews(serviceId);
      setRows(previews);
      onError(null);
    } catch (e) {
      onError(e instanceof Error ? e.message : "Failed to load previews");
    } finally {
      setLoading(false);
    }
  }, [serviceId, onError]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onDelete(preview: ServicePreview) {
    setBusyId(preview.id);
    try {
      await deleteServicePreview(serviceId, preview.id);
      await refresh();
    } catch (e) {
      onError(e instanceof Error ? e.message : "Delete failed");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <section className={RESOURCE_DETAIL_CARD} style={resourceDetailCardStyle()}>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">
            Preview deployments
          </h2>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            One environment per open pull request against your deploy branch.
          </p>
        </div>
        <Button type="button" variant="outline" onClick={() => void refresh()}>
          Refresh
        </Button>
      </div>

      {loading ? (
        <p className="mt-6 text-sm" style={{ color: "var(--muted-text)" }}>
          Loading previews…
        </p>
      ) : rows.length === 0 ? (
        <div className="mt-4">
          <EmptyState
            title="No preview deployments"
            description="Enable previews in Settings and open a pull request on GitHub."
          />
        </div>
      ) : (
        <ul className="mt-4 divide-y" style={{ borderColor: "var(--border)" }}>
          {rows.map((row) => {
            const href = row.primary_hostname
              ? `https://${row.primary_hostname}`
              : undefined;
            const when = row.preview_last_push_at
              ? formatTimeAgo(row.preview_last_push_at)
              : "—";
            return (
              <li
                key={row.id}
                className="flex flex-wrap items-center gap-3 py-4 first:pt-0"
              >
                <StatusBadge label={`PR #${row.preview_pr}`} variant="info" />
                <span className="min-w-0 flex-1 truncate text-[13px] font-medium">
                  {row.name}
                </span>
                {row.preview_head_sha ? (
                  <span
                    className="font-mono text-[12px] tabular-nums"
                    style={{ color: "var(--muted-text)" }}
                  >
                    {row.preview_head_sha.slice(0, 7)}
                  </span>
                ) : null}
                <span
                  className="text-[12px] tabular-nums"
                  style={{ color: "var(--muted-text)" }}
                >
                  {when}
                </span>
                {href ? (
                  <a
                    href={href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-[13px] underline"
                    style={{ color: "var(--link-blue)" }}
                  >
                    Open
                    <ExternalLink className="size-3.5" strokeWidth={1.8} />
                  </a>
                ) : null}
                <Button
                  type="button"
                  variant="outline"
                  disabled={busyId === row.id}
                  aria-label={`Delete preview for PR ${row.preview_pr}`}
                  onClick={() => void onDelete(row)}
                >
                  <Trash2 className="size-4" strokeWidth={1.8} />
                </Button>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
