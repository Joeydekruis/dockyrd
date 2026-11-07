"use client";

import { Hammer, Rocket, Settings } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { ServiceBuildLocationPanel } from "@/components/dockyard/service-build-location-panel";
import { ServicePreviewSettingsCard } from "@/components/dockyard/service-preview-settings-card";
import { Button } from "@/components/dockyard/ui/button";
import { DangerZone } from "@/components/dockyard/ui/danger-zone";
import { Field } from "@/components/dockyard/ui/field";
import { Modal } from "@/components/dockyard/ui/modal";
import { SelectField } from "@/components/dockyard/ui/select-field";
import {
  SettingsCopyPill,
  SettingsDetailBlock,
  SettingsEnabledStatus,
  SettingsGroupHeading,
} from "@/components/dockyard/ui/settings-detail-block";
import { Switch } from "@/components/dockyard/ui/switch";
import {
  createServiceDeployToken,
  deletePanelService,
  getServiceDeployTokenStatus,
  getServiceGitHubWorkflow,
  type Project,
  patchService,
  type Service,
} from "@/lib/api";
import { useCopyToClipboard } from "@/lib/snackbar-context";

type BuilderUi = "railpack" | "dockerfile" | "buildpacks" | "nixpacks";

function parseAppConfig(config: Record<string, unknown>) {
  const source = config.source as
    | { type?: string; repo?: string; branch?: string; ref?: string }
    | undefined;
  const builder = String(
    config.builder ?? "railpack",
  ).toLowerCase() as BuilderUi;
  const port = typeof config.port === "number" ? config.port : 80;
  const autoDeploy = Boolean(config.autoDeploy);
  const healthPath = String(config.healthPath ?? "/");
  const memLimitMB =
    typeof config.memLimitMB === "number" ? config.memLimitMB : 512;
  return { source, builder, port, autoDeploy, healthPath, memLimitMB };
}

function builderLabel(b: string): string {
  const map: Record<string, string> = {
    railpack: "Railpack",
    dockerfile: "Dockerfile",
    buildpacks: "Buildpacks",
    nixpacks: "Nixpacks",
  };
  return map[b] ?? b;
}

function cardClassName() {
  return "rounded-[20px] border px-6 py-2";
}

type ApplicationSettingsTabProps = {
  service: Service;
  project: Project;
  onUpdated: () => Promise<void>;
  onError: (message: string | null) => void;
};

export function ApplicationSettingsTab({
  service,
  project,
  onUpdated,
  onError,
}: ApplicationSettingsTabProps) {
  const router = useRouter();
  const copyToClipboard = useCopyToClipboard();
  const cfg = parseAppConfig(service.config);
  const [editBuilder, setEditBuilder] = useState(false);
  const [editHealth, setEditHealth] = useState(false);
  const [editResources, setEditResources] = useState(false);
  const [deployTokenOpen, setDeployTokenOpen] = useState(false);
  const [deployTokenPlain, setDeployTokenPlain] = useState<string | null>(null);
  const [deployTokenConfigured, setDeployTokenConfigured] = useState<
    boolean | null
  >(null);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [confirmSlug, setConfirmSlug] = useState("");
  const [deleteBusy, setDeleteBusy] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [draftBuilder, setDraftBuilder] = useState(cfg.builder);
  const [draftHealth, setDraftHealth] = useState(cfg.healthPath);
  const [draftPort, setDraftPort] = useState(String(cfg.port));
  const [draftMem, setDraftMem] = useState(String(cfg.memLimitMB));

  const hookUrl = useMemo(() => {
    if (typeof window === "undefined") return "";
    return `${window.location.origin}/api/v1/hooks/deploy/${service.id}`;
  }, [service.id]);

  async function saveConfig(patch: Record<string, unknown>) {
    setBusy(true);
    onError(null);
    try {
      await patchService(service.id, {
        config: { ...service.config, ...patch },
      });
      await onUpdated();
    } catch (e) {
      onError(e instanceof Error ? e.message : "Save failed");
    } finally {
      setBusy(false);
    }
  }

  async function toggleAutoDeploy(next: boolean) {
    await saveConfig({ autoDeploy: next });
  }

  const sourceType = String(cfg.source?.type ?? "image").toLowerCase();
  const gitSourced = sourceType === "github" || sourceType === "git";

  useEffect(() => {
    if (!gitSourced) {
      setDeployTokenConfigured(null);
      return;
    }
    let cancelled = false;
    void getServiceDeployTokenStatus(service.id)
      .then((res) => {
        if (!cancelled) setDeployTokenConfigured(res.configured);
      })
      .catch(() => {
        if (!cancelled) setDeployTokenConfigured(false);
      });
    return () => {
      cancelled = true;
    };
  }, [gitSourced, service.id]);

  async function downloadWorkflowYaml() {
    setBusy(true);
    onError(null);
    try {
      const res = await getServiceGitHubWorkflow(service.id);
      const blob = new Blob([res.yaml], { type: "text/yaml;charset=utf-8" });
      const url = URL.createObjectURL(blob);
      const tail = res.workflow.includes("/")
        ? res.workflow.split("/").pop()
        : null;
      const name = tail && tail.length > 0 ? tail : "dockyard.yml";
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = name;
      anchor.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      onError(e instanceof Error ? e.message : "Could not load workflow");
    } finally {
      setBusy(false);
    }
  }

  async function regenerateDeployToken() {
    setBusy(true);
    onError(null);
    try {
      const res = await createServiceDeployToken(service.id);
      setDeployTokenPlain(res.token);
      setDeployTokenConfigured(true);
      setDeployTokenOpen(true);
    } catch (e) {
      onError(
        e instanceof Error
          ? e.message
          : "Could not create deploy token. Admin role required.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function confirmDelete() {
    setDeleteBusy(true);
    setDeleteError(null);
    try {
      await deletePanelService(service.id, { confirm_slug: confirmSlug });
      router.push("/applications");
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : "Delete failed");
    } finally {
      setDeleteBusy(false);
    }
  }

  return (
    <div className="space-y-8">
      <div>
        <SettingsGroupHeading icon={Hammer} label="Build & deploy" />
        <section
          className={cardClassName()}
          style={{
            borderColor: "var(--border)",
            background: "var(--surface)",
          }}
        >
          <SettingsDetailBlock
            title="Build strategy"
            description="How git-sourced apps are built into a container image."
            actionLabel="Update strategy"
            onAction={() => {
              setDraftBuilder(cfg.builder);
              setEditBuilder(true);
            }}
            rows={[
              { label: "Strategy", value: builderLabel(cfg.builder) },
              {
                label: "Notes",
                value:
                  cfg.builder === "buildpacks"
                    ? "Buildpacks build via Dockerfile in the repository root."
                    : cfg.builder === "nixpacks"
                      ? "Nixpacks uses the Railpack build pipeline."
                      : "—",
                span: 2,
              },
            ]}
          />
          <SettingsDetailBlock
            title="Source"
            description="Repository or image used for deployments."
            actionLabel="Open project"
            onAction={() => {
              window.location.href = `/projects/${project.slug}`;
            }}
            rows={[
              {
                label: "Source",
                value: cfg.source?.repo
                  ? `${cfg.source.repo} · ${cfg.source.branch ?? "main"}`
                  : (cfg.source?.ref ?? cfg.source?.type ?? "—"),
                span: 2,
              },
            ]}
          />
          <SettingsDetailBlock
            title="Auto-deploy"
            description="Deploy automatically when the default branch updates."
            rows={[
              {
                label: "Status",
                value: <SettingsEnabledStatus enabled={cfg.autoDeploy} />,
              },
            ]}
          />
          <div
            className="border-t py-6"
            style={{ borderColor: "var(--border)" }}
          >
            <Switch
              label="Deploy on push"
              checked={cfg.autoDeploy}
              onChange={(v) => void toggleAutoDeploy(v)}
            />
          </div>
          <SettingsDetailBlock
            title="Deploy hook"
            description="POST a built image reference with a deploy token (GitHub Actions or custom CI)."
            actionLabel={gitSourced ? "Regenerate token" : undefined}
            onAction={
              gitSourced ? () => void regenerateDeployToken() : undefined
            }
            rows={[
              {
                label: "Hook URL",
                value: hookUrl ? <SettingsCopyPill value={hookUrl} /> : "—",
                span: 2,
              },
              {
                label: "Status",
                value: gitSourced
                  ? deployTokenConfigured === null
                    ? "Checking token…"
                    : deployTokenConfigured
                      ? "Deploy token configured"
                      : "No deploy token yet — regenerate and add to CI"
                  : "Image-based apps deploy from the dashboard",
              },
            ]}
          />
          {gitSourced ? (
            <SettingsDetailBlock
              title="GitHub Actions"
              description="Workflow that builds on GitHub and deploys via the hook above."
              actionLabel="Download workflow"
              onAction={() => void downloadWorkflowYaml()}
              rows={[
                {
                  label: "File",
                  value: ".github/workflows/dockyard.yml",
                  span: 2,
                },
              ]}
            />
          ) : null}
          {!gitSourced ? (
            <p
              className="border-t py-4 text-[13px]"
              style={{
                borderColor: "var(--border)",
                color: "var(--muted-text)",
              }}
            >
              Connect a GitHub repository in{" "}
              <Link
                href={`/projects/${project.slug}`}
                className="underline"
                style={{ color: "var(--link-blue)" }}
              >
                project settings
              </Link>{" "}
              to use deploy hooks and workflows.
            </p>
          ) : null}
        </section>
      </div>

      <div>
        <SettingsGroupHeading icon={Settings} label="Runtime" />
        <section
          className={cardClassName()}
          style={{
            borderColor: "var(--border)",
            background: "var(--surface)",
          }}
        >
          <SettingsDetailBlock
            title="Health check"
            description="HTTP path probed before switching traffic to a new deployment."
            actionLabel="Update path"
            onAction={() => {
              setDraftHealth(cfg.healthPath);
              setEditHealth(true);
            }}
            rows={[{ label: "Path", value: cfg.healthPath }]}
          />
          <SettingsDetailBlock
            title="Resources"
            description="Container listen port and memory limit from service config."
            actionLabel="Update resources"
            onAction={() => {
              setDraftPort(String(cfg.port));
              setDraftMem(String(cfg.memLimitMB));
              setEditResources(true);
            }}
            rows={[
              { label: "Port", value: String(cfg.port) },
              { label: "Memory", value: `${cfg.memLimitMB} MB` },
            ]}
          />
        </section>
      </div>

      <div>
        <SettingsGroupHeading icon={Rocket} label="Previews & build location" />
        <ServicePreviewSettingsCard
          service={service}
          onSaved={() => void onUpdated()}
          onError={onError}
        />
        <ServiceBuildLocationPanel
          service={service}
          onSaved={() => void onUpdated()}
          cardClassName={cardClassName()}
          cardStyle={{
            borderColor: "var(--border)",
            background: "var(--surface)",
          }}
        />
      </div>

      <DangerZone
        title="Danger zone"
        description="Deletes the application, running containers, domains, deployments, and environment variables."
        action={
          <Button
            type="button"
            variant="danger"
            className="font-mono text-[11px] uppercase tracking-widest"
            onClick={() => {
              setConfirmSlug("");
              setDeleteError(null);
              setDeleteOpen(true);
            }}
          >
            Delete application
          </Button>
        }
      />

      <Modal
        open={editBuilder}
        title="Build strategy"
        onClose={() => setEditBuilder(false)}
        footer={
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => setEditBuilder(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy}
              onClick={() =>
                void saveConfig({ builder: draftBuilder }).then(() =>
                  setEditBuilder(false),
                )
              }
            >
              Save
            </Button>
          </>
        }
      >
        <SelectField
          label="Strategy"
          value={draftBuilder}
          onChange={(v) => setDraftBuilder(v as BuilderUi)}
          options={[
            { value: "railpack", label: "Railpack (default)" },
            { value: "dockerfile", label: "Dockerfile" },
            { value: "buildpacks", label: "Buildpacks" },
            { value: "nixpacks", label: "Nixpacks" },
          ]}
        />
        {(draftBuilder === "buildpacks" || draftBuilder === "nixpacks") && (
          <p
            className="mt-3 text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            {draftBuilder === "buildpacks"
              ? "Buildpacks run the Dockerfile pipeline on your repo root."
              : "Nixpacks is mapped to Railpack for builds."}
          </p>
        )}
      </Modal>

      <Modal
        open={editResources}
        title="Resources"
        onClose={() => setEditResources(false)}
        footer={
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => setEditResources(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy}
              onClick={() => {
                const port = Number.parseInt(draftPort, 10);
                const mem = Number.parseInt(draftMem, 10);
                if (!Number.isFinite(port) || port < 1 || port > 65535) {
                  onError("Port must be between 1 and 65535.");
                  return;
                }
                if (!Number.isFinite(mem) || mem < 128) {
                  onError("Memory must be at least 128 MB.");
                  return;
                }
                void saveConfig({ port, memLimitMB: mem }).then(() =>
                  setEditResources(false),
                );
              }}
            >
              Save
            </Button>
          </>
        }
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field
            label="Listen port"
            value={draftPort}
            onChange={(e) => setDraftPort(e.target.value)}
            inputMode="numeric"
          />
          <Field
            label="Memory limit (MB)"
            value={draftMem}
            onChange={(e) => setDraftMem(e.target.value)}
            inputMode="numeric"
            hint="Applied on the next deployment."
          />
        </div>
      </Modal>

      <Modal
        open={deployTokenOpen}
        title="Deploy token"
        description="Copy this token now. It will not be shown again."
        onClose={() => {
          setDeployTokenOpen(false);
          setDeployTokenPlain(null);
        }}
        footer={
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setDeployTokenOpen(false);
                setDeployTokenPlain(null);
              }}
            >
              Done
            </Button>
            {deployTokenPlain ? (
              <Button
                type="button"
                onClick={() =>
                  void copyToClipboard(deployTokenPlain, "Copied deploy token")
                }
              >
                Copy token
              </Button>
            ) : null}
          </>
        }
      >
        {deployTokenPlain ? (
          <SettingsCopyPill value={deployTokenPlain} />
        ) : null}
        <p className="mt-4 text-[13px]" style={{ color: "var(--muted-text)" }}>
          Use header <code className="font-mono">Authorization: Bearer …</code>{" "}
          when POSTing to the hook URL with{" "}
          <code className="font-mono">{'{ "image": "ghcr.io/..." }'}</code>.
        </p>
      </Modal>

      <Modal
        open={deleteOpen}
        title="Delete application"
        description={`This permanently deletes "${service.name}" and cannot be undone.`}
        onClose={() => setDeleteOpen(false)}
        footer={
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => setDeleteOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="danger"
              disabled={deleteBusy || confirmSlug !== service.slug}
              onClick={() => void confirmDelete()}
            >
              Delete
            </Button>
          </>
        }
      >
        {deleteError ? (
          <p className="mb-4 text-sm text-red-600" role="alert">
            {deleteError}
          </p>
        ) : null}
        <Field
          label={`Confirm slug (${service.slug})`}
          value={confirmSlug}
          onChange={(e) => setConfirmSlug(e.target.value)}
          autoFocus
        />
      </Modal>

      <Modal
        open={editHealth}
        title="Health check"
        onClose={() => setEditHealth(false)}
        footer={
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => setEditHealth(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy}
              onClick={() =>
                void saveConfig({ healthPath: draftHealth.trim() || "/" }).then(
                  () => setEditHealth(false),
                )
              }
            >
              Save
            </Button>
          </>
        }
      >
        <Field
          label="HTTP path"
          value={draftHealth}
          onChange={(e) => setDraftHealth(e.target.value)}
          placeholder="/health"
          hint="Used for zero-downtime deploy readiness probes."
        />
        <p className="mt-3 text-[13px]">
          <Link
            href={`/projects/${project.slug}`}
            className="underline"
            style={{ color: "var(--link-blue)" }}
          >
            Project settings
          </Link>{" "}
          may override registry and Git credentials.
        </p>
      </Modal>
    </div>
  );
}
