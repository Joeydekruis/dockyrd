"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import {
  patchServicePreviewConfig,
  type Service,
  type ServicePreviewConfig,
} from "@/lib/api";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type Props = {
  service: Service;
  onSaved: () => void;
  onError: (message: string | null) => void;
  cardClassName?: string;
  cardStyle?: React.CSSProperties;
};

function readPreviewConfig(service: Service): ServicePreviewConfig {
  const cfg = service.config as { preview?: ServicePreviewConfig };
  return cfg.preview ?? { enabled: false };
}

export function ServicePreviewSettingsCard({
  service,
  onSaved,
  onError,
  cardClassName = RESOURCE_DETAIL_CARD,
  cardStyle = resourceDetailCardStyle(),
}: Props) {
  const saved = readPreviewConfig(service);
  const [enabled, setEnabled] = useState(saved.enabled);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    setEnabled(readPreviewConfig(service).enabled);
  }, [service]);

  async function onSave() {
    setBusy(true);
    try {
      await patchServicePreviewConfig(service.id, { enabled });
      onError(null);
      onSaved();
    } catch (e) {
      onError(e instanceof Error ? e.message : "Save failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className={cardClassName} style={cardStyle}>
      <h2 className="text-lg font-semibold tracking-tight">
        Preview deployments
      </h2>
      <p className="mt-2 text-sm" style={{ color: "var(--muted-text)" }}>
        Spin up a temporary app URL for each pull request against your deploy
        branch. Requires GitHub App{" "}
        <span className="font-mono text-[12px]">pull_request</span> webhooks.
      </p>
      <label className="mt-4 flex cursor-pointer items-center gap-3 text-sm">
        <input
          type="checkbox"
          checked={enabled}
          onChange={(e) => setEnabled(e.target.checked)}
          className="size-4 rounded border"
        />
        Enable preview deployments for this service
      </label>
      <div className="mt-4">
        <Button type="button" disabled={busy} onClick={() => void onSave()}>
          Save preview settings
        </Button>
      </div>
    </section>
  );
}
