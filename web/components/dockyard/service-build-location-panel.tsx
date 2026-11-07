"use client";

import { Hammer } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/dockyrd/ui/button";
import { SelectField } from "@/components/dockyrd/ui/select-field";
import {
  SettingsDetailBlock,
  SettingsGroupHeading,
} from "@/components/dockyrd/ui/settings-detail-block";
import {
  listBuilders,
  patchServiceBuildLocation,
  type Service,
} from "@/lib/api";
import { formatBuildLocation } from "@/lib/build-resilience";

function readBuildLocation(service: Service): string {
  const cfg = (service.config ?? {}) as Record<string, unknown>;
  const raw = cfg.buildLocation;
  return typeof raw === "string" && raw.trim() ? raw.trim() : "local";
}

type ServiceBuildLocationPanelProps = {
  service: Service;
  onSaved?: () => void | Promise<void>;
  cardClassName: string;
  cardStyle?: React.CSSProperties;
};

export function ServiceBuildLocationPanel({
  service,
  onSaved,
  cardClassName,
  cardStyle,
}: ServiceBuildLocationPanelProps) {
  const current = readBuildLocation(service);
  const [builders, setBuilders] = useState<
    Awaited<ReturnType<typeof listBuilders>>["builders"]
  >([]);
  const [choice, setChoice] = useState(current);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setChoice(readBuildLocation(service));
  }, [service]);

  useEffect(() => {
    void listBuilders()
      .then((r) => setBuilders(r.builders))
      .catch(() => setBuilders([]));
  }, []);

  const options = [
    { value: "local", label: "On this server" },
    { value: "github_actions", label: "GitHub Actions (CI build)" },
    { value: "auto", label: "Auto — remote when configured" },
    ...builders.map((b) => ({
      value: `remote:${b.id}`,
      label: `Remote — ${b.name}`,
    })),
  ];

  const onSave = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      await patchServiceBuildLocation(service.id, choice);
      await onSaved?.();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not save");
    } finally {
      setBusy(false);
    }
  }, [choice, onSaved, service]);

  return (
    <div>
      <SettingsGroupHeading icon={Hammer} label="Build location" />
      <section
        className={cardClassName}
        style={cardStyle}
        aria-label="Build location"
      >
        <SettingsDetailBlock
          title="Where builds run"
          description="Use GitHub Actions or a remote BuildKit node when this server is too small for heavy builds."
          rows={[
            {
              label: "Current",
              value: formatBuildLocation(current),
            },
          ]}
        />
        <div className="mt-4 space-y-3 border-t pt-4">
          <SelectField
            label="Build location"
            value={choice}
            onChange={(v) => setChoice(v)}
            options={options}
          />
          {error ? (
            <p className="text-sm text-[var(--danger)]">{error}</p>
          ) : null}
          <Button
            type="button"
            disabled={busy || choice === current}
            onClick={() => void onSave()}
          >
            {busy ? "Saving…" : "Save build location"}
          </Button>
        </div>
      </section>
    </div>
  );
}
