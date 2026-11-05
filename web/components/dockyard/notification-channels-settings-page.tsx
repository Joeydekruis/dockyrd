"use client";

import { MoreHorizontal, Plus } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { NotificationChannelModal } from "@/components/dockyard/notification-channel-modal";
import { NotificationDeliveryLogModal } from "@/components/dockyard/notification-delivery-log-modal";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { PopoverMenu } from "@/components/dockyard/ui/popover-menu";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import {
  deleteNotificationChannel,
  getNotificationSettings,
  listNotificationChannels,
  patchNotificationChannel,
  testNotificationChannel,
  type NotificationChannel,
} from "@/lib/api";

function statusVariant(status: string): "ok" | "warn" | "err" | "idle" {
  switch (status) {
    case "active":
      return "ok";
    case "degraded":
      return "warn";
    case "disabled":
      return "idle";
    default:
      return "idle";
  }
}

export function NotificationChannelsSettingsPage() {
  const [loading, setLoading] = useState(true);
  const [channels, setChannels] = useState<NotificationChannel[]>([]);
  const [eventTypes, setEventTypes] = useState<string[]>([]);
  const [addOpen, setAddOpen] = useState(false);
  const [logChannel, setLogChannel] = useState<NotificationChannel | null>(null);
  const [error, setError] = useState<string | null>(null);

  const reload = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [settings, list] = await Promise.all([
        getNotificationSettings(),
        listNotificationChannels(),
      ]);
      setEventTypes(settings.event_types);
      setChannels(list.channels);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void reload();
  }, [reload]);

  async function toggleEnabled(ch: NotificationChannel) {
    await patchNotificationChannel(ch.id, { enabled: !ch.enabled });
    await reload();
  }

  async function runTest(ch: NotificationChannel) {
    await testNotificationChannel(ch.id);
    await reload();
  }

  async function remove(ch: NotificationChannel) {
    if (!window.confirm(`Delete channel "${ch.name}"?`)) return;
    await deleteNotificationChannel(ch.id);
    await reload();
  }

  if (loading) {
    return (
      <DashboardShell title="Notification channels" breadcrumb="Server settings">
        <DockyrdPageLoading />
      </DashboardShell>
    );
  }

  return (
    <DashboardShell title="Notification channels" breadcrumb="Server settings">
      <SettingsPageBody>
        <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
          <p className="max-w-xl text-sm text-[var(--muted-foreground)]">
            Get alerted when deploys fail, backups slip, or the server needs
            attention via Slack, Discord, webhook, or email (instance SMTP).
            Configure SMTP under Email for inbox mail and email channels.
          </p>
          <Button type="button" onClick={() => setAddOpen(true)}>
            <Plus className="mr-2 size-4" aria-hidden />
            Add channel
          </Button>
        </div>
        {error ? (
          <p className="mb-4 text-sm text-[var(--destructive)]">{error}</p>
        ) : null}
        {channels.length === 0 ? (
          <div className="rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-8 text-center text-sm text-[var(--muted-foreground)]">
            No alert channels yet. Add Slack, Discord, webhook, or email to
            receive alerts outside the dashboard.
          </div>
        ) : (
          <ul className="space-y-2">
            {channels.map((ch) => (
              <li key={ch.id}>
                <ResourceRow
                  name={ch.name}
                  href="#"
                  meta={`${ch.kind} · ${ch.secret_hint ?? (ch.secret_set ? "Secret saved" : "No secret")}`}
                  subMeta={
                    ch.last_success_at
                      ? `Last sent ${new Date(ch.last_success_at * 1000).toLocaleString()}`
                      : ch.disabled_reason || ch.last_error || "—"
                  }
                  statusLabel={ch.status}
                  statusVariant={statusVariant(ch.status)}
                  trailing={
                    <PopoverMenu
                      trigger={
                        <button
                          type="button"
                          className="relative z-[2] rounded-lg p-2 hover:bg-[var(--background)]"
                          aria-label={`Actions for ${ch.name}`}
                        >
                          <MoreHorizontal className="size-4" />
                        </button>
                      }
                      items={[
                        {
                          id: "test",
                          label: "Send test",
                          onSelect: () => void runTest(ch),
                        },
                        {
                          id: "toggle",
                          label: ch.enabled ? "Disable" : "Enable",
                          onSelect: () => void toggleEnabled(ch),
                        },
                        {
                          id: "log",
                          label: "Delivery log",
                          onSelect: () => setLogChannel(ch),
                        },
                        {
                          id: "delete",
                          label: "Delete",
                          destructive: true,
                          onSelect: () => void remove(ch),
                        },
                      ]}
                    />
                  }
                />
              </li>
            ))}
          </ul>
        )}
      </SettingsPageBody>
      <NotificationChannelModal
        open={addOpen}
        onClose={() => setAddOpen(false)}
        onSaved={() => void reload()}
        eventTypes={eventTypes}
      />
      <NotificationDeliveryLogModal
        channelId={logChannel?.id ?? null}
        channelName={logChannel?.name ?? ""}
        onClose={() => setLogChannel(null)}
      />
    </DashboardShell>
  );
}
