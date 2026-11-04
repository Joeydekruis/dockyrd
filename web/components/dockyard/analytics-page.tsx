"use client";

import { useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  containerTableRows,
  CpuTab,
  DiskTab,
  EventsAndOpsSection,
  MemoryTab,
  NetworkTab,
  OverviewTab,
  rangeLabelFromRange,
} from "@/components/dockyard/analytics/analytics-tab-views";
import {
  type AnalyticsTabId,
  useAnalyticsData,
} from "@/components/dockyard/analytics/use-analytics-data";
import { Chip } from "@/components/dockyard/ui/chip";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { Tabs } from "@/components/dockyard/ui/tabs";
import { TimeRangeToggle } from "@/components/dockyard/ui/time-range-toggle";
import { formatBytes, formatUptime } from "@/lib/analytics-stats";

const analyticsTabs = [
  { id: "overview", label: "Overview" },
  { id: "cpu", label: "CPU" },
  { id: "memory", label: "Memory" },
  { id: "disk", label: "Disk" },
  { id: "network", label: "Network" },
] as const;

export function AnalyticsPage() {
  const data = useAnalyticsData();
  const [tab, setTab] = useState<AnalyticsTabId>("overview");

  const points = data.metrics?.points ?? [];
  const latest = points.at(-1);
  const rangeLabel = rangeLabelFromRange(data.range);
  const containerRows = useMemo(
    () => containerTableRows(data.containers),
    [data.containers],
  );

  const tabProps = {
    points,
    latest,
    serverInfo: data.serverInfo,
    containerMetrics: data.containerMetrics,
    rangeLabel,
    onSelectTab: setTab,
  };

  const metaChips = (
    <div className="mb-4 flex flex-wrap gap-2">
      {data.serverInfo?.hostname ? (
        <Chip>{data.serverInfo.hostname.toUpperCase()}</Chip>
      ) : null}
      {data.serverInfo?.cpu_count ? (
        <Chip>{data.serverInfo.cpu_count} VCPU</Chip>
      ) : null}
      {latest?.mem_total ? (
        <Chip>{formatBytes(latest.mem_total)} RAM</Chip>
      ) : null}
      {data.serverInfo?.uptime_sec ? (
        <Chip>UP {formatUptime(data.serverInfo.uptime_sec)}</Chip>
      ) : null}
    </div>
  );

  return (
    <DashboardShell title="Analytics">
      {metaChips}
      <PageHeader
        title="Analytics"
        subtitle="Monitor your server, from the key numbers to the detailed charts."
        action={
          <TimeRangeToggle value={data.range} onChange={data.setRange} />
        }
      />

      <Tabs
        tabs={[...analyticsTabs]}
        activeId={tab}
        onChange={(id) => setTab(id as AnalyticsTabId)}
      />

      <div className="mt-8 space-y-10">
        {data.error ? (
          <p
            className="rounded-lg border px-3 py-2 text-sm"
            style={{
              borderColor: "rgba(179,38,30,0.3)",
              background: "rgba(179,38,30,0.08)",
              color: "#b3261e",
            }}
          >
            {data.error}
          </p>
        ) : null}

        {data.loading && !data.metrics ? (
          <p className="text-sm" style={{ color: "var(--muted-text)" }}>
            Loading metrics…
          </p>
        ) : null}

        {tab === "overview" ? (
          <OverviewTab {...tabProps} />
        ) : null}
        {tab === "cpu" ? <CpuTab {...tabProps} /> : null}
        {tab === "memory" ? <MemoryTab {...tabProps} /> : null}
        {tab === "disk" ? (
          <DiskTab {...tabProps} cleanup={data.cleanup} />
        ) : null}
        {tab === "network" ? <NetworkTab {...tabProps} /> : null}

        {tab === "overview" ? (
          <EventsAndOpsSection
            events={data.events}
            allEvents={data.allEvents}
            eventFilter={data.eventFilter}
            setEventFilter={data.setEventFilter}
            onAck={(id) => void data.ack(id)}
            cleanup={data.cleanup}
            isAdmin={data.isAdmin}
            cleanupBusy={data.cleanupBusy}
            onCleanup={() =>
              void data.runCleanup([
                "dangling_images",
                "build_cache",
                "stopped_unmanaged",
                "old_releases",
              ])
            }
            containerRows={containerRows}
          />
        ) : null}
      </div>
    </DashboardShell>
  );
}
