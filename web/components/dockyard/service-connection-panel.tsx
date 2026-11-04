"use client";

import { Copy } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import { KeyValueRow } from "@/components/dockyard/ui/key-value-row";
import { getServiceConnection } from "@/lib/api";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

function CopyableMono({ value }: { value: string }) {
  return (
    <span className="inline-flex max-w-full items-center gap-2">
      <span className="break-all font-mono text-sm">{value}</span>
      <IconButton
        icon={Copy}
        label="Copy"
        onClick={() => {
          void navigator.clipboard.writeText(value);
        }}
      />
    </span>
  );
}

type ServiceConnectionPanelProps = {
  serviceId: string;
  title?: string;
  description?: string;
};

export function ServiceConnectionPanel({
  serviceId,
  title = "Connection",
  description = "Use these values from other services on the same project network. Reveal to copy secrets.",
}: ServiceConnectionPanelProps) {
  const [revealed, setRevealed] = useState(false);
  const [variables, setVariables] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    void getServiceConnection(serviceId, revealed)
      .then(({ variables: vars }) => {
        if (!cancelled) setVariables(vars);
      })
      .catch((e) => {
        if (!cancelled) {
          setError(
            e instanceof Error ? e.message : "Failed to load connection",
          );
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [serviceId, revealed]);

  const keys = Object.keys(variables).sort();

  return (
    <section
      className={RESOURCE_DETAIL_CARD}
      style={resourceDetailCardStyle()}
      aria-label={title}
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            {description}
          </p>
        </div>
        <Button
          type="button"
          variant="outline"
          onClick={() => setRevealed((v) => !v)}
        >
          {revealed ? "Hide secrets" : "Reveal secrets"}
        </Button>
      </div>
      {error ? (
        <p className="mt-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : loading ? (
        <p className="mt-4 text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : (
        <dl className="mt-4">
          {keys.map((key) => (
            <KeyValueRow
              key={key}
              label={key}
              value={<CopyableMono value={variables[key] ?? ""} />}
            />
          ))}
          {keys.length === 0 ? (
            <p className="text-sm" style={{ color: "var(--muted-text)" }}>
              No connection variables for this service.
            </p>
          ) : null}
        </dl>
      )}
    </section>
  );
}
