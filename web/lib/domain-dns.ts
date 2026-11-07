import type { StatusVariant } from "@/components/dockyard/ui/status-badge";
import type { ServiceDomain } from "@/lib/api";

export type DomainStatusView = {
  label: string;
  variant: StatusVariant;
  needsVerification: boolean;
};

/** User-facing domain status (Sevalla-style labels). */
export function domainStatusDisplay(domain: ServiceDomain): DomainStatusView {
  const dns = (domain.dns_status ?? "").trim().toLowerCase();
  if (domain.is_platform || dns === "active" || dns === "ok") {
    return { label: "Active", variant: "ok", needsVerification: false };
  }
  if (dns === "wrong_target") {
    return { label: "Wrong target", variant: "err", needsVerification: true };
  }
  if (dns === "pending" || dns === "") {
    return { label: "Verification", variant: "warn", needsVerification: true };
  }
  return {
    label: domain.dns_status ?? "Unknown",
    variant: "idle",
    needsVerification: false,
  };
}

/** Suggested A record for pointing a custom hostname at dockyrd. */
export function aRecordForHostname(hostname: string): {
  name: string;
  nameAlt?: string;
} {
  const host = hostname.trim().toLowerCase().replace(/\.$/, "");
  const parts = host.split(".").filter(Boolean);
  if (parts.length <= 2) {
    return { name: "@", nameAlt: host };
  }
  const label = parts[0];
  return { name: label, nameAlt: `${label}.${parts.slice(1).join(".")}` };
}

export { writeClipboardText as copyText } from "@/lib/clipboard";
