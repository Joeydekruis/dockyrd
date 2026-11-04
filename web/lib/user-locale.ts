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

function calendarDayKey(tsUnixSec: number, timeZone: string): string {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date(tsUnixSec * 1000));
}

function metricSpanMultipleDays(
  points: { ts: number }[],
  timeZone: string,
): boolean {
  if (points.length < 2) return false;
  const first = points[0]?.ts;
  const last = points.at(-1)?.ts;
  if (first == null || last == null) return false;
  return calendarDayKey(first, timeZone) !== calendarDayKey(last, timeZone);
}

/** Full date + time for “last updated” style labels (respects DST via IANA zone). */
export function formatDateTime(
  tsUnixSec: number,
  prefs: UserLocalePrefs,
): string {
  const d = new Date(tsUnixSec * 1000);
  const dateOpts: Intl.DateTimeFormatOptions = {
    timeZone: prefs.timezone,
    day: "numeric",
    month: "short",
    year: "numeric",
  };
  const timeOpts: Intl.DateTimeFormatOptions = {
    timeZone: prefs.timezone,
    hour: "2-digit",
    minute: "2-digit",
    hour12: prefs.time_format === "12h",
    timeZoneName: "shortOffset",
  };
  const date = new Intl.DateTimeFormat("en-GB", dateOpts).format(d);
  const time = new Intl.DateTimeFormat("en-GB", timeOpts).format(d);
  return `${date}, ${time}`;
}

export function formatTimeAxisLabel(
  tsUnixSec: number,
  index: number,
  total: number,
  prefs: UserLocalePrefs,
  includeDate?: boolean,
): string {
  if (index === total - 1) return "now";
  if (includeDate) {
    const d = new Date(tsUnixSec * 1000);
    const day = new Intl.DateTimeFormat("en-GB", {
      timeZone: prefs.timezone,
      day: "numeric",
      month: "short",
    }).format(d);
    return `${day} ${formatClockTime(tsUnixSec, prefs)}`;
  }
  return formatClockTime(tsUnixSec, prefs);
}

export function metricTimeLabels(
  points: { ts: number }[],
  prefs: UserLocalePrefs,
): string[] {
  const includeDate = metricSpanMultipleDays(points, prefs.timezone);
  return points.map((p, i, arr) =>
    formatTimeAxisLabel(p.ts, i, arr.length, prefs, includeDate),
  );
}
