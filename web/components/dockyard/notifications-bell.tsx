"use client";

import { Bell } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  type InboxFilter,
  NotificationFilterPills,
  NotificationRow,
  NotificationsPanelHeader,
} from "@/components/dockyard/notifications-ui";
import {
  getInboxSummary,
  type InboxNotification,
  listInboxNotifications,
  markAllInboxNotificationsRead,
  markInboxNotificationRead,
} from "@/lib/api";

type NotificationsBellProps = {
  className?: string;
};

export function NotificationsBell({ className }: NotificationsBellProps) {
  const [open, setOpen] = useState(false);
  const [unread, setUnread] = useState(0);
  const [filter, setFilter] = useState<InboxFilter>("all");
  const [items, setItems] = useState<InboxNotification[]>([]);
  const [loading, setLoading] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);

  const refreshSummary = useCallback(async () => {
    try {
      const summary = await getInboxSummary();
      setUnread(summary.unread);
    } catch {
      setUnread(0);
    }
  }, []);

  const loadList = useCallback(async (nextFilter: InboxFilter) => {
    setLoading(true);
    try {
      const category =
        nextFilter === "deploys" ||
        nextFilter === "usage" ||
        nextFilter === "system"
          ? nextFilter
          : undefined;
      const res = await listInboxNotifications({
        category,
        unread: nextFilter === "unread",
        limit: 12,
      });
      setItems(res.notifications);
    } catch {
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refreshSummary();
    const id = window.setInterval(() => void refreshSummary(), 60_000);
    return () => window.clearInterval(id);
  }, [refreshSummary]);

  useEffect(() => {
    if (!open) return;
    void loadList(filter);
  }, [open, filter, loadList]);

  useEffect(() => {
    if (!open) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open]);

  async function handleMarkRead(id: string) {
    try {
      await markInboxNotificationRead(id);
      setItems((prev) =>
        prev.map((n) =>
          n.id === id ? { ...n, read_at: Math.floor(Date.now() / 1000) } : n,
        ),
      );
      await refreshSummary();
    } catch {
      /* ignore */
    }
  }

  async function handleMarkAllRead() {
    try {
      await markAllInboxNotificationsRead();
      setItems((prev) =>
        prev.map((n) => ({
          ...n,
          read_at: n.read_at ?? Math.floor(Date.now() / 1000),
        })),
      );
      setUnread(0);
    } catch {
      /* ignore */
    }
  }

  return (
    <>
      <button
        type="button"
        className={`btn-outline relative flex size-10 items-center justify-center rounded-full p-0 ${className ?? ""}`}
        aria-label={
          unread > 0 ? `Notifications, ${unread} unread` : "Notifications"
        }
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
      >
        <Bell className="size-[18px]" strokeWidth={1.8} aria-hidden />
        {unread > 0 ? (
          <span
            className="absolute -top-0.5 -right-0.5 flex min-w-[18px] items-center justify-center rounded-full px-1 text-[10px] font-semibold text-white"
            style={{ background: "#EF4444" }}
          >
            {unread > 99 ? "99+" : unread}
          </span>
        ) : null}
      </button>

      {open ? (
        <div className="fixed inset-0 z-[60]">
          <button
            type="button"
            className="absolute inset-0 border-0"
            style={{ background: "rgba(11, 31, 58, 0.45)" }}
            aria-label="Close notifications"
            onClick={() => setOpen(false)}
          />
          <div
            ref={panelRef}
            role="dialog"
            aria-modal="true"
            aria-label="Notifications"
            className="absolute top-[calc(var(--dockyard-announcement-height,44px)+4.5rem)] right-8 z-10 flex max-h-[min(640px,calc(100dvh-7rem))] w-[min(420px,calc(100vw-2rem))] flex-col overflow-hidden rounded-[20px] border shadow-lg"
            style={{
              background: "var(--surface)",
              borderColor: "var(--border)",
            }}
          >
            <NotificationsPanelHeader
              title="Notifications"
              onClose={() => setOpen(false)}
              onMarkAllRead={() => void handleMarkAllRead()}
              markAllDisabled={unread === 0}
            />
            <div
              className="border-b px-5 py-3"
              style={{ borderColor: "var(--border)" }}
            >
              <NotificationFilterPills
                value={filter}
                unreadCount={unread}
                onChange={setFilter}
              />
            </div>
            <div className="min-h-0 flex-1 overflow-y-auto">
              {loading ? (
                <p
                  className="px-5 py-8 text-center text-sm"
                  style={{ color: "var(--muted-text)" }}
                >
                  Loading…
                </p>
              ) : items.length === 0 ? (
                <p
                  className="px-5 py-8 text-center text-sm"
                  style={{ color: "var(--muted-text)" }}
                >
                  No notifications yet.
                </p>
              ) : (
                items.map((item) => (
                  <NotificationRow
                    key={item.id}
                    item={item}
                    compact
                    onMarkRead={(id) => void handleMarkRead(id)}
                    onNavigate={() => setOpen(false)}
                  />
                ))
              )}
            </div>
            <div
              className="border-t px-5 py-3 text-center"
              style={{ borderColor: "var(--border)" }}
            >
              <Link
                href="/notifications"
                className="btn-link text-sm font-medium"
                onClick={() => setOpen(false)}
              >
                View all notifications →
              </Link>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
