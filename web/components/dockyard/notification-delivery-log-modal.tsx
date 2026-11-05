"use client";

import { useEffect, useState } from "react";
import { Modal } from "@/components/dockyard/ui/modal";
import { Button } from "@/components/dockyard/ui/button";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  listNotificationDeliveries,
  retryNotificationDelivery,
  type NotificationDelivery,
} from "@/lib/api";

type Props = {
  channelId: string | null;
  channelName: string;
  onClose: () => void;
};

export function NotificationDeliveryLogModal({
  channelId,
  channelName,
  onClose,
}: Props) {
  const [rows, setRows] = useState<NotificationDelivery[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!channelId) return;
    setLoading(true);
    setError(null);
    listNotificationDeliveries(channelId)
      .then((res) => setRows(res.deliveries))
      .catch((e) =>
        setError(e instanceof Error ? e.message : "Could not load log"),
      )
      .finally(() => setLoading(false));
  }, [channelId]);

  async function retry(id: string) {
    await retryNotificationDelivery(id);
    if (channelId) {
      const res = await listNotificationDeliveries(channelId);
      setRows(res.deliveries);
    }
  }

  return (
    <Modal
      open={channelId !== null}
      title="Delivery log"
      description={channelName}
      onClose={onClose}
      size="lg"
    >
      {loading ? (
        <p className="mt-4 text-sm text-[var(--muted-foreground)]">Loading…</p>
      ) : null}
      {error ? (
        <p className="mt-4 text-sm text-[var(--destructive)]">{error}</p>
      ) : null}
      <div className="mt-4 max-h-96 overflow-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-[var(--border)] text-[var(--muted-foreground)]">
              <th className="py-2 pr-2">Time</th>
              <th className="py-2 pr-2">Status</th>
              <th className="py-2 pr-2">Attempts</th>
              <th className="py-2 pr-2">HTTP</th>
              <th className="py-2" />
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id} className="border-b border-[var(--border)]">
                <td className="py-2 pr-2 font-mono text-xs">
                  {new Date(row.created_at * 1000).toLocaleString()}
                </td>
                <td className="py-2 pr-2">
                  <StatusBadge
                    label={row.status}
                    variant={
                      row.status === "sent"
                        ? "ok"
                        : row.status === "dead"
                          ? "err"
                          : "warn"
                    }
                  />
                </td>
                <td className="py-2 pr-2">{row.attempts}</td>
                <td className="py-2 pr-2">{row.http_status ?? "—"}</td>
                <td className="py-2 text-right">
                  {row.status === "dead" ? (
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => retry(row.id)}
                    >
                      Retry
                    </Button>
                  ) : null}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {rows.length === 0 && !loading ? (
          <p className="py-6 text-center text-[var(--muted-foreground)]">
            No deliveries yet.
          </p>
        ) : null}
      </div>
    </Modal>
  );
}
