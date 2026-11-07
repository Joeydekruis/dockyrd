"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/dockyrd/ui/button";
import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyrd/ui/status-badge";
import { getServerContainers, type Service } from "@/lib/api";
import { resourceStatusDisplay } from "@/lib/resource-dashboard";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type ApplicationProcessesTabProps = {
  service: Service;
  port: number;
  onViewLogs?: (composeService?: string) => void;
};

function stateVariant(state: string): StatusVariant {
  const s = state.trim().toLowerCase();
  if (s === "running") return "ok";
  if (s === "exited" || s === "dead") return "err";
  return "idle";
}

export function ApplicationProcessesTab({
  service,
  port,
  onViewLogs,
}: ApplicationProcessesTabProps) {
  const [rows, setRows] = useState<
    { id: string; name: string; state: string }[]
  >([]);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const { containers } = await getServerContainers();
      setRows(
        containers
          .filter((c) => c.service_id === service.id && c.kind === "app")
          .map((c) => ({
            id: c.id,
            name: c.name.replace(/^\//, ""),
            state: c.state,
          })),
      );
    } catch {
      setRows([]);
    } finally {
      setLoading(false);
    }
  }, [service.id]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const desired = resourceStatusDisplay(service.desired_state, "application");
  const processRows = useMemo(() => {
    if (rows.length > 0) return rows;
    return [
      {
        id: "web",
        name: "web",
        state: desired.variant === "ok" ? "running" : "unknown",
      },
    ];
  }, [desired.variant, rows]);

  return (
    <div className="space-y-6">
      <section
        className={RESOURCE_DETAIL_CARD}
        style={resourceDetailCardStyle()}
      >
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold tracking-tight">Processes</h2>
            <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
              Containers running for this application on this server.
            </p>
          </div>
          <Button
            type="button"
            variant="outline"
            onClick={() => void refresh()}
          >
            Refresh
          </Button>
        </div>

        {loading && rows.length === 0 ? (
          <p className="mt-6 text-sm" style={{ color: "var(--muted-text)" }}>
            Loading processes…
          </p>
        ) : null}

        <ul className="mt-6 space-y-3">
          {processRows.map((proc) => (
            <li key={proc.id}>
              <article
                className="rounded-2xl border px-5 py-4"
                style={{
                  borderColor: "var(--border)",
                  background: "var(--hover-row)",
                }}
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="text-[15px] font-semibold tracking-tight">
                      {proc.name.includes("-")
                        ? proc.name.split("-").slice(-2).join("-")
                        : proc.name}
                    </p>
                    <p
                      className="mt-1 font-mono text-[12px]"
                      style={{ color: "var(--muted-text)" }}
                    >
                      {proc.name === "web" ? `HTTP · port ${port}` : proc.name}
                    </p>
                  </div>
                  <StatusBadge
                    label={proc.state === "running" ? "Running" : proc.state}
                    variant={stateVariant(proc.state)}
                  />
                </div>
                {onViewLogs ? (
                  <Button
                    type="button"
                    variant="outline"
                    className="mt-4"
                    onClick={() =>
                      onViewLogs(proc.name === "web" ? undefined : proc.name)
                    }
                  >
                    Open logs
                  </Button>
                ) : null}
              </article>
            </li>
          ))}
        </ul>

        {rows.length === 0 && !loading ? (
          <p
            className="mt-8 border-t pt-6 text-[13px]"
            style={{ borderColor: "var(--border)", color: "var(--muted-text)" }}
          >
            No extra worker processes are configured. When you add background
            workers, they will appear here alongside the web container.
          </p>
        ) : null}
      </section>
    </div>
  );
}
