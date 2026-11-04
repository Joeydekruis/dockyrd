export type TimeFormat = "24h" | "12h";

export type UserLocalePrefs = {
  timezone: string;
  time_format: TimeFormat;
};

export const defaultUserLocalePrefs: UserLocalePrefs = {
  timezone: "UTC",
  time_format: "24h",
};

export function browserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}

/** IANA zones for the settings control (browser when available). */
export function listTimeZones(): string[] {
  if (typeof Intl !== "undefined" && "supportedValuesOf" in Intl) {
    return Intl.supportedValuesOf("timeZone");
  }
  return [
    "UTC",
    "Europe/Amsterdam",
    "Europe/London",
    "Europe/Berlin",
    "America/New_York",
    "America/Los_Angeles",
    "Asia/Tokyo",
  ];
}

export function formatClockTime(
  tsUnixSec: number,
  prefs: UserLocalePrefs,
): string {
  const d = new Date(tsUnixSec * 1000);
  if (prefs.time_format === "12h") {
    const parts = new Intl.DateTimeFormat("en-US", {
      timeZone: prefs.timezone,
      hour: "numeric",
      minute: "2-digit",
      hour12: true,
    }).formatToParts(d);
    const hour = parts.find((p) => p.type === "hour")?.value ?? "";
    const minute = parts.find((p) => p.type === "minute")?.value ?? "";
    const dayPeriod =
      parts.find((p) => p.type === "dayPeriod")?.value?.toUpperCase() ?? "";
    if (minute === "00") {
      return `${hour}${dayPeriod}`;
    }
    return `${hour}:${minute}${dayPeriod}`;
  }
  return new Intl.DateTimeFormat("en-GB", {
    timeZone: prefs.timezone,
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(d);
}

export function formatTimeAxisLabel(
  tsUnixSec: number,
  index: number,
  total: number,
  prefs: UserLocalePrefs,
): string {
  if (index === total - 1) return "now";
  return formatClockTime(tsUnixSec, prefs);
}

export function metricTimeLabels(
  points: { ts: number }[],
  prefs: UserLocalePrefs,
): string[] {
  return points.map((p, i, arr) =>
    formatTimeAxisLabel(p.ts, i, arr.length, prefs),
  );
}
