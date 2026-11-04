"use client";

import { Copy } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { DangerZone } from "@/components/dockyard/ui/danger-zone";
import { Field } from "@/components/dockyard/ui/field";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import { Modal } from "@/components/dockyard/ui/modal";
import {
  SettingsDetailBlock,
  type SettingsDetailRow,
} from "@/components/dockyard/ui/settings-detail-block";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import { Switch } from "@/components/dockyard/ui/switch";
import { patchService, type Service } from "@/lib/api";
import {
  type StaticSiteConfig,
  splitBuildCmd,
} from "@/lib/static-site-display";

type EditModal =
  | null
  | "name"
  | "source"
  | "build"
  | "autoDeploy"
  | "spaFallback";

function cardClassName() {
  return "rounded-[20px] border px-6 py-5";
}

function joinBuildCmd(install: string, build: string): string {
  const i = install.trim();
  const b = build.trim();
  if (!i) return b;
  if (!b) return i;
  return `${i} && ${b}`;
}

function CopyableValue({ value }: { value: string }) {
  return (
    <span className="inline-flex max-w-full items-center gap-2">
      <span className="truncate font-mono text-sm">{value}</span>
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

function enabledRow(
  enabled: boolean,
  detail?: string,
): SettingsDetailRow["value"] {
  return (
    <span className="inline-flex flex-wrap items-center justify-end gap-2">
      <StatusBadge
        label={enabled ? "Enabled" : "Disabled"}
        variant={enabled ? "ok" : "idle"}
      />
      {detail ? (
        <span
          className="font-mono text-[13px]"
          style={{ color: "var(--muted-text)" }}
        >
          {detail}
        </span>
      ) : null}
    </span>
  );
}

type StaticSiteSettingsTabProps = {
  service: Service;
  cfg: StaticSiteConfig;
  onUpdated: () => Promise<void>;
  onError: (message: string | null) => void;
};

export function StaticSiteSettingsTab({
  service,
  cfg,
  onUpdated,
  onError,
}: StaticSiteSettingsTabProps) {
  const buildParts = splitBuildCmd(cfg.buildCmd ?? "");
  const [modal, setModal] = useState<EditModal>(null);
  const [busy, setBusy] = useState(false);

  const [displayName, setDisplayName] = useState(service.name);
  const [repo, setRepo] = useState(cfg.source?.repo ?? "");
  const [branch, setBranch] = useState(cfg.source?.branch ?? "main");
  const [installCmd, setInstallCmd] = useState(buildParts.install);
  const [buildCmd, setBuildCmd] = useState(buildParts.build);
  const [outputDir, setOutputDir] = useState(cfg.outputDir ?? "dist");
  const [autoDeploy, setAutoDeploy] = useState(cfg.autoDeploy ?? false);
  const [spaFallback, setSpaFallback] = useState(cfg.spaFallback ?? false);

  useEffect(() => {
    setDisplayName(service.name);
  }, [service.name]);

  useEffect(() => {
    const parts = splitBuildCmd(cfg.buildCmd ?? "");
    setRepo(cfg.source?.repo ?? "");
    setBranch(cfg.source?.branch ?? "main");
    setInstallCmd(parts.install);
    setBuildCmd(parts.build);
    setOutputDir(cfg.outputDir ?? "dist");
    setAutoDeploy(cfg.autoDeploy ?? false);
    setSpaFallback(cfg.spaFallback ?? false);
  }, [cfg]);

  function closeModal() {
    if (busy) return;
    setModal(null);
  }

  async function savePatch(body: Parameters<typeof patchService>[1]) {
    setBusy(true);
    onError(null);
    try {
      await patchService(service.id, body);
      await onUpdated();
      setModal(null);
    } catch (e) {
      onError(e instanceof Error ? e.message : "Could not save settings");
    } finally {
      setBusy(false);
    }
  }

  const sourceType =
    (cfg.source?.type ?? "github").toLowerCase() === "github"
      ? "GitHub"
      : (cfg.source?.type ?? "git");

  return (
    <div className="space-y-6">
      <section
        className={cardClassName()}
        style={{
          borderColor: "var(--border)",
          background: "var(--surface)",
        }}
        aria-label="General"
      >
        <h2 className="text-lg font-semibold tracking-tight">General</h2>
        <SettingsDetailBlock
          title="Details"
          description="How the site is named in the dashboard, and its identifiers for the API."
          rows={[
            { label: "Display name", value: service.name },
            { label: "Name", value: <CopyableValue value={service.slug} /> },
            { label: "ID", value: <CopyableValue value={service.id} /> },
          ]}
          onAction={() => setModal("name")}
        />
      </section>

      <section
        className={cardClassName()}
        style={{
          borderColor: "var(--border)",
          background: "var(--surface)",
        }}
        aria-label="Deploy"
      >
        <h2 className="text-lg font-semibold tracking-tight">Deploy</h2>
        <SettingsDetailBlock
          title="Source"
          description="Where the code that every deployment builds comes from."
          rows={[
            {
              label: "Source type",
              value: cfg.source?.type ? "Private repository" : "—",
            },
            { label: "Git provider", value: sourceType },
            {
              label: "Repository",
              value: (
                <span className="font-mono text-sm">
                  {cfg.source?.repo ?? "—"}
                </span>
              ),
            },
            {
              label: "Default branch",
              value: (
                <span className="font-mono text-sm">
                  {cfg.source?.branch ?? "main"}
                </span>
              ),
            },
          ]}
          onAction={() => setModal("source")}
        />
        <SettingsDetailBlock
          title="Auto-deploy"
          description="Enable or disable automatic deployments from the default branch."
          rows={[
            {
              label: "Status",
              value: enabledRow(
                cfg.autoDeploy ?? false,
                (cfg.source?.branch ?? "main").toUpperCase(),
              ),
            },
          ]}
          onAction={() => setModal("autoDeploy")}
        />
        <SettingsDetailBlock
          title="Build settings"
          description="Install and build commands and the folder that gets published."
          rows={[
            { label: "Install command", value: buildParts.install },
            { label: "Build command", value: buildParts.build },
            {
              label: "Output directory",
              value: cfg.outputDir ?? "dist",
            },
          ]}
          onAction={() => setModal("build")}
        />
        <SettingsDetailBlock
          title="Single-page app fallback"
          description="Serve index.html for unknown paths (client-side routers)."
          rows={[
            {
              label: "Status",
              value: enabledRow(cfg.spaFallback ?? false),
            },
          ]}
          onAction={() => setModal("spaFallback")}
        />
      </section>

      <DangerZone
        title="Delete static site"
        description="Removes the site, its domains, and deployment history from Dockyard. This cannot be undone."
        action={
          <Button type="button" variant="outline" disabled>
            Delete static site
          </Button>
        }
      />

      <Modal
        open={modal === "name"}
        title="Edit display name"
        description="Shown in the dashboard and project resource lists."
        onClose={closeModal}
        footer={
          <>
            <Button type="button" variant="outline" onClick={closeModal}>
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy || !displayName.trim()}
              onClick={() => void savePatch({ name: displayName.trim() })}
            >
              Save
            </Button>
          </>
        }
      >
        <Field
          label="Display name"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          autoFocus
        />
      </Modal>

      <Modal
        open={modal === "source"}
        title="Edit source"
        description="Repository and branch used for every deployment."
        onClose={closeModal}
        size="lg"
        footer={
          <>
            <Button type="button" variant="outline" onClick={closeModal}>
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy || !repo.trim()}
              onClick={() =>
                void savePatch({
                  config: {
                    ...(service.config ?? {}),
                    source: {
                      ...(cfg.source ?? {}),
                      type: cfg.source?.type ?? "github",
                      repo: repo.trim(),
                      branch: branch.trim() || "main",
                    },
                  },
                })
              }
            >
              Save
            </Button>
          </>
        }
      >
        <div className="grid gap-4">
          <Field label="Git provider" value={sourceType} readOnly />
          <Field
            label="Repository"
            value={repo}
            onChange={(e) => setRepo(e.target.value)}
            placeholder="org/repo"
          />
          <Field
            label="Default branch"
            value={branch}
            onChange={(e) => setBranch(e.target.value)}
            placeholder="main"
          />
        </div>
      </Modal>

      <Modal
        open={modal === "build"}
        title="Edit build settings"
        description="Commands run in the build container and the publish directory."
        onClose={closeModal}
        size="lg"
        footer={
          <>
            <Button type="button" variant="outline" onClick={closeModal}>
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy || !buildCmd.trim()}
              onClick={() =>
                void savePatch({
                  config: {
                    ...(service.config ?? {}),
                    buildCmd: joinBuildCmd(installCmd, buildCmd),
                    outputDir: outputDir.trim() || "dist",
                  },
                })
              }
            >
              Save
            </Button>
          </>
        }
      >
        <div className="grid gap-4 md:grid-cols-2">
          <Field
            label="Install command"
            value={installCmd}
            onChange={(e) => setInstallCmd(e.target.value)}
          />
          <Field
            label="Build command"
            value={buildCmd}
            onChange={(e) => setBuildCmd(e.target.value)}
          />
          <Field
            label="Output directory"
            value={outputDir}
            onChange={(e) => setOutputDir(e.target.value)}
            className="md:col-span-2"
          />
        </div>
      </Modal>

      <Modal
        open={modal === "autoDeploy"}
        title="Edit auto-deploy"
        description={`Deploy automatically on pushes to ${branch.trim() || "main"}.`}
        onClose={closeModal}
        footer={
          <>
            <Button type="button" variant="outline" onClick={closeModal}>
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy}
              onClick={() =>
                void savePatch({
                  config: {
                    ...(service.config ?? {}),
                    autoDeploy,
                  },
                })
              }
            >
              Save
            </Button>
          </>
        }
      >
        <Switch
          label="Auto-deploy"
          checked={autoDeploy}
          onChange={setAutoDeploy}
        />
      </Modal>

      <Modal
        open={modal === "spaFallback"}
        title="Edit SPA fallback"
        description="When enabled, unknown paths return your site's index.html."
        onClose={closeModal}
        footer={
          <>
            <Button type="button" variant="outline" onClick={closeModal}>
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy}
              onClick={() =>
                void savePatch({
                  config: {
                    ...(service.config ?? {}),
                    spaFallback,
                  },
                })
              }
            >
              Save
            </Button>
          </>
        }
      >
        <Switch
          label="Single-page app fallback"
          checked={spaFallback}
          onChange={setSpaFallback}
        />
      </Modal>
    </div>
  );
}
