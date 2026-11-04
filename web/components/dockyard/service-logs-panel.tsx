"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { fetchServiceLogs } from "@/lib/api";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type ServiceLogsPanelProps = {
  serviceId: string;
  title?: string;
  description?: string;
};

export function ServiceLogsPanel({
  serviceId,
  title = "Runtime logs",
  description = "Latest output from the Docker container (json-file driver).",
}: ServiceLogsPanelProps) {
  const [log, setLog] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const text = await fetchServiceLogs(serviceId);
      setLog(text);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load logs");
      setLog("");
    } finally {
      setLoading(false);
    }
  }, [serviceId]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return (
    <section
      className={RESOURCE_DETAIL_CARD}
      style={resourceDetailCardStyle()}
      aria-label={title}
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            {description}
          </p>
        </div>
        <Button type="button" variant="outline" onClick={() => void refresh()}>
          Refresh
        </Button>
      </div>
      {error ? (
        <p className="mt-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
      <pre
        className="mt-4 max-h-[480px] overflow-auto rounded-xl border p-4 font-mono text-[12px] leading-relaxed whitespace-pre-wrap"
        style={{
          borderColor: "var(--border)",
          background: "var(--bg)",
          color: "var(--text)",
        }}
      >
        {loading ? "Loading…" : log.trim() || "No logs yet."}
      </pre>
    </section>
  );
}
