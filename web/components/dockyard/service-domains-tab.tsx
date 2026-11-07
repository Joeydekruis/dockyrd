"use client";

import { Plus } from "lucide-react";
import { type CSSProperties, useEffect, useState } from "react";
import { AddCustomDomainModal } from "@/components/dockyard/add-custom-domain-modal";
import { AppsDomainNudge } from "@/components/dockyard/apps-domain-nudge";
import { DomainDnsRecordsModal } from "@/components/dockyard/domain-dns-records-modal";
import { Button } from "@/components/dockyard/ui/button";
import { ConfirmDialog } from "@/components/dockyard/ui/confirm-dialog";
import {
  PopoverMenu,
  type PopoverMenuItem,
} from "@/components/dockyard/ui/popover-menu";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  createServiceDomain,
  deleteServiceDomain,
  getServerInfo,
  patchServiceDomain,
  type ServiceDomain,
} from "@/lib/api";
import { domainStatusDisplay } from "@/lib/domain-dns";

type ServiceDomainsTabProps = {
  serviceId: string;
  domains: ServiceDomain[];
  onRefresh: () => Promise<void>;
  onError: (message: string | null) => void;
  layout?: "static" | "app" | "compose";
  composeTarget?: { service: string; port: number };
  addDomainPlaceholder?: string;
  cardClassName?: string;
  cardStyle?: CSSProperties;
};

const tableHead =
  "grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-3 px-4 py-2 font-mono text-[11px] uppercase tracking-widest";
const tableRow =
  "grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-3 border-t px-4 py-3";

function DomainTableRow({
  domain,
  busy,
  onShowDns,
  onDelete,
  onMakePrimary,
}: {
  domain: ServiceDomain;
  busy: boolean;
  onShowDns: (d: ServiceDomain) => void;
  onDelete: (d: ServiceDomain) => void;
  onMakePrimary: (d: ServiceDomain) => void;
}) {
  const st = domainStatusDisplay(domain);
  const menuItems: PopoverMenuItem[] = [
    {
      id: "primary",
      label: "Make primary domain",
      disabled: domain.is_primary || st.needsVerification,
      onSelect: () => onMakePrimary(domain),
    },
    ...(domain.is_platform
      ? []
      : [
          {
            id: "dns",
            label: "Show DNS records",
            onSelect: () => onShowDns(domain),
          } satisfies PopoverMenuItem,
        ]),
    {
      id: "delete",
      label: "Delete",
      destructive: true,
      dividerBefore: true,
      onSelect: () => onDelete(domain),
    },
  ];

  return (
    <div className={tableRow} style={{ borderColor: "var(--border)" }}>
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <span className="truncate font-mono text-sm">{domain.hostname}</span>
        {domain.is_primary ? (
          <span
            className="rounded-full border px-2 py-0.5 text-[11px] font-medium"
            style={{ borderColor: "var(--border)", color: "var(--muted-text)" }}
          >
            Primary
          </span>
        ) : null}
      </div>
      <div className="flex flex-wrap items-center justify-end gap-2">
        <StatusBadge label={st.label} variant={st.variant} />
        {st.needsVerification ? (
          <Button
            type="button"
            variant="outline"
            className="text-[13px] px-3 py-1.5 min-h-0"
            disabled={busy}
            onClick={() => onShowDns(domain)}
          >
            Show DNS records
          </Button>
        ) : null}
      </div>
      <div className="flex justify-end">
        <PopoverMenu
          items={menuItems}
          label={`Actions for ${domain.hostname}`}
        />
      </div>
    </div>
  );
}

function DomainSection({
  title,
  description,
  action,
  domains,
  emptyMessage,
  busy,
  onShowDns,
  onDelete,
  onMakePrimary,
}: {
  title: string;
  description: string;
  action?: React.ReactNode;
  domains: ServiceDomain[];
  emptyMessage?: React.ReactNode;
  busy: boolean;
  onShowDns: (d: ServiceDomain) => void;
  onDelete: (d: ServiceDomain) => void;
  onMakePrimary: (d: ServiceDomain) => void;
}) {
  return (
    <section>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            {description}
          </p>
        </div>
        {action}
      </div>
      {domains.length === 0 ? (
        (emptyMessage ?? null)
      ) : (
        <div
          className="mt-5 overflow-hidden rounded-xl border"
          style={{ borderColor: "var(--border)" }}
        >
          <div
            className={tableHead}
            style={{
              background: "var(--hover-row)",
              color: "var(--muted-text)",
            }}
          >
            <span>Domain</span>
            <span className="text-right">Status</span>
            <span className="sr-only">Actions</span>
          </div>
          {domains.map((d) => (
            <DomainTableRow
              key={d.id}
              domain={d}
              busy={busy}
              onShowDns={onShowDns}
              onDelete={onDelete}
              onMakePrimary={onMakePrimary}
            />
          ))}
        </div>
      )}
    </section>
  );
}

export function ServiceDomainsTab({
  serviceId,
  domains,
  onRefresh,
  onError,
  layout = "app",
  composeTarget,
  cardClassName = "rounded-[20px] border px-6 py-5",
  cardStyle,
}: ServiceDomainsTabProps) {
  const [busy, setBusy] = useState(false);
  const [publicIPv4, setPublicIPv4] = useState("");
  const [addOpen, setAddOpen] = useState(false);
  const [dnsDomain, setDnsDomain] = useState<ServiceDomain | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ServiceDomain | null>(null);

  const customDomains = domains.filter((d) => !d.is_platform);
  const platformDomains = domains.filter((d) => d.is_platform);
  const hasPrimary = domains.some((d) => d.is_primary);

  useEffect(() => {
    void getServerInfo()
      .then((info) => setPublicIPv4(info.public_ipv4 ?? ""))
      .catch(() => setPublicIPv4(""));
  }, []);

  const surfaceStyle = {
    borderColor: "var(--border)",
    background: "var(--surface)",
    ...cardStyle,
  };

  async function run<T>(fn: () => Promise<T>): Promise<T | undefined> {
    setBusy(true);
    onError(null);
    try {
      return await fn();
    } catch (e) {
      onError(e instanceof Error ? e.message : "Something went wrong");
      return undefined;
    } finally {
      setBusy(false);
    }
  }

  async function handleDeleteConfirm() {
    const domain = deleteTarget;
    if (!domain) return;
    setDeleteTarget(null);
    await run(async () => {
      await deleteServiceDomain(domain.id);
      await onRefresh();
    });
  }

  async function handleMakePrimary(domain: ServiceDomain) {
    await run(async () => {
      await patchServiceDomain(domain.id, { primary: true });
      await onRefresh();
    });
  }

  async function handleAddCustom(hostname: string) {
    const success = await run(async () => {
      await createServiceDomain(serviceId, {
        hostname,
        primary: !hasPrimary,
        ...(layout === "compose" && composeTarget
          ? {
              target_service: composeTarget.service,
              target_port: composeTarget.port,
            }
          : {}),
      });
      await onRefresh();
      return true;
    });
    if (success) setAddOpen(false);
  }

  async function ensurePlatformDomain() {
    await run(async () => {
      await createServiceDomain(serviceId, {
        platform: true,
        primary: !hasPrimary,
        ...(layout === "compose" && composeTarget
          ? {
              target_service: composeTarget.service,
              target_port: composeTarget.port,
            }
          : {}),
      });
      await onRefresh();
    });
  }

  const modals = (
    <>
      <AddCustomDomainModal
        open={addOpen}
        busy={busy}
        onClose={() => setAddOpen(false)}
        onSubmit={handleAddCustom}
      />
      <DomainDnsRecordsModal
        open={dnsDomain !== null}
        domain={dnsDomain}
        publicIPv4={publicIPv4}
        onClose={() => setDnsDomain(null)}
        onRefresh={() => void onRefresh()}
      />
      <ConfirmDialog
        open={deleteTarget !== null}
        title="Delete domain"
        description={
          deleteTarget
            ? `Remove ${deleteTarget.hostname} from this service? Traffic on that hostname will stop.`
            : undefined
        }
        confirmLabel="Delete"
        destructive
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => void handleDeleteConfirm()}
      />
    </>
  );

  const addCustomButton = (
    <Button
      type="button"
      iconLeft={Plus}
      disabled={busy}
      onClick={() => setAddOpen(true)}
    >
      Add custom domain
    </Button>
  );

  const addPlatformButton =
    platformDomains.length > 0 ? (
      <Button
        type="button"
        variant="outline"
        disabled={busy}
        onClick={() => void ensurePlatformDomain()}
      >
        Add default URL
      </Button>
    ) : null;

  const platformEmpty = (
    <div className="mt-5 space-y-3">
      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
        No default URL yet. HTTPS routing needs at least one platform hostname.
      </p>
      <Button
        type="button"
        variant="outline"
        disabled={busy}
        onClick={() => void ensurePlatformDomain()}
      >
        Add default URL
      </Button>
    </div>
  );

  const composeLead =
    layout === "compose"
      ? composeTarget
        ? `HTTPS routes to compose service “${composeTarget.service}” on port ${composeTarget.port}.`
        : "Set target service and port before adding domains."
      : "Dockyard-generated hostname for this service. Uses your apps domain when verified, otherwise sslip.io.";

  return (
    <div className="space-y-6">
      {modals}
      <AppsDomainNudge scope="service" />
      <div className={cardClassName} style={surfaceStyle}>
        <DomainSection
          title="Default URL"
          description={composeLead}
          action={addPlatformButton}
          domains={platformDomains}
          busy={busy}
          onShowDns={setDnsDomain}
          onDelete={setDeleteTarget}
          onMakePrimary={handleMakePrimary}
          emptyMessage={platformEmpty}
        />
      </div>
      <div className={cardClassName} style={surfaceStyle}>
        <DomainSection
          title="Your domains"
          description="Hostnames you own, verified with DNS and served over HTTPS."
          action={addCustomButton}
          domains={customDomains}
          busy={busy}
          onShowDns={setDnsDomain}
          onDelete={setDeleteTarget}
          onMakePrimary={handleMakePrimary}
        />
      </div>
    </div>
  );
}
