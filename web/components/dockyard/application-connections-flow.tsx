"use client";

import {
  ArrowRight,
  Globe,
  Link2,
  Monitor,
  Server,
  Unlink,
} from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";
import { useCallback, useEffect, useState } from "react";
import { AddConnectionModal } from "@/components/dockyrd/add-connection-modal";
import { RemoveConnectionModal } from "@/components/dockyrd/remove-connection-modal";
import { Button } from "@/components/dockyrd/ui/button";
import { EmptyState } from "@/components/dockyrd/ui/empty-state";
import { StatusBadge } from "@/components/dockyrd/ui/status-badge";
import {
  deleteServiceLink,
  listServiceLinks,
  type Project,
  type Service,
  type ServiceDomain,
  type ServiceLink,
} from "@/lib/api";
import { resourceStatusDisplay } from "@/lib/resource-dashboard";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type ApplicationConnectionsFlowProps = {
  service: Service;
  project: Project;
  domains: ServiceDomain[];
  port: number;
  onChanged?: () => void;
  onGoDomains?: () => void;
  onGoLogs?: () => void;
};

function peerHref(kind: string, slug: string): string {
  if (kind === "postgres" || kind === "redis") return `/databases/${slug}`;
  if (kind === "bucket") return `/object-storage/${slug}`;
  if (kind === "static_site") return `/static-sites/${slug}`;
  return `/applications/${slug}`;
}

function FlowCard({
  title,
  subtitle,
  children,
  className = "",
}: {
  title: string;
  subtitle?: string;
  children?: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={`rounded-2xl border px-4 py-3 ${className}`}
      style={{
        borderColor: "var(--border)",
        background: "var(--surface)",
      }}
    >
      <p
        className="font-mono text-[10px] font-medium uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {title}
      </p>
      {subtitle ? (
        <p
          className="mt-1 text-[13px] font-medium"
          style={{ color: "var(--ink-navy)" }}
        >
          {subtitle}
        </p>
      ) : null}
      {children}
    </div>
  );
}

export function ApplicationConnectionsFlow({
  service,
  project,
  domains,
  port,
  onChanged,
  onGoDomains,
  onGoLogs,
}: ApplicationConnectionsFlowProps) {
  const [links, setLinks] = useState<ServiceLink[]>([]);
  const [loading, setLoading] = useState(true);
  const [addOpen, setAddOpen] = useState(false);
  const [removeLink, setRemoveLink] = useState<ServiceLink | null>(null);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listServiceLinks(service.id);
      setLinks(res.outgoing);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load connections");
    } finally {
      setLoading(false);
    }
  }, [service.id]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const primary = domains.find((d) => d.is_primary) ?? domains[0];
  const appStatus = resourceStatusDisplay(service.desired_state, "application");
  const enforced = project.links_mode === "strict";
  const mem =
    typeof service.config.memLimitMB === "number"
      ? service.config.memLimitMB
      : 512;

  async function confirmRemove(removeVars: boolean) {
    if (!removeLink) return;
    await deleteServiceLink(removeLink.id, removeVars);
    setRemoveLink(null);
    await refresh();
    onChanged?.();
  }

  return (
    <section className={RESOURCE_DETAIL_CARD} style={resourceDetailCardStyle()}>
      <div
        className="flex flex-wrap items-center justify-between gap-3 border-b pb-4"
        style={{ borderColor: "var(--border)" }}
      >
        <div>
          <h2 className="text-lg font-semibold tracking-tight">Connections</h2>
          <p
            className="mt-1 flex flex-wrap items-center gap-2 text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            Traffic and linked services for this application.
            <span
              className="rounded-full border px-2 py-0.5 font-mono text-[10px] uppercase tracking-wide"
              style={{ borderColor: "var(--border)" }}
            >
              {enforced ? "Enforced" : "Advisory"}
            </span>
          </p>
        </div>
        <Button
          type="button"
          variant="outline"
          onClick={() => setAddOpen(true)}
        >
          Link resource
        </Button>
      </div>

      {error ? (
        <p className="mt-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      <div
        className="relative mt-6 overflow-hidden rounded-2xl border p-6"
        style={{
          borderColor: "var(--border)",
          background:
            "repeating-linear-gradient(0deg, transparent, transparent 11px, var(--border) 11px, var(--border) 12px), repeating-linear-gradient(90deg, transparent, transparent 11px, var(--border) 11px, var(--border) 12px)",
          backgroundSize: "12px 12px",
          backgroundColor: "var(--hover-row)",
        }}
      >
        <div className="relative grid gap-4 lg:grid-cols-[1fr_auto_1.2fr_auto_1fr] lg:items-center">
          <FlowCard
            title="Public access"
            subtitle={primary?.hostname ?? "No domain yet"}
          >
            <p
              className="mt-2 text-[12px]"
              style={{ color: "var(--muted-text)" }}
            >
              {primary?.cert_status?.toLowerCase() === "active" ||
              primary?.cert_status?.toLowerCase() === "ready"
                ? "HTTPS active"
                : "HTTPS when certificate is ready"}
            </p>
            {onGoDomains ? (
              <Button
                type="button"
                variant="outline"
                className="mt-3 w-full"
                onClick={onGoDomains}
              >
                Domain settings
              </Button>
            ) : null}
          </FlowCard>

          <div className="hidden items-center justify-center lg:flex">
            <ArrowRight
              className="size-5"
              style={{ color: "var(--muted-text)" }}
              aria-hidden
            />
          </div>

          <div className="flex flex-col items-center gap-2">
            <span
              className="rounded-full border px-3 py-1 font-mono text-[12px]"
              style={{
                borderColor: "var(--border)",
                background: "var(--surface)",
              }}
            >
              :{port}
            </span>
            <FlowCard
              title="Web service"
              subtitle={service.name}
              className="w-full min-w-[200px]"
            >
              <div className="mt-2 flex flex-wrap items-center gap-2">
                <StatusBadge
                  label={appStatus.label}
                  variant={appStatus.variant}
                />
                <Monitor
                  className="size-4"
                  style={{ color: "var(--muted-text)" }}
                  aria-hidden
                />
              </div>
              <p
                className="mt-2 font-mono text-[11px]"
                style={{ color: "var(--muted-text)" }}
              >
                {mem} MB · port {port}
              </p>
              {onGoLogs ? (
                <Button
                  type="button"
                  variant="outline"
                  className="mt-3 w-full"
                  onClick={onGoLogs}
                >
                  View logs
                </Button>
              ) : null}
            </FlowCard>
          </div>

          <div className="hidden items-center justify-center lg:flex">
            <Link2
              className="size-5"
              style={{ color: "var(--muted-text)" }}
              aria-hidden
            />
          </div>

          <div>
            <p
              className="mb-2 font-mono text-[10px] font-medium uppercase tracking-widest"
              style={{ color: "var(--muted-text)" }}
            >
              Connected services
            </p>
            {loading ? (
              <p className="text-[13px]" style={{ color: "var(--muted-text)" }}>
                Loading…
              </p>
            ) : links.length === 0 ? (
              <div
                className="rounded-2xl border border-dashed px-4 py-6 text-center"
                style={{
                  borderColor: "var(--border)",
                  background: "var(--surface)",
                }}
              >
                <Globe
                  className="mx-auto size-6"
                  style={{ color: "var(--muted-text)" }}
                  aria-hidden
                />
                <p className="mt-2 text-sm font-medium">No linked services</p>
                <p
                  className="mt-1 text-[13px]"
                  style={{ color: "var(--muted-text)" }}
                >
                  Link a database, cache, bucket, or private app so this
                  application can reach it on the internal network.
                </p>
              </div>
            ) : (
              <ul className="space-y-2">
                {links.map((link) => {
                  const st = resourceStatusDisplay(
                    link.peer.status === "healthy"
                      ? "running"
                      : link.peer.status,
                    link.peer.kind === "postgres" || link.peer.kind === "redis"
                      ? "database"
                      : "application",
                  );
                  return (
                    <li
                      key={link.id}
                      className="flex items-center gap-2 rounded-xl border px-3 py-2"
                      style={{
                        borderColor: "var(--border)",
                        background: "var(--surface)",
                      }}
                    >
                      <Server
                        className="size-4 shrink-0"
                        style={{ color: "var(--link-blue)" }}
                        aria-hidden
                      />
                      <div className="min-w-0 flex-1">
                        <Link
                          href={peerHref(link.peer.kind, link.peer.slug)}
                          className="truncate text-[13px] font-medium hover:underline"
                          style={{ color: "var(--ink-navy)" }}
                        >
                          {link.peer.name}
                        </Link>
                        <p
                          className="font-mono text-[11px]"
                          style={{ color: "var(--muted-text)" }}
                        >
                          {link.peer.kind} · {link.variables_count} vars
                        </p>
                      </div>
                      <StatusBadge label={st.label} variant={st.variant} />
                      <Button
                        type="button"
                        variant="outline"
                        aria-label={`Remove connection to ${link.peer.name}`}
                        onClick={() => setRemoveLink(link)}
                      >
                        <Unlink className="size-4" aria-hidden />
                      </Button>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>
        </div>
      </div>

      {!loading && links.length === 0 ? (
        <div className="mt-4">
          <EmptyState
            title="No connections yet"
            description="Link resources so environment references resolve and traffic can reach private services."
          />
        </div>
      ) : null}

      <AddConnectionModal
        open={addOpen}
        onOpenChange={setAddOpen}
        serviceId={service.id}
        onCreated={() => {
          void refresh();
          onChanged?.();
        }}
      />

      <RemoveConnectionModal
        open={removeLink !== null}
        onOpenChange={(open) => {
          if (!open) setRemoveLink(null);
        }}
        peerName={removeLink?.peer.name ?? ""}
        variablesCount={removeLink?.variables_count ?? 0}
        onConfirm={confirmRemove}
      />
    </section>
  );
}
