"use client";

import { AlertTriangle, Check, Info, Tag, X, XCircle } from "lucide-react";
import Link from "next/link";
import type { InboxNotification } from "@/lib/api";
import { formatTimeAgo } from "@/lib/resource-dashboard";

const levelStyles: Record<
  InboxNotification["level"],
  { bg: string; Icon: typeof XCircle }
> = {
  error: { bg: "rgba(220, 38, 38, 0.12)", Icon: XCircle },
  warning: { bg: "rgba(234, 179, 8, 0.15)", Icon: AlertTriangle },
  success: { bg: "rgba(22, 163, 74, 0.12)", Icon: Check },
  info: { bg: "rgba(59, 147, 245, 0.12)", Icon: Info },
};

type NotificationRowProps = {
  item: InboxNotification;
  compact?: boolean;
  onMarkRead?: (id: string) => void;
  onNavigate?: () => void;
};

export function NotificationRow({
  item,
  compact,
  onMarkRead,
  onNavigate,
}: NotificationRowProps) {
  const unread = !item.read_at;
  const { bg, Icon } = levelStyles[item.level];
  const time = formatTimeAgo(item.created_at) ?? "";

  async function handleCtaClick() {
    if (unread && onMarkRead) {
      onMarkRead(item.id);
    }
    onNavigate?.();
  }

  return (
    <article
      className={`flex gap-3 border-b py-4 last:border-b-0 ${compact ? "px-5" : "px-4"}`}
      style={{ borderColor: "var(--border)" }}
    >
      <div className="flex w-3 shrink-0 justify-center pt-2">
        {unread ? (
          <span
            className="size-2 rounded-full"
            style={{ background: "var(--accent-blue)" }}
            aria-hidden
          />
        ) : null}
      </div>
      <div
        className="flex size-10 shrink-0 items-center justify-center rounded-xl"
        style={{ background: bg }}
        aria-hidden
      >
        <Icon
          className="size-5"
          strokeWidth={1.8}
          style={{
            color:
              item.level === "error"
                ? "#DC2626"
                : item.level === "warning"
                  ? "#CA8A04"
                  : item.level === "success"
                    ? "#16A34A"
                    : "var(--accent-blue)",
          }}
        />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-3">
          <span
            className="inline-flex max-w-[min(100%,14rem)] min-w-0 items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] leading-snug"
            style={{ borderColor: "var(--border)", color: "var(--muted-text)" }}
          >
            <Tag className="size-3 shrink-0" strokeWidth={1.8} aria-hidden />
            <span className="truncate">{item.source_label}</span>
          </span>
          <time
            className="shrink-0 text-xs whitespace-nowrap"
            style={{ color: "var(--muted-text)" }}
            dateTime={new Date(item.created_at * 1000).toISOString()}
          >
            {time}
          </time>
        </div>
        <h3 className="mt-2 text-base font-semibold tracking-tight">
          {item.title}
        </h3>
        {item.body ? (
          <p
            className="mt-1 text-sm leading-relaxed"
            style={{ color: "var(--muted-text)" }}
          >
            {item.body}
          </p>
        ) : null}
        {item.cta_label && item.cta_href ? (
          <Link
            href={item.cta_href}
            className="btn-link mt-2 inline-block text-sm font-medium"
            onClick={() => void handleCtaClick()}
          >
            {item.cta_label} →
          </Link>
        ) : null}
      </div>
    </article>
  );
}

export type InboxFilter = "all" | "unread" | "deploys" | "usage" | "system";

type NotificationFilterPillsProps = {
  value: InboxFilter;
  unreadCount: number;
  onChange: (value: InboxFilter) => void;
};

const filters: { id: InboxFilter; label: (unread: number) => string }[] = [
  { id: "all", label: () => "All" },
  { id: "unread", label: (n) => (n > 0 ? `Unread · ${n}` : "Unread") },
  { id: "deploys", label: () => "Deploys" },
  { id: "usage", label: () => "Usage" },
  { id: "system", label: () => "System" },
];

export function NotificationFilterPills({
  value,
  unreadCount,
  onChange,
}: NotificationFilterPillsProps) {
  return (
    <div
      className="flex flex-wrap gap-2"
      role="tablist"
      aria-label="Filter notifications"
    >
      {filters.map(({ id, label }) => {
        const active = value === id;
        return (
          <button
            key={id}
            type="button"
            role="tab"
            aria-selected={active}
            className="rounded-full border px-3 py-1.5 font-mono text-[11px] uppercase tracking-widest transition-colors"
            style={{
              borderColor: "var(--border)",
              background: active ? "var(--ink-navy)" : "var(--surface)",
              color: active ? "#fff" : "var(--muted-text)",
            }}
            onClick={() => onChange(id)}
          >
            {label(unreadCount)}
          </button>
        );
      })}
    </div>
  );
}

export function groupNotificationsByDay(items: InboxNotification[]) {
  const now = new Date();
  const startToday = new Date(
    now.getFullYear(),
    now.getMonth(),
    now.getDate(),
  ).getTime();
  const startYesterday = startToday - 86_400_000;

  const groups: { label: string; items: InboxNotification[] }[] = [];
  const buckets = new Map<string, InboxNotification[]>();

  for (const item of items) {
    const t = item.created_at * 1000;
    let label = "Earlier";
    if (t >= startToday) label = "Today";
    else if (t >= startYesterday) label = "Yesterday";
    const list = buckets.get(label) ?? [];
    list.push(item);
    buckets.set(label, list);
  }

  for (const label of ["Today", "Yesterday", "Earlier"]) {
    const list = buckets.get(label);
    if (list?.length) groups.push({ label, items: list });
  }
  return groups;
}

export function NotificationsPanelHeader({
  title,
  onClose,
  onMarkAllRead,
  markAllDisabled,
}: {
  title: string;
  onClose?: () => void;
  onMarkAllRead?: () => void;
  markAllDisabled?: boolean;
}) {
  return (
    <div
      className="flex items-start justify-between gap-3 border-b px-6 py-5"
      style={{ borderColor: "var(--border)" }}
    >
      <h2 className="text-xl font-semibold tracking-tight">{title}</h2>
      <div className="flex items-center gap-3">
        {onMarkAllRead ? (
          <button
            type="button"
            className="btn-link text-sm font-medium"
            disabled={markAllDisabled}
            onClick={onMarkAllRead}
          >
            Mark all as read
          </button>
        ) : null}
        {onClose ? (
          <button
            type="button"
            className="btn-ghost rounded-lg p-2"
            aria-label="Close notifications"
            onClick={onClose}
          >
            <X className="size-5" strokeWidth={1.8} />
          </button>
        ) : null}
      </div>
    </div>
  );
}
