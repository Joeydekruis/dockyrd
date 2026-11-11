"use client";

import { useEffect, useMemo, useState } from "react";
import { areaLineOption } from "@/components/dockyard/analytics/chart-options";
import { EChart } from "@/components/dockyard/ui/echart";
import {
  average,
  cpuPercentFraction,
  diskUsedPercent,
  formatBytes,
  memUsedPercent,
  peak,
} from "@/lib/analytics-stats";
import {
  getServerInfo,
  getServerMetrics,
  type HostMetricPoint,
} from "@/lib/api";
import { useUserLocale } from "@/lib/user-locale-context";

type ChartCard = {
  label: string;
  now: string;
  nowSub: string;
  values: number[];
  yFormatter: (v: number) => string;
  avgLabel: string;
  peakLabel: string;
};

export function DashboardMonitoringSection() {
  const { metricTimeLabels } = useUserLocale();
  const [points, setPoints] = useState<HostMetricPoint[]>([]);
  const [cpuCount, setCpuCount] = useState<number | undefined>();

  useEffect(() => {
    let cancelled = false;
    Promise.all([getServerMetrics("7d"), getServerInfo()])
      .then(([metrics, info]) => {
        if (cancelled) return;
        setPoints(metrics.points);
        setCpuCount(info.cpu_count);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, []);

  const latest = points.at(-1);

  const cards = useMemo((): ChartCard[] => {
    if (points.length === 0) {
      return [
        {
          label: "CPU",
          now: "—",
          nowSub: "",
          values: [],
          yFormatter: (v) => `${v.toFixed(0)}%`,
          avgLabel: "—",
          peakLabel: "—",
        },
        {
          label: "Memory",
          now: "—",
          nowSub: "",
          values: [],
          yFormatter: (v) => `${v.toFixed(0)}%`,
          avgLabel: "—",
          peakLabel: "—",
        },
        {
          label: "Disk",
          now: "—",
          nowSub: "",
          values: [],
          yFormatter: (v) => `${v.toFixed(0)}%`,
          avgLabel: "—",
          peakLabel: "—",
        },
      ];
    }

    const cpuSeries = points.map((p) => cpuPercentFraction(p.cpu));
    const memSeries = points.map((p) => memUsedPercent(p));
    const diskSeries = points.map((p) => diskUsedPercent(p));

    const cpuNow = latest ? cpuPercentFraction(latest.cpu) : 0;
    const memNow = latest ? memUsedPercent(latest) : 0;
    const diskNow = latest ? diskUsedPercent(latest) : 0;

    return [
      {
        label: "CPU",
        now: `${cpuNow.toFixed(0)}%`,
        nowSub: cpuCount ? `of ${cpuCount} vCPU` : "",
        values: cpuSeries,
        yFormatter: (v) => `${v.toFixed(0)}%`,
        avgLabel: `${average(cpuSeries).toFixed(0)}%`,
        peakLabel: `${peak(cpuSeries).toFixed(0)}%`,
      },
      {
        label: "Memory",
        now: latest ? formatBytes(latest.mem_used) : "—",
        nowSub: latest
          ? `of ${formatBytes(latest.mem_total)} · ${memNow.toFixed(0)}%`
          : "",
        values: memSeries,
        yFormatter: (v) => `${v.toFixed(0)}%`,
        avgLabel: `${average(memSeries).toFixed(0)}%`,
        peakLabel: `${peak(memSeries).toFixed(0)}%`,
      },
      {
        label: "Disk",
        now: latest ? formatBytes(latest.disk_used) : "—",
        nowSub: latest
          ? `of ${formatBytes(latest.disk_total)} · ${diskNow.toFixed(0)}%`
          : "",
        values: diskSeries,
        yFormatter: (v) => `${v.toFixed(0)}%`,
        avgLabel: `${average(diskSeries).toFixed(0)}%`,
        peakLabel: `${peak(diskSeries).toFixed(0)}%`,
      },
    ];
  }, [cpuCount, latest, points]);

  const labels = metricTimeLabels(points);

  return (
    <section
      aria-label="Server monitoring, last 7 days"
      className="flex flex-col gap-4"
    >
      {cards.map((card) => (
        <article
          key={card.label}
          className="flex flex-col gap-3 rounded-[20px] border p-5"
          style={{
            background: "var(--surface)",
            borderColor: "var(--border)",
          }}
        >
          <div className="flex flex-wrap items-baseline justify-between gap-3">
            <p className="text-base font-semibold">
              {card.label}{" "}
              <span
                className="text-[13px] font-normal"
                style={{ color: "var(--muted-text)" }}
              >
                · last 7 days
              </span>
            </p>
            <p className="flex items-baseline gap-1.5">
              <span className="text-base font-semibold tracking-tight">
                {card.now}
              </span>
              {card.nowSub ? (
                <span
                  className="text-[13px]"
                  style={{ color: "var(--muted-text)" }}
                >
                  {card.nowSub}
                </span>
              ) : null}
            </p>
          </div>
          {card.values.length > 0 ? (
            <EChart
              height={140}
              ariaLabel={`${card.label} over the last 7 days`}
              option={areaLineOption({
                labels,
                values: card.values,
                yMax: 100,
                yFormatter: card.yFormatter,
                seriesName: card.label,
                unit: "%",
              })}
            />
          ) : (
            <p className="text-sm" style={{ color: "var(--muted-text)" }}>
              Metrics will appear after the server collector has samples.
            </p>
          )}
          {card.values.length > 0 ? (
            <p className="text-[13px]" style={{ color: "var(--muted-text)" }}>
              Avg {card.avgLabel} · Peak {card.peakLabel}
            </p>
          ) : null}
        </article>
      ))}
    </section>
  );
}
