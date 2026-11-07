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
  type NotificationChannel,
  patchNotificationChannel,
  patchNotificationSettings,
  testNotificationChannel,
} from "@/lib/api";

const settingsCardClass =
  "mb-8 space-y-4 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6";

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
  const [logChannel, setLogChannel] = useState<NotificationChannel | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);
  const [diskWarnPct, setDiskWarnPct] = useState(85);
  const [diskCritPct, setDiskCritPct] = useState(95);
  const [certWarnPct, setCertWarnPct] = useState(15);
  const [heartbeatURL, setHeartbeatURL] = useState("");
  const [thresholdSaving, setThresholdSaving] = useState(false);
  const [thresholdSaved, setThresholdSaved] = useState(false);

  const reload = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [settings, list] = await Promise.all([
        getNotificationSettings(),
        listNotificationChannels(),
      ]);
      setDiskWarnPct(settings.settings.disk_warn_pct);
      setDiskCritPct(settings.settings.disk_crit_pct);
      setCertWarnPct(settings.settings.cert_warn_pct);
      setHeartbeatURL(settings.settings.heartbeat_url ?? "");
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

  async function saveThresholds() {
    setThresholdSaving(true);
    setThresholdSaved(false);
    setError(null);
    try {
      await patchNotificationSettings({
        disk_warn_pct: diskWarnPct,
        disk_crit_pct: diskCritPct,
        cert_warn_pct: certWarnPct,
      });
      setThresholdSaved(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save thresholds");
    } finally {
      setThresholdSaving(false);
    }
  }

  if (loading) {
    return (
      <DashboardShell
        title="Notification channels"
        breadcrumb="Server settings"
      >
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
        <section
          className={settingsCardClass}
          aria-labelledby="alert-thresholds"
        >
          <div>
            <h2
              id="alert-thresholds"
              className="text-base font-semibold text-[var(--foreground)]"
            >
              Alert thresholds
            </h2>
            <p className="mt-1 text-sm text-[var(--muted-foreground)]">
              Disk alerts resolve below 80% usage. Certificate warnings use
              remaining lifetime as a percentage of the cert validity period.
            </p>
          </div>
          <div className="grid gap-4 sm:grid-cols-3">
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">
                Disk warning (%)
              </span>
              <input
                type="number"
                min={1}
                max={100}
                className="mt-1 w-full rounded-xl border border-[var(--border)] bg-[var(--background)] px-3 py-2"
                value={diskWarnPct}
                onChange={(e) => setDiskWarnPct(Number(e.target.value))}
              />
            </label>
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">
                Disk critical (%)
              </span>
              <input
                type="number"
                min={1}
                max={100}
                className="mt-1 w-full rounded-xl border border-[var(--border)] bg-[var(--background)] px-3 py-2"
                value={diskCritPct}
                onChange={(e) => setDiskCritPct(Number(e.target.value))}
              />
            </label>
            <label className="block text-sm">
              <span className="text-[var(--muted-foreground)]">
                Certificate warning (% life left)
              </span>
              <input
                type="number"
                min={1}
                max={50}
                className="mt-1 w-full rounded-xl border border-[var(--border)] bg-[var(--background)] px-3 py-2"
                value={certWarnPct}
                onChange={(e) => setCertWarnPct(Number(e.target.value))}
              />
            </label>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <Button
              type="button"
              disabled={thresholdSaving}
              onClick={() => void saveThresholds()}
            >
              {thresholdSaving ? "Saving…" : "Save thresholds"}
            </Button>
            {thresholdSaved ? (
              <span className="text-sm text-[var(--muted-foreground)]">
                Saved
              </span>
            ) : null}
          </div>
        </section>
        <section
          className={settingsCardClass}
          aria-labelledby="external-heartbeat"
        >
          <h2 id="external-heartbeat" className="text-base font-semibold">
            External heartbeat
          </h2>
          <p className="mt-1 text-sm text-[var(--muted-foreground)]">
            Optional GET ping every five minutes for dead-man monitoring (e.g.
            Healthchecks.io, Uptime Kuma). dockyrd cannot alert when the host is
            fully down — this URL lets an external service notice missing pings.
          </p>
          <label className="mt-4 block text-sm">
            <span className="text-[var(--muted-foreground)]">Ping URL</span>
            <input
              className="mt-1 w-full rounded-xl border border-[var(--border)] bg-[var(--background)] px-3 py-2 font-mono text-[13px]"
              value={heartbeatURL}
              onChange={(e) => setHeartbeatURL(e.target.value)}
              placeholder="https://healthchecks.io/ping/…"
            />
          </label>
          <div className="mt-4">
            <Button
              type="button"
              disabled={thresholdSaving}
              onClick={() =>
                void (async () => {
                  setThresholdSaving(true);
                  try {
                    await patchNotificationSettings({
                      heartbeat_url: heartbeatURL.trim(),
                    });
                    setThresholdSaved(true);
                  } catch (e) {
                    setError(
                      e instanceof Error
                        ? e.message
                        : "Failed to save heartbeat",
                    );
                  } finally {
                    setThresholdSaving(false);
                  }
                })()
              }
            >
              Save heartbeat URL
            </Button>
          </div>
        </section>
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
