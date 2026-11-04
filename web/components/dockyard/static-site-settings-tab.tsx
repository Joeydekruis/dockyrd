"use client";

import {
  GitBranch,
  Hammer,
  Rocket,
  Settings,
  TriangleAlert,
} from "lucide-react";
import { useEffect, useState } from "react";
import { Alert } from "@/components/dockyard/ui/alert";
import { Button } from "@/components/dockyard/ui/button";
import { DangerZone } from "@/components/dockyard/ui/danger-zone";
import { Field } from "@/components/dockyard/ui/field";
import { Modal } from "@/components/dockyard/ui/modal";
import { SelectField } from "@/components/dockyard/ui/select-field";
import {
  SettingsBranchPill,
  SettingsCopyPill,
  SettingsDetailBlock,
  SettingsEnabledStatus,
  SettingsGroupHeading,
} from "@/components/dockyard/ui/settings-detail-block";
import { Switch } from "@/components/dockyard/ui/switch";
import { patchService, type Service } from "@/lib/api";
import {
  formatBuildPart,
  isBuildEnabled,
  joinBuildCmd,
  nodeMajorFromImage,
  type StaticSiteConfig,
  staticSiteBuildParts,
} from "@/lib/static-site-display";

type EditModal =
  | null
  | "name"
  | "source"
  | "build"
  | "autoDeploy"
  | "spaFallback";

function cardClassName() {
  return "rounded-[20px] border px-6 py-2";
}

function buildSettingsPatch(
  base: Record<string, unknown>,
  cfg: StaticSiteConfig,
  state: {
    buildEnabled: boolean;
    installCmd: string;
    buildScript: string;
    outputDir: string;
    rootDir: string;
    nodeMajor: string;
    indexFile: string;
    errorFile: string;
  },
): Record<string, unknown> {
  const ic = state.buildEnabled ? state.installCmd.trim() : "";
  const bc = state.buildEnabled ? state.buildScript.trim() : "";
  const prevSource =
    typeof base.source === "object" && base.source !== null
      ? (base.source as Record<string, unknown>)
      : {};
  const mergedSource = {
    ...prevSource,
    type: cfg.source?.type ?? prevSource.type ?? "github",
    repo: cfg.source?.repo ?? prevSource.repo,
    branch: cfg.source?.branch ?? prevSource.branch,
    rootDir: state.rootDir.trim(),
  };
  return {
    ...base,
    buildEnabled: state.buildEnabled,
    installCmd: ic,
    buildScript: bc,
    buildCmd: state.buildEnabled ? joinBuildCmd(ic, bc) : "",
    outputDir: state.outputDir.trim() || "dist",
    nodeImage: `node:${state.nodeMajor}-alpine`,
    indexFile: state.indexFile.trim() || "index.html",
    errorFile: state.errorFile.trim(),
    source: mergedSource,
  };
}

function formatRootDir(rootDir?: string): string {
  const r = (rootDir ?? "").trim();
  if (!r || r === "." || r === "/") return "/";
  return r.startsWith("/") ? r : `/${r}`;
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
  const buildParts = staticSiteBuildParts(cfg);
  const buildOn = isBuildEnabled(cfg);
  const [modal, setModal] = useState<EditModal>(null);
  const [busy, setBusy] = useState(false);

  const [displayName, setDisplayName] = useState(service.name);
  const [repo, setRepo] = useState(cfg.source?.repo ?? "");
  const [branch, setBranch] = useState(cfg.source?.branch ?? "main");
  const [buildEnabled, setBuildEnabled] = useState(buildOn);
  const [installCmd, setInstallCmd] = useState(buildParts.install);
  const [buildScript, setBuildScript] = useState(buildParts.build);
  const [outputDir, setOutputDir] = useState(cfg.outputDir ?? "dist");
  const [rootDir, setRootDir] = useState(cfg.source?.rootDir ?? "");
  const [nodeMajor, setNodeMajor] = useState(nodeMajorFromImage(cfg.nodeImage));
  const [indexFile, setIndexFile] = useState(cfg.indexFile ?? "index.html");
  const [errorFile, setErrorFile] = useState(cfg.errorFile ?? "");
  const [autoDeploy, setAutoDeploy] = useState(cfg.autoDeploy ?? false);
  const [spaFallback, setSpaFallback] = useState(cfg.spaFallback ?? false);

  useEffect(() => {
    setDisplayName(service.name);
  }, [service.name]);

  useEffect(() => {
    const parts = staticSiteBuildParts(cfg);
    setRepo(cfg.source?.repo ?? "");
    setBranch(cfg.source?.branch ?? "main");
    setBuildEnabled(isBuildEnabled(cfg));
    setInstallCmd(parts.install);
    setBuildScript(parts.build);
    setOutputDir(cfg.outputDir ?? "dist");
    setRootDir(cfg.source?.rootDir ?? "");
    setNodeMajor(nodeMajorFromImage(cfg.nodeImage));
    setIndexFile(cfg.indexFile ?? "index.html");
    setErrorFile(cfg.errorFile ?? "");
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
  const branchLabel = (cfg.source?.branch ?? "main").trim() || "main";
  const outputLabel = cfg.outputDir?.trim() ? cfg.outputDir : "dist";

  return (
    <div className="space-y-8">
      <div>
        <SettingsGroupHeading icon={Settings} label="General" />
        <section
          className={cardClassName()}
          style={{
            borderColor: "var(--border)",
            background: "var(--surface)",
          }}
          aria-label="General"
        >
          <SettingsDetailBlock
            title="Details"
            description="How the site is named in the dashboard, and its identifiers for the API."
            actionLabel="Update display name"
            rows={[
              { label: "Display name", value: service.name },
              {
                label: "Name",
                value: <SettingsCopyPill value={service.slug} />,
              },
              {
                label: "ID",
                value: <SettingsCopyPill value={service.id} />,
                span: 2,
              },
            ]}
            onAction={() => setModal("name")}
          />
        </section>
      </div>

      <div>
        <SettingsGroupHeading icon={Rocket} label="Deploy" />
        <section
          className={cardClassName()}
          style={{
            borderColor: "var(--border)",
            background: "var(--surface)",
          }}
          aria-label="Deploy"
        >
          <SettingsDetailBlock
            title="Source"
            description="Where the code that every deployment builds comes from."
            actionLabel="Update source"
            rows={[
              {
                label: "Source type",
                value: cfg.source?.type ? "Private repository" : "—",
              },
              { label: "Git provider", value: sourceType },
              {
                label: "Repository",
                value: (
                  <span className="inline-flex items-center gap-2 font-mono text-[13px]">
                    <GitBranch
                      className="size-3.5 shrink-0"
                      strokeWidth={1.8}
                      style={{ color: "var(--muted-text)" }}
                      aria-hidden
                    />
                    {cfg.source?.repo ?? "—"}
                  </span>
                ),
                span: 2,
              },
            ]}
            onAction={() => setModal("source")}
          />
          <SettingsDetailBlock
            title="Auto-deploy"
            description="Enable or disable automatic deployments from the default branch."
            actionLabel="Update auto-deploy"
            rows={[
              {
                label: "Status",
                value: (
                  <SettingsEnabledStatus enabled={cfg.autoDeploy ?? false} />
                ),
              },
              {
                label: "Default branch",
                value: <SettingsBranchPill branch={branchLabel} />,
              },
            ]}
            onAction={() => setModal("autoDeploy")}
          />
          <SettingsDetailBlock
            title="Single-page app fallback"
            description="Serve index.html for unknown paths (client-side routers)."
            actionLabel="Update SPA fallback"
            rows={[
              {
                label: "Status",
                value: (
                  <SettingsEnabledStatus enabled={cfg.spaFallback ?? false} />
                ),
              },
            ]}
            onAction={() => setModal("spaFallback")}
          />
        </section>
      </div>

      <div>
        <SettingsGroupHeading icon={Hammer} label="Build" />
        <section
          className={cardClassName()}
          style={{
            borderColor: "var(--border)",
            background: "var(--surface)",
          }}
          aria-label="Build"
        >
          <SettingsDetailBlock
            title="Build settings"
            description="Install and build commands, publish folder, and runtime used in the build container."
            actionLabel="Update build settings"
            rows={[
              {
                label: "Build before publish",
                value: <SettingsEnabledStatus enabled={buildOn} />,
              },
              {
                label: "Install command",
                value: (
                  <span className="font-mono text-[13px]">
                    {buildOn ? formatBuildPart(buildParts.install) : "—"}
                  </span>
                ),
              },
              {
                label: "Build command",
                value: (
                  <span className="font-mono text-[13px]">
                    {buildOn ? formatBuildPart(buildParts.build) : "—"}
                  </span>
                ),
              },
              {
                label: "Publish directory",
                value: (
                  <span className="font-mono text-[13px]">{outputLabel}</span>
                ),
              },
              {
                label: "Root directory",
                value: (
                  <span className="font-mono text-[13px]">
                    {formatRootDir(cfg.source?.rootDir)}
                  </span>
                ),
              },
              {
                label: "Node version",
                value: (
                  <span className="font-semibold">
                    {nodeMajorFromImage(cfg.nodeImage)} LTS
                  </span>
                ),
              },
              {
                label: "Index file",
                value: (
                  <span className="font-mono text-[13px]">
                    {cfg.indexFile?.trim() || "index.html"}
                  </span>
                ),
              },
              {
                label: "Error file",
                value: (
                  <span className="font-mono text-[13px]">
                    {cfg.errorFile?.trim() ? cfg.errorFile : "—"}
                  </span>
                ),
              },
            ]}
            onAction={() => setModal("build")}
          />
        </section>
      </div>

      <div>
        <SettingsGroupHeading
          icon={TriangleAlert}
          label="Danger zone"
          tone="danger"
        />
        <DangerZone
          title="Delete static site"
          description="Removes the site, its domains, and deployment history from Dockyard. This cannot be undone."
          action={
            <Button
              type="button"
              variant="danger"
              disabled
              className="font-mono text-[11px] uppercase tracking-widest"
            >
              Delete site
            </Button>
          }
        />
      </div>

      <Modal
        open={modal === "name"}
        title="Update display name"
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
        title="Update source"
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
        title="Update build settings"
        description="How Dockyard builds and publishes this site on each deployment."
        onClose={closeModal}
        size="lg"
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
                  config: buildSettingsPatch(service.config ?? {}, cfg, {
                    buildEnabled,
                    installCmd,
                    buildScript,
                    outputDir,
                    rootDir,
                    nodeMajor,
                    indexFile,
                    errorFile,
                  }),
                })
              }
            >
              Save
            </Button>
          </>
        }
      >
        <div className="space-y-6">
          <div>
            <Switch
              label="Build site before publishing"
              checked={buildEnabled}
              onChange={setBuildEnabled}
            />
            <p
              className="mt-2 pl-14 text-[13px] leading-relaxed"
              style={{ color: "var(--muted-text)" }}
            >
              Run install/build commands in Node.js. Turn off to publish files
              from the repository as they are (using publish directory below).
            </p>
          </div>
          {buildEnabled ? (
            <div className="grid gap-4 md:grid-cols-2">
              <Field
                label="Install command"
                value={installCmd}
                onChange={(e) => setInstallCmd(e.target.value)}
                placeholder="Optional — e.g. npm ci"
              />
              <Field
                label="Build command"
                value={buildScript}
                onChange={(e) => setBuildScript(e.target.value)}
                placeholder="Optional — e.g. npm run build"
              />
            </div>
          ) : null}
          <div className="grid gap-4 md:grid-cols-2">
            <Field
              label="Root directory"
              value={rootDir}
              onChange={(e) => setRootDir(e.target.value)}
              placeholder="./docs-site"
              hint="Where the build runs. Empty is the repository root."
            />
            <Field
              label="Publish directory"
              value={outputDir}
              onChange={(e) => setOutputDir(e.target.value)}
              placeholder="dist"
              hint="The folder that holds the built site."
            />
            <SelectField
              label="Node version"
              value={nodeMajor}
              onChange={setNodeMajor}
              options={[
                { value: "22", label: "22 LTS" },
                { value: "20", label: "20 LTS" },
              ]}
            />
            <Field
              label="Index file"
              value={indexFile}
              onChange={(e) => setIndexFile(e.target.value)}
              placeholder="index.html"
            />
            <Field
              label="Error file"
              value={errorFile}
              onChange={(e) => setErrorFile(e.target.value)}
              placeholder="404.html"
              hint="Served for pages that do not exist."
              className="md:col-span-2"
            />
          </div>
          <Alert
            tone="info"
            description="Changes apply from the next deployment. Turn off build before publish when the site is plain HTML or already built in the repo."
          />
        </div>
      </Modal>

      <Modal
        open={modal === "autoDeploy"}
        title="Update auto-deploy"
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
        title="Update SPA fallback"
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
