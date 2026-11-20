"use client";

import { ChevronRight } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Modal } from "@/components/dockyard/ui/modal";
import { formatBytes } from "@/lib/analytics-stats";
import {
  getServerDiskUsageDetail,
  type DiskUsageDetail,
  type DiskUsageDetailEntry,
} from "@/lib/api";

export type DiskDrillTarget =
  | { scope: "data"; path: string; title: string }
  | { scope: "docker"; path: string; title: string }
  | { scope: "service"; serviceId: string; title: string };

type Props = {
  open: boolean;
  target: DiskDrillTarget | null;
  onClose: () => void;
};

function detailParams(t: DiskDrillTarget) {
  if (t.scope === "service") {
    return { scope: t.scope, serviceId: t.serviceId };
  }
  return { scope: t.scope, path: t.path };
}

export function DiskUsageDetailModal({ open, target, onClose }: Props) {
  const [detail, setDetail] = useState<DiskUsageDetail | null>(null);
  const [stack, setStack] = useState<DiskDrillTarget[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async (t: DiskDrillTarget) => {
    setLoading(true);
    setError(null);
    try {
      const data = await getServerDiskUsageDetail(detailParams(t));
      setDetail(data);
    } catch (e) {
      setDetail(null);
      setError(e instanceof Error ? e.message : "Failed to load details");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!open || !target) {
      setDetail(null);
      setStack([]);
      setError(null);
      return;
    }
    setStack([target]);
    void load(target);
  }, [open, target, load]);

  const drillInto = (entry: DiskUsageDetailEntry) => {
    if (detail?.scope !== "data" || !entry.drillable || !entry.path) return;
    const next: DiskDrillTarget = {
      scope: "data",
      path: entry.path,
      title: entry.name,
    };
    setStack((s) => [...s, next]);
    void load(next);
  };

  const popTo = (index: number) => {
    const t = stack[index];
    if (!t) return;
    setStack(stack.slice(0, index + 1));
    void load(t);
  };

  const active = stack.at(-1) ?? target;

  return (
    <Modal
      open={open}
      onClose={onClose}
      size="lg"
      title={detail?.title ?? active?.title ?? "Disk usage"}
      description={
        detail?.scope === "data" && detail.path
          ? `Relative path: ${detail.path}`
          : detail?.note
      }
    >
      {stack.length > 1 ? (
        <nav
          className="mb-4 flex flex-wrap items-center gap-1 text-xs"
          style={{ color: "var(--muted-text)" }}
          aria-label="Breadcrumb"
        >
          {stack.map((crumb, i) => (
            <span
              key={`${crumb.scope}-${"path" in crumb ? crumb.path : crumb.serviceId}-${i}`}
              className="inline-flex items-center gap-1"
            >
              {i > 0 ? <ChevronRight className="size-3" /> : null}
              <button
                type="button"
                className={
                  i === stack.length - 1
                    ? "font-medium text-[var(--text)]"
                    : "underline-offset-2 hover:underline"
                }
                onClick={() => popTo(i)}
              >
                {crumb.title}
              </button>
            </span>
          ))}
        </nav>
      ) : null}

      {loading ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Scanning…
        </p>
      ) : null}
      {error ? (
        <p className="text-sm text-[var(--danger)]">{error}</p>
      ) : null}

      {detail && !loading ? (
        <>
          <p className="mb-4 text-sm">
            Total here:{" "}
            <span className="font-mono font-medium">
              {formatBytes(detail.bytes)}
            </span>
          </p>
          {detail.children && detail.children.length > 0 ? (
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-text)]">
                  <th className="pb-2 pr-3">Name</th>
                  <th className="pb-2">Size</th>
                </tr>
              </thead>
              <tbody>
                {detail.children.map((row) => (
                  <tr
                    key={`${row.name}-${row.path ?? ""}-${row.bytes}`}
                    className="border-t border-[var(--border)]"
                  >
                    <td className="py-2 pr-3">
                      {row.drillable && detail.scope === "data" ? (
                        <button
                          type="button"
                          className="text-left font-medium underline-offset-2 hover:underline"
                          onClick={() => drillInto(row)}
                        >
                          {row.name}
                        </button>
                      ) : (
                        <span>{row.name}</span>
                      )}
                      {row.hint ? (
                        <p
                          className="text-xs"
                          style={{ color: "var(--muted-text)" }}
                        >
                          {row.hint}
                        </p>
                      ) : null}
                    </td>
                    <td className="py-2 font-mono tabular-nums">
                      {formatBytes(row.bytes)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <p className="text-sm" style={{ color: "var(--muted-text)" }}>
              Nothing else to list at this level.
            </p>
          )}
        </>
      ) : null}
    </Modal>
  );
}
