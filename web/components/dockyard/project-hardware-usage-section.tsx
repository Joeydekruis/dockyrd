"use client";

import Link from "next/link";
import { ChartPanel } from "@/components/dockyrd/ui/chart-panel";
import {
  type TimeRange,
  TimeRangeToggle,
} from "@/components/dockyrd/ui/time-range-toggle";

type ProjectHardwareUsageSectionProps = {
  projectName: string;
  serverHostname?: string;
  range: TimeRange;
  onRangeChange: (range: TimeRange) => void;
  cpuSummary?: string;
  memorySummary?: string;
  diskSummary?: string;
  networkSummary?: string;
};

export function ProjectHardwareUsageSection({
  projectName,
  serverHostname,
  range,
  onRangeChange,
  cpuSummary = "—",
  memorySummary = "—",
  diskSummary = "—",
  networkSummary = "—",
}: ProjectHardwareUsageSectionProps) {
  const host = serverHostname ?? "this server";
  return (
    <section className="mt-10" aria-label="Hardware usage">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 className="text-xl font-semibold tracking-tight">
            Hardware usage
          </h2>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            What all resources of {projectName} use together on {host}.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <TimeRangeToggle value={range} onChange={onRangeChange} />
          <Link
            href="/analytics"
            className="text-sm font-medium hover:underline"
            style={{ color: "var(--link-blue)" }}
          >
            Server analytics →
          </Link>
        </div>
      </div>
      <div className="grid gap-4 lg:grid-cols-2">
        <article
          className="rounded-[20px] border p-5 lg:col-span-1"
          style={{
            background: "var(--surface)",
            borderColor: "var(--border)",
          }}
        >
          <p
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            CPU
          </p>
          <p className="mt-2 text-3xl font-semibold tracking-tight">
            {cpuSummary}
          </p>
          <div className="mt-2 [&_article]:border-0 [&_article]:p-0">
            <ChartPanel title="CPU trend" height={100} />
          </div>
        </article>
        <article
          className="rounded-[20px] border p-5 lg:col-span-1"
          style={{
            background: "var(--surface)",
            borderColor: "var(--border)",
          }}
        >
          <p
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Memory
          </p>
          <p className="mt-2 text-3xl font-semibold tracking-tight">
            {memorySummary}
          </p>
          <div className="mt-2 [&_article]:border-0 [&_article]:p-0">
            <ChartPanel title="Memory trend" height={100} />
          </div>
        </article>
        <article
          className="rounded-[20px] border p-5"
          style={{
            background: "var(--surface)",
            borderColor: "var(--border)",
          }}
        >
          <p
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Disk
          </p>
          <p className="mt-2 text-3xl font-semibold tracking-tight">
            {diskSummary}
          </p>
        </article>
        <article
          className="rounded-[20px] border p-5"
          style={{
            background: "var(--surface)",
            borderColor: "var(--border)",
          }}
        >
          <p
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Network
          </p>
          <p className="mt-2 text-3xl font-semibold tracking-tight">
            {networkSummary}
          </p>
          <p
            className="mt-1 text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            In + out (last 24h)
          </p>
        </article>
      </div>
    </section>
  );
}
