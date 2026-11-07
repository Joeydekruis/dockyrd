"use client";

import { useState } from "react";
import { Switch } from "@/components/dockyrd/ui/switch";
import { type Project, patchProject } from "@/lib/api";

type ProjectLinksSettingsProps = {
  project: Project;
  onUpdated: () => void;
};

export function ProjectLinksSettings({
  project,
  onUpdated,
}: ProjectLinksSettingsProps) {
  const strict = project.links_mode === "strict";
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function setMode(nextStrict: boolean) {
    setBusy(true);
    setError(null);
    try {
      await patchProject(project.id, {
        links_mode: nextStrict ? "strict" : "open",
      });
      onUpdated();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not update project");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div
      className="mt-6 border-t pt-6"
      style={{ borderColor: "var(--border)" }}
    >
      <h3 className="text-base font-semibold">Service links</h3>
      <p className="mt-2 text-sm" style={{ color: "var(--muted-text)" }}>
        In strict mode, only linked services can reach each other on the Docker
        network. Reference variables always require an explicit link.
      </p>
      <div className="mt-4">
        <Switch
          checked={strict}
          disabled={busy}
          onChange={(v) => void setMode(v)}
          label="Strict links (network enforced)"
        />
      </div>
      <p
        className="mt-2 font-mono text-[12px]"
        style={{ color: "var(--muted-text)" }}
      >
        Current mode:{" "}
        {strict ? "Enforced" : "Not enforced (legacy open project network)"}
      </p>
      {error ? (
        <p className="mt-2 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
