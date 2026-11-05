"use client";

import { Check, Circle } from "lucide-react";
import { Alert } from "@/components/dockyard/ui/alert";
import { Button } from "@/components/dockyard/ui/button";
import { DnsRecordBlock } from "@/components/dockyard/ui/dns-record-block";
import { Modal } from "@/components/dockyard/ui/modal";
import type { ServiceDomain } from "@/lib/api";
import { checkServiceDomain } from "@/lib/api";
import { aRecordForHostname, domainStatusDisplay } from "@/lib/domain-dns";

type DomainDnsRecordsModalProps = {
  open: boolean;
  domain: ServiceDomain | null;
  publicIPv4: string;
  onClose: () => void;
  onRefresh: () => void;
};

function StepMarker({ done, active }: { done?: boolean; active?: boolean }) {
  if (done) {
    return (
      <span
        className="flex size-6 shrink-0 items-center justify-center rounded-full"
        style={{
          background: "var(--status-ok-soft)",
          color: "var(--status-ok)",
        }}
      >
        <Check className="size-3.5" strokeWidth={2.5} aria-hidden />
      </span>
    );
  }
  return (
    <span
      className="flex size-6 shrink-0 items-center justify-center rounded-full border border-dashed"
      style={{
        borderColor: active ? "var(--ink-navy)" : "var(--border)",
        color: "var(--muted-text)",
      }}
    >
      <Circle className="size-2 fill-current" aria-hidden />
    </span>
  );
}

export function DomainDnsRecordsModal({
  open,
  domain,
  publicIPv4,
  onClose,
  onRefresh,
}: DomainDnsRecordsModalProps) {
  if (!domain) return null;

  const d = domain;
  const ip =
    publicIPv4.trim() ||
    "Set DOCKYARD_PUBLIC_IPV4 on the server (Instance → General)";
  const { name, nameAlt } = aRecordForHostname(d.hostname);
  const statusView = domainStatusDisplay(d);
  const verified = !d.is_platform && !statusView.needsVerification;

  async function handleRefreshStatus() {
    if (!d.is_platform) {
      await checkServiceDomain(d.id);
    }
    await onRefresh();
  }

  if (!domain) {
    return null;
  }

  return (
    <Modal
      open={open}
      title="DNS records"
      description="Add these records where your domain's DNS is managed."
      onClose={onClose}
      size="lg"
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Close
          </Button>
          <Button type="button" onClick={() => void handleRefreshStatus()}>
            Refresh status
          </Button>
        </>
      }
    >
      <p className="mb-6 text-sm" style={{ color: "var(--muted-text)" }}>
        Status:{" "}
        <span className="font-medium" style={{ color: "var(--ink-navy)" }}>
          {statusView.label}
        </span>
      </p>

      <ol className="space-y-8">
        <li className="flex gap-4">
          <StepMarker done />
          <div className="min-w-0 flex-1">
            <p className="font-medium">Add domain</p>
            <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
              Added{" "}
              <span className="font-mono" style={{ color: "var(--ink-navy)" }}>
                {domain.hostname}
              </span>
            </p>
          </div>
        </li>

        <li className="flex gap-4">
          <StepMarker active={!verified} done={verified} />
          <div className="min-w-0 flex-1 space-y-3">
            <p className="font-medium">Point domain</p>
            <Alert
              tone="info"
              description="Remove conflicting A/AAAA records for this host at your DNS provider. If you use Cloudflare, turn off the orange proxy (DNS only) until certificates issue."
            />
            <DnsRecordBlock
              type="A"
              name={name}
              nameAlt={nameAlt}
              value={ip}
              copyable
            />
            <p className="text-[13px]" style={{ color: "var(--muted-text)" }}>
              Once HTTP traffic reaches this server on ports 80 and 443,
              Dockyard requests a Let&apos;s Encrypt certificate automatically.
            </p>
          </div>
        </li>

        <li className="flex gap-4 opacity-80">
          <StepMarker done={verified} />
          <div className="min-w-0 flex-1">
            <p className="font-medium">HTTPS</p>
            <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
              {verified
                ? "Certificate active — this domain is served over HTTPS."
                : "Available after DNS points at this server and verification succeeds."}
            </p>
          </div>
        </li>
      </ol>
    </Modal>
  );
}
