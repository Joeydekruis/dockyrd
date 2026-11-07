"use client";

import type { LucideIcon } from "lucide-react";
import { Box, Database, Globe, HardDrive, Layers } from "lucide-react";
import Link from "next/link";
import { CommitChip } from "@/components/dockyrd/ui/commit-chip";
import { StatusBadge } from "@/components/dockyrd/ui/status-badge";
import type { Deployment, Service } from "@/lib/api";
import {
  resourceKindMeta,
  servicePublicHost,
} from "@/lib/project-resource-catalog";
import { formatTimeAgo, resourceStatusDisplay } from "@/lib/resource-dashboard";
import { serviceDetailHref } from "@/lib/service-routes";

function kindIcon(kind: string): LucideIcon {
  switch (kind) {
    case "static_site":
      return Globe;
    case "postgres":
    case "redis":
      return Database;
    case "bucket":
      return HardDrive;
    case "compose":
      return Layers;
    default:
      return Box;
  }
}

function formatCpu(pct?: number): string {
  if (pct === undefined || Number.isNaN(pct)) return "—";
  return `${pct.toFixed(1)}%`;
}

function formatMemory(bytes?: number): string {
  if (!bytes) return "—";
  if (bytes >= 1_073_741_824) {
    return `${(bytes / 1_073_741_824).toFixed(1)} GB`;
  }
  return `${Math.round(bytes / 1_048_576)} MB`;
}

export type ProjectResourceRow = {
  service: Service;
  deployment?: Deployment;
  cpuPercent?: number;
  memoryBytes?: number;
  outgoingLinkCount?: number;
};

type ProjectResourceTableProps = {
  rows: ProjectResourceRow[];
};

export function ProjectResourceTable({ rows }: ProjectResourceTableProps) {
  return (
    <div
      className="overflow-x-auto rounded-[20px] border"
      style={{ borderColor: "var(--border)", background: "var(--surface)" }}
    >
      <table className="w-full min-w-[880px] border-collapse text-left text-sm">
        <thead>
          <tr
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            <th className="px-6 py-4 font-medium">Resource</th>
            <th className="px-4 py-4 font-medium">Status</th>
            <th className="px-4 py-4 font-medium">Links</th>
            <th className="px-4 py-4 font-medium">Endpoint</th>
            <th className="px-4 py-4 font-medium">CPU</th>
            <th className="px-4 py-4 font-medium">Memory</th>
            <th className="px-6 py-4 font-medium">Last activity</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(
            ({
              service,
              deployment,
              cpuPercent,
              memoryBytes,
              outgoingLinkCount,
            }) => {
              const meta = resourceKindMeta(service.kind);
              const status = resourceStatusDisplay(
                service.desired_state,
                meta.surface,
              );
              const Icon = kindIcon(service.kind);
              const href = serviceDetailHref(service.kind, service.slug);
              const host = servicePublicHost(service);
              const dep = deployment;
              return (
                <tr
                  key={service.id}
                  className="border-t"
                  style={{ borderColor: "var(--border)" }}
                >
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div
                        className="flex size-10 shrink-0 items-center justify-center rounded-xl"
                        style={{
                          background: "var(--brand-soft)",
                          color: "var(--link-blue)",
                        }}
                      >
                        <Icon
                          strokeWidth={1.8}
                          className="size-5"
                          aria-hidden
                        />
                      </div>
                      <div>
                        <Link
                          href={href}
                          className="font-semibold hover:underline"
                          style={{ color: "var(--ink-navy)" }}
                        >
                          {service.name}
                        </Link>
                        <p
                          className="mt-0.5 font-mono text-[11px] uppercase tracking-wide"
                          style={{ color: "var(--muted-text)" }}
                        >
                          {meta.tag}
                        </p>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-4">
                    <StatusBadge
                      label={status.label}
                      variant={status.variant}
                    />
                  </td>
                  <td
                    className="px-4 py-4 font-mono text-[12px]"
                    style={{ color: "var(--muted-text)" }}
                  >
                    {service.kind === "app" && (outgoingLinkCount ?? 0) > 0
                      ? `${outgoingLinkCount} link${outgoingLinkCount === 1 ? "" : "s"}`
                      : "—"}
                  </td>
                  <td className="max-w-[200px] truncate px-4 py-4">
                    {host ? (
                      <a
                        href={
                          host.startsWith("http") ? host : `https://${host}`
                        }
                        target="_blank"
                        rel="noopener noreferrer"
                        className="font-medium hover:underline"
                        style={{ color: "var(--link-blue)" }}
                      >
                        {host}
                      </a>
                    ) : (
                      <span style={{ color: "var(--muted-text)" }}>
                        Internal only
                      </span>
                    )}
                  </td>
                  <td
                    className="px-4 py-4 font-mono text-[13px]"
                    style={{ color: "var(--secondary-text)" }}
                  >
                    {formatCpu(cpuPercent)}
                  </td>
                  <td
                    className="px-4 py-4 font-mono text-[13px]"
                    style={{ color: "var(--secondary-text)" }}
                  >
                    {formatMemory(memoryBytes)}
                  </td>
                  <td className="px-6 py-4">
                    {dep?.commit_sha ? (
                      <CommitChip
                        sha={dep.commit_sha}
                        message={
                          dep.branch
                            ? `on ${dep.branch}`
                            : dep.trigger || undefined
                        }
                        timeAgo={formatTimeAgo(
                          dep.finished_at ?? dep.started_at,
                        )}
                      />
                    ) : (
                      <span
                        className="text-[13px]"
                        style={{ color: "var(--muted-text)" }}
                      >
                        No deploys yet
                      </span>
                    )}
                  </td>
                </tr>
              );
            },
          )}
        </tbody>
      </table>
    </div>
  );
}
