"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { NotificationRow } from "@/components/dockyard/notifications-ui";
import {
  type InboxNotification,
  listInboxNotifications,
  markAllInboxNotificationsRead,
  markInboxNotificationRead,
} from "@/lib/api";

function isImportantNotification(item: InboxNotification): boolean {
  return item.level === "error" || item.level === "warning";
}

type DashboardActivityPanelProps = {
  className?: string;
};

export function DashboardActivityPanel({
  className = "",
}: DashboardActivityPanelProps) {
  const [items, setItems] = useState<InboxNotification[]>([]);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listInboxNotifications({ limit: 40 });
      setItems(res.notifications.filter(isImportantNotification));
    } catch {
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const importantUnread = useMemo(
    () => items.filter((n) => !n.read_at).length,
    [items],
  );

  async function markAllRead() {
    try {
      await markAllInboxNotificationsRead();
      const now = Math.floor(Date.now() / 1000);
      setItems((prev) =>
        prev.map((n) => ({ ...n, read_at: n.read_at ?? now })),
      );
    } catch {
      /* ignore */
    }
  }

  async function markRead(id: string) {
    try {
      await markInboxNotificationRead(id);
      const now = Math.floor(Date.now() / 1000);
      setItems((prev) =>
        prev.map((n) =>
          n.id === id ? { ...n, read_at: n.read_at ?? now } : n,
        ),
      );
    } catch {
      /* ignore */
    }
  }

  return (
    <section
      id="activity"
      aria-labelledby="dashboard-activity-heading"
      className={`flex min-w-0 flex-1 flex-col overflow-hidden rounded-[20px] border ${className}`}
      style={{
        background: "var(--surface)",
        borderColor: "var(--border)",
      }}
    >
      <div
        className="flex flex-col gap-3 border-b px-5 py-4"
        style={{ borderColor: "var(--border)" }}
      >
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div>
            <h2
              id="dashboard-activity-heading"
              className="text-xl font-semibold tracking-tight"
            >
              Activity
            </h2>
            <p
              className="mt-0.5 text-sm"
              style={{ color: "var(--muted-text)" }}
            >
              Failed builds, resource limits and server warnings come first
            </p>
          </div>
          <button
            type="button"
            className="text-sm font-medium hover:underline"
            style={{ color: "var(--link-blue)" }}
            onClick={() => void markAllRead()}
          >
            Mark all as read
          </button>
        </div>
        <span
          className="inline-flex w-fit items-center rounded-full border px-3.5 py-1 font-mono text-[11px] font-medium uppercase tracking-widest"
          style={{
            borderColor: "var(--ink-navy)",
            background: "var(--ink-navy)",
            color: "#ffffff",
          }}
        >
          Important · {importantUnread || items.length}
        </span>
      </div>

      <div className="max-h-[520px] flex-1 overflow-y-auto">
        {loading ? (
          <p
            className="px-5 py-8 text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            Loading notifications…
          </p>
        ) : items.length === 0 ? (
          <p
            className="px-5 py-8 text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            No important warnings or failures right now.
          </p>
        ) : (
          items
            .slice(0, 8)
            .map((item) => (
              <NotificationRow
                key={item.id}
                item={item}
                compact
                onMarkRead={(id) => void markRead(id)}
              />
            ))
        )}
      </div>

      <div
        className="border-t px-5 py-3"
        style={{ borderColor: "var(--border)" }}
      >
        <Link
          href="/notifications"
          className="text-sm font-medium hover:underline"
          style={{ color: "var(--link-blue)" }}
        >
          View all notifications →
        </Link>
      </div>
    </section>
  );
}
