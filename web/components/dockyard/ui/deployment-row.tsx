import { CommitChip } from "@/components/dockyard/ui/commit-chip";
import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyard/ui/status-badge";

type DeploymentRowProps = {
  name: string;
  statusLabel: string;
  statusVariant?: StatusVariant;
  sha: string;
  commitMessage?: string;
  meta?: string;
  href?: string;
};

export function DeploymentRow({
  name,
  statusLabel,
  statusVariant = "ok",
  sha,
  commitMessage,
  meta,
  href = "#",
}: DeploymentRowProps) {
  return (
    <div
      className="flex flex-wrap items-center gap-3 border-t py-4 first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <div className="min-w-0 flex-1">
        <a href={href} className="font-semibold hover:underline">
          {name}
        </a>
        {meta ? (
          <p
            className="mt-0.5 font-mono text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            {meta}
          </p>
        ) : null}
      </div>
      <StatusBadge label={statusLabel} variant={statusVariant} />
      <CommitChip sha={sha} message={commitMessage} />
    </div>
  );
}
