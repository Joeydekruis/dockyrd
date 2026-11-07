"use client";

import { Container, Trash2 } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { GitHubBranchSelectField } from "@/components/dockyard/github-branch-select-field";
import type { ResourceProvisionContext } from "@/components/dockyard/resource-provisioning-modal";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import { SearchSelectField } from "@/components/dockyard/ui/search-select-field";
import { SelectField } from "@/components/dockyard/ui/select-field";
import { Switch } from "@/components/dockyard/ui/switch";
import {
  WizardModalShell,
  type WizardStepDef,
} from "@/components/dockyard/ui/wizard-modal-shell";
import {
  createService,
  createServiceDomain,
  type GitHubStatus,
  getGitHubStatus,
  getInstanceSettings,
  type InstanceSettings,
  listGitHubRepos,
  listProjects,
  type Project,
  putServiceVariables,
} from "@/lib/api";
import { previewDefaultPlatformHostname } from "@/lib/domain-dns";

const STEPS: WizardStepDef[] = [
  { id: "source", label: "Source" },
  { id: "name", label: "Name & project" },
  { id: "build", label: "Build" },
  { id: "resources", label: "Resources" },
  { id: "environment", label: "Environment" },
  { id: "domain", label: "Domain" },
];

type SourceKind = "github" | "image";
type BuilderChoice = "railpack" | "dockerfile" | "buildpacks" | "nixpacks";

type EnvRow = { id: string; key: string; value: string; secret: boolean };

type Draft = {
  sourceKind: SourceKind;
  repo: string;
  branch: string;
  imageRef: string;
  appName: string;
  projectId: string;
  builder: BuilderChoice;
  port: string;
  memPreset: string;
  envVars: EnvRow[];
  customDomain: string;
  autoHttps: boolean;
};

function slugifyName(name: string): string {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 48);
}

function newEnvRow(): EnvRow {
  return {
    id:
      typeof crypto !== "undefined" && crypto.randomUUID
        ? crypto.randomUUID()
        : `env-${Date.now()}`,
    key: "",
    value: "",
    secret: false,
  };
}

type CreateApplicationWizardProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (context: ResourceProvisionContext) => void;
  defaultProjectId?: string;
};

export function CreateApplicationWizard({
  open,
  onClose,
  onCreated,
  defaultProjectId,
}: CreateApplicationWizardProps) {
  const [step, setStep] = useState(1);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [repos, setRepos] = useState<
    { full_name: string; default_branch: string }[]
  >([]);
  const [github, setGithub] = useState<GitHubStatus | null>(null);
  const [instance, setInstance] = useState<InstanceSettings | null>(null);
  const [draft, setDraft] = useState<Draft>(() => ({
    sourceKind: "github",
    repo: "",
    branch: "main",
    imageRef: "nginx:alpine",
    appName: "",
    projectId: "",
    builder: "railpack",
    port: "3000",
    memPreset: "512",
    envVars: [],
    customDomain: "",
    autoHttps: true,
  }));

  const patch = (p: Partial<Draft>) => setDraft((d) => ({ ...d, ...p }));

  const loadMeta = useCallback(async () => {
    const [ps, gh, inst] = await Promise.all([
      listProjects(),
      getGitHubStatus().catch(() => null),
      getInstanceSettings().catch(() => null),
    ]);
    setProjects(ps.projects);
    setGithub(gh);
    setInstance(inst);
    if (gh?.linked) {
      const { repos: r } = await listGitHubRepos();
      setRepos(r);
      if (r.length > 0) {
        setDraft((d) => ({
          ...d,
          repo: d.repo || r[0].full_name,
          branch: d.repo ? d.branch : r[0].default_branch || "main",
        }));
      }
    }
  }, []);

  useEffect(() => {
    if (!open) return;
    setStep(1);
    setError(null);
    if (defaultProjectId) {
      setDraft((d) => ({ ...d, projectId: defaultProjectId }));
    }
    void loadMeta();
  }, [open, loadMeta, defaultProjectId]);

  const defaultUrl = useMemo(
    () =>
      previewDefaultPlatformHostname(
        slugifyName(draft.appName || "app") || "app",
        instance,
      ),
    [draft.appName, instance],
  );

  function validateStep(): string | null {
    if (step === 1) {
      if (draft.sourceKind === "github") {
        if (!github?.linked)
          return "Connect GitHub in Settings → Integrations first.";
        if (!draft.repo.trim()) return "Pick a repository.";
      } else if (!draft.imageRef.trim()) {
        return "Image reference is required.";
      }
    }
    if (step === 2) {
      if (!draft.appName.trim()) return "Application name is required.";
      if (!draft.projectId) return "Pick a project.";
    }
    if (step === 4) {
      const p = Number.parseInt(draft.port, 10);
      if (!Number.isFinite(p) || p <= 0) return "Enter a valid port.";
    }
    return null;
  }

  function goNext() {
    const msg = validateStep();
    if (msg) {
      setError(msg);
      return;
    }
    setError(null);
    if (step < STEPS.length) {
      setStep((s) => s + 1);
      return;
    }
    void submit();
  }

  function goBack() {
    setError(null);
    setStep((s) => Math.max(1, s - 1));
  }

  async function submit() {
    setBusy(true);
    setError(null);
    const port = Number.parseInt(draft.port, 10) || 3000;
    const memLimitMB = Number.parseInt(draft.memPreset, 10) || 512;
    const source =
      draft.sourceKind === "github"
        ? {
            type: "github",
            repo: draft.repo.trim(),
            branch: draft.branch.trim() || "main",
          }
        : { type: "image", ref: draft.imageRef.trim() };

    try {
      const service = await createService({
        project_id: draft.projectId,
        kind: "app",
        name: draft.appName.trim(),
        config: {
          source,
          builder: draft.builder,
          port,
          memLimitMB,
          autoDeploy: false,
          healthPath: "/",
        },
      });

      const envPayload = draft.envVars
        .map((r) => ({
          key: r.key.trim(),
          value: r.value,
          is_secret: r.secret,
          at_build: false,
          at_runtime: true,
        }))
        .filter((v) => v.key.length > 0);
      if (envPayload.length > 0) {
        await putServiceVariables(service.id, envPayload);
      }

      const host = draft.customDomain.trim();
      if (host) {
        await createServiceDomain(service.id, {
          hostname: host,
          primary: true,
        });
      }

      onCreated({
        kind: "application",
        serviceId: service.id,
        name: draft.appName.trim(),
        label:
          draft.sourceKind === "image"
            ? "Pulling image and starting the container."
            : "Building from repository and starting the container.",
      });
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not create application");
    } finally {
      setBusy(false);
    }
  }

  const nextLabel = step === STEPS.length ? "Create application" : "Next";

  return (
    <WizardModalShell
      open={open}
      title="Create application"
      subtitle="Six steps — source, build, resources, and optional domain."
      steps={STEPS}
      step={step}
      onClose={onClose}
      onBack={step > 1 ? goBack : undefined}
      onNext={goNext}
      nextLabel={nextLabel}
      busy={busy}
      error={error}
    >
      {step === 1 ? (
        <StepSource
          github={github}
          repos={repos}
          draft={draft}
          onChange={patch}
        />
      ) : null}
      {step === 2 ? (
        <StepNameProject
          projects={projects}
          draft={draft}
          hideProject={Boolean(defaultProjectId)}
          onChange={patch}
        />
      ) : null}
      {step === 3 ? <StepBuild draft={draft} onChange={patch} /> : null}
      {step === 4 ? <StepResources draft={draft} onChange={patch} /> : null}
      {step === 5 ? <StepEnvironment draft={draft} onChange={patch} /> : null}
      {step === 6 ? (
        <StepDomain draft={draft} defaultUrl={defaultUrl} onChange={patch} />
      ) : null}
    </WizardModalShell>
  );
}

function StepHeading({ title, hint }: { title: string; hint?: string }) {
  return (
    <div className="mb-6">
      <h3 className="text-lg font-semibold tracking-tight">{title}</h3>
      {hint ? (
        <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
          {hint}
        </p>
      ) : null}
    </div>
  );
}

function StepSource({
  github,
  repos,
  draft,
  onChange,
}: {
  github: GitHubStatus | null;
  repos: { full_name: string; default_branch: string }[];
  draft: Draft;
  onChange: (p: Partial<Draft>) => void;
}) {
  return (
    <>
      <StepHeading
        title="Source"
        hint="Deploy from a GitHub repository or a container image."
      />
      <div className="mb-6 flex flex-wrap gap-2">
        <Button
          type="button"
          variant={draft.sourceKind === "github" ? "primary" : "outline"}
          onClick={() => onChange({ sourceKind: "github" })}
        >
          GitHub repository
        </Button>
        <Button
          type="button"
          variant={draft.sourceKind === "image" ? "primary" : "outline"}
          onClick={() => onChange({ sourceKind: "image" })}
        >
          <Container className="mr-1.5 size-4" aria-hidden />
          Container image
        </Button>
      </div>
      {draft.sourceKind === "github" ? (
        github?.linked ? (
          <div className="grid gap-4 md:grid-cols-2">
            <SearchSelectField
              label="Repository"
              value={draft.repo}
              onChange={(v) => {
                const match = repos.find((r) => r.full_name === v);
                onChange({
                  repo: v,
                  branch: match?.default_branch || draft.branch,
                });
              }}
              options={repos.map((r) => ({
                value: r.full_name,
                label: r.full_name,
              }))}
            />
            <GitHubBranchSelectField
              repo={draft.repo}
              value={draft.branch}
              preferredBranch={
                repos.find((r) => r.full_name === draft.repo)?.default_branch
              }
              onChange={(branch) => onChange({ branch })}
              hint="Deploys run on pushes to this branch."
            />
          </div>
        ) : (
          <p className="text-sm">
            <Link
              href="/settings/integrations"
              className="underline"
              style={{ color: "var(--link-blue)" }}
            >
              Connect GitHub
            </Link>{" "}
            to pick a repository.
          </p>
        )
      ) : (
        <Field
          label="Image reference"
          value={draft.imageRef}
          onChange={(e) => onChange({ imageRef: e.target.value })}
          placeholder="ghcr.io/org/app:latest"
        />
      )}
    </>
  );
}

function StepNameProject({
  projects,
  draft,
  hideProject,
  onChange,
}: {
  projects: Project[];
  draft: Draft;
  hideProject?: boolean;
  onChange: (p: Partial<Draft>) => void;
}) {
  return (
    <>
      <StepHeading
        title="Name & project"
        hint="Shown in the sidebar and URLs."
      />
      <div className="grid gap-4 md:grid-cols-2">
        <Field
          label="Application name"
          value={draft.appName}
          onChange={(e) => onChange({ appName: e.target.value })}
          placeholder="web-api"
        />
        {hideProject ? null : (
          <SelectField
            label="Project"
            placeholder="Select project"
            value={draft.projectId}
            onChange={(v) => onChange({ projectId: v })}
            options={projects.map((p) => ({ value: p.id, label: p.name }))}
          />
        )}
      </div>
    </>
  );
}

const BUILDERS: {
  id: BuilderChoice;
  label: string;
  tag?: string;
  desc: string;
}[] = [
  {
    id: "railpack",
    label: "Railpack",
    tag: "Default",
    desc: "Detects language and framework without a Dockerfile.",
  },
  {
    id: "dockerfile",
    label: "Dockerfile",
    desc: "Build using a Dockerfile in the repository.",
  },
  {
    id: "buildpacks",
    label: "Buildpacks",
    desc: "Cloud Native Buildpacks with a Procfile.",
  },
  {
    id: "nixpacks",
    label: "Nixpacks",
    tag: "Deprecated",
    desc: "Uses the Railpack build pipeline.",
  },
];

function StepBuild({
  draft,
  onChange,
}: {
  draft: Draft;
  onChange: (p: Partial<Draft>) => void;
}) {
  if (draft.sourceKind === "image") {
    return (
      <>
        <StepHeading
          title="Build"
          hint="Image-based apps skip the build step."
        />
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          dockyrd pulls <code className="font-mono">{draft.imageRef}</code>{" "}
          directly.
        </p>
      </>
    );
  }
  return (
    <>
      <StepHeading
        title="Build strategy"
        hint="How the repository becomes a container image."
      />
      <ul className="space-y-2">
        {BUILDERS.map((b) => (
          <li key={b.id}>
            <button
              type="button"
              className="w-full rounded-2xl border px-4 py-3 text-left transition-colors"
              style={{
                borderColor:
                  draft.builder === b.id ? "var(--link-blue)" : "var(--border)",
                background:
                  draft.builder === b.id
                    ? "var(--hover-row)"
                    : "var(--surface)",
              }}
              onClick={() => onChange({ builder: b.id })}
            >
              <div className="flex items-center gap-2">
                <span className="font-medium">{b.label}</span>
                {b.tag ? (
                  <span
                    className="rounded-full px-2 py-0.5 font-mono text-[10px] uppercase"
                    style={{ background: "var(--hover-row)" }}
                  >
                    {b.tag}
                  </span>
                ) : null}
              </div>
              <p
                className="mt-1 text-[13px]"
                style={{ color: "var(--muted-text)" }}
              >
                {b.desc}
              </p>
            </button>
          </li>
        ))}
      </ul>
    </>
  );
}

function StepResources({
  draft,
  onChange,
}: {
  draft: Draft;
  onChange: (p: Partial<Draft>) => void;
}) {
  return (
    <>
      <StepHeading
        title="Resources"
        hint="Port and memory for the web process."
      />
      <div className="grid gap-4 md:grid-cols-2">
        <Field
          label="Container port"
          value={draft.port}
          onChange={(e) => onChange({ port: e.target.value })}
          placeholder="3000"
        />
        <SelectField
          label="Instance size"
          value={draft.memPreset}
          onChange={(v) => onChange({ memPreset: v })}
          options={[
            { value: "256", label: "256 MB" },
            { value: "512", label: "512 MB (default)" },
            { value: "1024", label: "1 GB" },
            { value: "2048", label: "2 GB" },
          ]}
        />
      </div>
    </>
  );
}

function StepEnvironment({
  draft,
  onChange,
}: {
  draft: Draft;
  onChange: (p: Partial<Draft>) => void;
}) {
  return (
    <>
      <StepHeading
        title="Environment variables"
        hint="Optional. Injected at runtime on the next deployment."
      />
      <div className="space-y-3">
        {draft.envVars.map((row) => (
          <div key={row.id} className="flex flex-wrap items-end gap-2">
            <Field
              label="Key"
              className="min-w-[140px] flex-1"
              value={row.key}
              onChange={(e) =>
                onChange({
                  envVars: draft.envVars.map((r) =>
                    r.id === row.id ? { ...r, key: e.target.value } : r,
                  ),
                })
              }
            />
            <Field
              label="Value"
              className="min-w-[180px] flex-[2]"
              value={row.value}
              onChange={(e) =>
                onChange({
                  envVars: draft.envVars.map((r) =>
                    r.id === row.id ? { ...r, value: e.target.value } : r,
                  ),
                })
              }
            />
            <IconButton
              icon={Trash2}
              label="Remove variable"
              onClick={() =>
                onChange({
                  envVars: draft.envVars.filter((r) => r.id !== row.id),
                })
              }
            />
          </div>
        ))}
      </div>
      <Button
        type="button"
        variant="outline"
        className="mt-4"
        onClick={() => onChange({ envVars: [...draft.envVars, newEnvRow()] })}
      >
        Add variable
      </Button>
    </>
  );
}

function StepDomain({
  draft,
  defaultUrl,
  onChange,
}: {
  draft: Draft;
  defaultUrl: string;
  onChange: (p: Partial<Draft>) => void;
}) {
  return (
    <>
      <StepHeading title="Domain" hint="Optional custom hostname." />
      <div className="grid gap-4 md:grid-cols-2">
        <Field label="Default URL" value={defaultUrl} readOnly />
        <Field
          label="Custom domain (optional)"
          value={draft.customDomain}
          onChange={(e) => onChange({ customDomain: e.target.value })}
          placeholder="api.example.com"
        />
      </div>
      <div className="mt-6">
        <Switch
          label="Automatic HTTPS"
          checked={draft.autoHttps}
          onChange={(v) => onChange({ autoHttps: v })}
        />
      </div>
    </>
  );
}
