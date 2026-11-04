import type { EChartsOption } from "echarts";
import { analyticsChartColors as c } from "@/lib/analytics-chart-theme";

const baseGrid = {
  left: 48,
  right: 16,
  top: 16,
  bottom: 28,
  containLabel: false,
};

const baseAxis = {
  axisLine: { show: false },
  axisTick: { show: false },
  axisLabel: { color: c.axis, fontSize: 11 },
  splitLine: { lineStyle: { color: c.grid, type: "dashed" as const } },
};

export function areaLineOption(input: {
  labels: string[];
  values: number[];
  yMax?: number;
  yFormatter?: (v: number) => string;
  alertAt?: number;
  alertLabel?: string;
  seriesName?: string;
  unit?: string;
}): EChartsOption {
  const yMax = input.yMax ?? Math.max(100, ...input.values, input.alertAt ?? 0);
  return {
    animation: false,
    grid: baseGrid,
    tooltip: {
      trigger: "axis",
      valueFormatter: (v) =>
        input.yFormatter
          ? input.yFormatter(Number(v))
          : `${v}${input.unit ?? ""}`,
    },
    xAxis: {
      type: "category",
      data: input.labels,
      boundaryGap: false,
      ...baseAxis,
      splitLine: { show: false },
    },
    yAxis: {
      type: "value",
      max: yMax,
      ...baseAxis,
      axisLabel: {
        ...baseAxis.axisLabel,
        formatter: (v: number) =>
          input.yFormatter ? input.yFormatter(v) : `${v}${input.unit ?? ""}`,
      },
    },
    series: [
      {
        name: input.seriesName ?? "Usage",
        type: "line",
        smooth: true,
        showSymbol: false,
        data: input.values,
        lineStyle: { color: c.primary, width: 2 },
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: c.primaryFillTop },
              { offset: 1, color: c.primaryFillBottom },
            ],
          },
        },
        markLine: input.alertAt
          ? {
              symbol: "none",
              lineStyle: { color: c.alert, type: "dashed", width: 1.5 },
              label: {
                formatter: input.alertLabel ?? `Alert at ${input.alertAt}%`,
                color: c.alert,
                fontSize: 11,
              },
              data: [{ yAxis: input.alertAt }],
            }
          : undefined,
      },
    ],
  };
}

export function dualAreaOption(input: {
  labels: string[];
  primary: number[];
  secondary: number[];
  primaryName: string;
  secondaryName: string;
  yFormatter: (v: number) => string;
}): EChartsOption {
  const max = Math.max(...input.primary, ...input.secondary, 1);
  return {
    animation: false,
    legend: {
      bottom: 0,
      icon: "roundRect",
      textStyle: { color: c.axis, fontSize: 11 },
    },
    grid: { ...baseGrid, bottom: 40 },
    tooltip: {
      trigger: "axis",
      valueFormatter: (v) => input.yFormatter(Number(v)),
    },
    xAxis: {
      type: "category",
      data: input.labels,
      boundaryGap: false,
      ...baseAxis,
      splitLine: { show: false },
    },
    yAxis: {
      type: "value",
      max: max * 1.1,
      ...baseAxis,
      axisLabel: {
        ...baseAxis.axisLabel,
        formatter: (v: number) => input.yFormatter(v),
      },
    },
    series: [
      {
        name: input.primaryName,
        type: "line",
        smooth: true,
        showSymbol: false,
        data: input.primary,
        lineStyle: { color: c.primary, width: 2 },
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: c.primaryFillTop },
              { offset: 1, color: c.primaryFillBottom },
            ],
          },
        },
      },
      {
        name: input.secondaryName,
        type: "line",
        smooth: true,
        showSymbol: false,
        data: input.secondary,
        lineStyle: { color: c.secondary, width: 2 },
      },
    ],
  };
}

export function sparklineOption(values: number[], yMax?: number): EChartsOption {
  return {
    animation: false,
    grid: { left: 0, right: 0, top: 4, bottom: 0 },
    xAxis: { type: "category", show: false, data: values.map((_, i) => i) },
    yAxis: { type: "value", show: false, max: yMax ?? Math.max(100, ...values) },
    series: [
      {
        type: "line",
        data: values,
        smooth: true,
        showSymbol: false,
        lineStyle: { color: c.primary, width: 1.5 },
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: c.primaryFillTop },
              { offset: 1, color: c.primaryFillBottom },
            ],
          },
        },
      },
    ],
  };
}

export function horizontalStackOption(segments: {
  name: string;
  value: number;
  color: string;
}[]): EChartsOption {
  const total = segments.reduce((s, x) => s + x.value, 0) || 1;
  return {
    animation: false,
    grid: { left: 0, right: 0, top: 8, bottom: 0 },
    xAxis: { type: "value", max: total, show: false },
    yAxis: { type: "category", data: [""], show: false },
    series: segments.map((seg) => ({
      name: seg.name,
      type: "bar",
      stack: "total",
      data: [seg.value],
      barWidth: 12,
      itemStyle: { color: seg.color, borderRadius: 2 },
    })),
  };
}
