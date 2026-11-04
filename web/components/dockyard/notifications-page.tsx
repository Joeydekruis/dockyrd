"use client";

import { Check } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  groupNotificationsByDay,
  NotificationFilterPills,
  NotificationRow,
  type InboxFilter,
} from "@/components/dockyard/notifications-ui";
import { Button } from "@/components/dockyard/ui/button";
import {
  getInboxSummary,
  listInboxNotifications,
  markAllInboxNotificationsRead,
  markInboxNotificationRead,
  type InboxNotification,
} from "@/lib/api";

export function NotificationsPage() {
  const [filter, setFilter] = useState<InboxFilter>("all");
  const [items, setItems] = useState<InboxNotification[]>([]);
  const [unread, setUnread] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);

  const refreshSummary = useCallback(async () => {
    const summary = await getInboxSummary();
    setUnread(summary.unread);
    setTotal(summary.total);
  }, []);

  const load = useCallback(
    async (opts?: { append?: boolean; before?: number }) => {
      if (opts?.append) setLoadingMore(true);
      else setLoading(true);
      try {
        const category =
          filter === "deploys" || filter === "usage" || filter === "system"
            ? filter
            : undefined;
        const res = await listInboxNotifications({
          category,
          unread: filter === "unread",
          limit: 25,
          before: opts?.before,
        });
        setItems((prev) =>
          opts?.append ? [...prev, ...res.notifications] : res.notifications,
        );
        await refreshSummary();
      } catch {
        if (!opts?.append) setItems([]);
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [filter, refreshSummary],
  );

  useEffect(() => {
    void load();
  }, [load]);

  async function handleMarkRead(id: string) {
    await markInboxNotificationRead(id);
    setItems((prev) =>
      prev.map((n) =>
        n.id === id ? { ...n, read_at: Math.floor(Date.now() / 1000) } : n,
      ),
    );
    await refreshSummary();
  }

  async function handleMarkAllRead() {
    await markAllInboxNotificationsRead();
    setItems((prev) =>
      prev.map((n) => ({
        ...n,
        read_at: n.read_at ?? Math.floor(Date.now() / 1000),
      })),
    );
    setUnread(0);
  }

  const groups = groupNotificationsByDay(items);
  const lastCreated = items.at(-1)?.created_at;

  return (
    <DashboardShell title="Notifications" breadcrumb="Inbox">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Inbox
          </p>
          <h1 className="mt-1 text-2xl font-semibold tracking-tight">
            Notifications
          </h1>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            {unread} unread, {total} in total
          </p>
        </div>
        <Button
          variant="outline"
          iconLeft={Check}
          onClick={() => void handleMarkAllRead()}
          disabled={unread === 0}
        >
          Mark all as read
        </Button>
      </div>

      <div className="mt-6">
        <NotificationFilterPills
          value={filter}
          unreadCount={unread}
          onChange={setFilter}
        />
      </div>

      <div
        className="mt-6 overflow-hidden rounded-[20px] border"
        style={{ borderColor: "var(--border)", background: "var(--surface)" }}
      >
        {loading ? (
          <p
            className="px-6 py-12 text-center text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            Loading…
          </p>
        ) : items.length === 0 ? (
          <p
            className="px-6 py-12 text-center text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            No notifications match this filter.
          </p>
        ) : (
          groups.map((group) => (
            <section key={group.label}>
              <h2
                className="border-b px-4 py-2 font-mono text-[10px] uppercase tracking-widest"
                style={{
                  borderColor: "var(--border)",
                  background: "var(--page-bg)",
                  color: "var(--muted-text)",
                }}
              >
                {group.label}
              </h2>
              {group.items.map((item) => (
                <NotificationRow
                  key={item.id}
                  item={item}
                  onMarkRead={(id) => void handleMarkRead(id)}
                />
              ))}
            </section>
          ))
        )}
        {!loading && items.length > 0 && items.length < total ? (
          <div
            className="border-t px-4 py-3 text-center"
            style={{ borderColor: "var(--border)" }}
          >
            <button
              type="button"
              className="btn-link text-sm font-medium"
              disabled={loadingMore}
              onClick={() =>
                void load({
                  append: true,
                  before: lastCreated,
                })
              }
            >
              {loadingMore ? "Loading…" : "Load older notifications"}
            </button>
          </div>
        ) : null}
      </div>
    </DashboardShell>
  );
}
