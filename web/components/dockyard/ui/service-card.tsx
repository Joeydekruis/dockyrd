import type { LucideIcon } from "lucide-react";
import { Box } from "lucide-react";
import { IconTile } from "@/components/dockyard/ui/icon-tile";
import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyard/ui/status-badge";

type ServiceCardProps = {
  name: string;
  domain?: string;
  statusLabel: string;
  statusVariant?: StatusVariant;
  icon?: LucideIcon;
  href?: string;
};

export function ServiceCard({
  name,
  domain,
  statusLabel,
  statusVariant = "ok",
  icon,
  href = "#",
}: ServiceCardProps) {
  return (
    <a
      href={href}
      className="flex items-start gap-4 rounded-[20px] border p-5 transition-colors hover:bg-[var(--hover-row)]"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <IconTile icon={icon ?? Box} label={name} />
      <div className="min-w-0 flex-1">
        <p className="font-semibold">{name}</p>
        {domain ? (
          <p
            className="mt-1 font-mono text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            {domain}
          </p>
        ) : null}
        <div className="mt-3">
          <StatusBadge label={statusLabel} variant={statusVariant} />
        </div>
      </div>
    </a>
  );
}
