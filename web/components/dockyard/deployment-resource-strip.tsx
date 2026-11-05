import type { Deployment } from "@/lib/api";
import {
  deploymentResourceStripRows,
  hasDeploymentResourceMetrics,
} from "@/lib/build-resilience";

type DeploymentResourceStripProps = {
  deployment: Deployment;
  className?: string;
};

export function DeploymentResourceStrip({
  deployment,
  className = "",
}: DeploymentResourceStripProps) {
  if (!hasDeploymentResourceMetrics(deployment)) {
    return null;
  }
  const rows = deploymentResourceStripRows(deployment);
  return (
    <section
      className={`rounded-[16px] border px-4 py-3 ${className}`}
      style={{
        borderColor: "var(--border)",
        background: "var(--surface-muted, var(--surface))",
      }}
      aria-label="Build resource usage"
    >
      <p
        className="font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        Build resources
      </p>
      <dl className="mt-2 grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
        {rows.map((row) => (
          <div key={row.label}>
            <dt className="text-[12px]" style={{ color: "var(--muted-text)" }}>
              {row.label}
            </dt>
            <dd className="text-sm font-medium tabular-nums">{row.value}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
