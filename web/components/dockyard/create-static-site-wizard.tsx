"use client";

import { GitBranch, Trash2 } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import type { StaticSiteCreateResult } from "@/components/dockyard/static-site-create-progress-modal";
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
  createDeployment,
  createService,
  type GitHubStatus,
  getGitHubStatus,
  getInstanceSettings,
  type InstanceSettings,
  listGitHubRepos,
  listProjects,
  type Project,
} from "@/lib/api";

const STEPS: WizardStepDef[] = [
  { id: "source", label: "Source" },
  { id: "name", label: "Name & project" },
  { id: "build", label: "Build" },
  { id: "environment", label: "Environment" },
  { id: "domain", label: "Domain" },
];

type FrameworkPreset = "astro" | "vite" | "next" | "hugo" | "plain";

const FRAMEWORKS: {
  id: FrameworkPreset;
  label: string;
  install: string;
  build: string;
  output: string;
  hint?: string;
}[] = [
  {
    id: "astro",
    label: "Astro",
    install: "npm install",
    build: "npm run build",
    output: "dist",
  },
  {
    id: "vite",
    label: "Vite",
    install: "npm install",
    build: "npm run build",
    output: "dist",
  },
  {
    id: "next",
    label: "Next.js (export)",
    install: "npm install",
    build: "npm run build",
    output: "out",
  },
  {
    id: "hugo",
    label: "Hugo",
    install: "",
    build: "hugo",
    output: "public",
  },
  {
    id: "plain",
    label: "Plain HTML",
    install: "",
    build: "true",
    output: ".",
  },
];

type EnvRow = { id: string; key: string; value: string };

type Draft = {
  repo: string;
  branch: string;
  rootDir: string;
  siteName: string;
  projectId: string;
  environment: string;
  framework: FrameworkPreset;
  installCmd: string;
  buildCmd: string;
  outputDir: string;
  nodeVersion: string;
  envVars: EnvRow[];
  customDomain: string;
  autoHttps: boolean;
  deployOnPush: boolean;
  spaFallback: boolean;
};

function slugifyName(name: string): string {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 48);
}

function defaultDeployUrl(
  siteName: string,
  instance: InstanceSettings | null,
): string {
  const slug = slugifyName(siteName) || "site";
  const host =
    instance?.wildcard_domain?.trim() ||
    (instance?.public_url
      ? (() => {
          try {
            return new URL(instance.public_url).hostname;
          } catch {
            return "your-server.example";
          }
        })()
      : "your-server.example");
  return `${slug}.deploy.${host}`;
}

function combineBuildCmd(install: string, build: string): string {
  const i = install.trim();
  const b = build.trim();
  if (i && b) return `${i} && ${b}`;
  return b || i || "npm ci && npm run build";
}

function newEnvRow(): EnvRow {
  return {
    id:
      typeof crypto !== "undefined" && crypto.randomUUID
        ? crypto.randomUUID()
        : `env-${Date.now()}`,
    key: "",
    value: "",
  };
}

type CreateStaticSiteWizardProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (result: StaticSiteCreateResult) => void;
};

export function CreateStaticSiteWizard({
  open,
  onClose,
  onCreated,
}: CreateStaticSiteWizardProps) {
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
    repo: "",
    branch: "main",
    rootDir: "/",
    siteName: "",
    projectId: "",
    environment: "production",
    framework: "astro",
    installCmd: "npm install",
    buildCmd: "npm run build",
    outputDir: "dist",
    nodeVersion: "22",
    envVars: [],
    customDomain: "",
    autoHttps: true,
    deployOnPush: true,
    spaFallback: false,
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
    void loadMeta();
  }, [open, loadMeta]);

  const defaultUrl = useMemo(
    () => defaultDeployUrl(draft.siteName || "marketing", instance),
    [draft.siteName, instance],
  );

  function applyFramework(id: FrameworkPreset) {
    const fw = FRAMEWORKS.find((f) => f.id === id) ?? FRAMEWORKS[0];
    patch({
      framework: id,
      installCmd: fw.install,
      buildCmd: fw.build,
      outputDir: fw.output,
    });
  }

  function validateStep(): string | null {
    if (step === 1) {
      if (!github?.linked)
        return "Connect GitHub in Settings → Integrations first.";
      if (!draft.repo.trim()) return "Pick a repository.";
      if (!draft.branch.trim()) return "Branch is required.";
    }
    if (step === 2) {
      if (!draft.siteName.trim()) return "Site name is required.";
      if (!draft.projectId) return "Pick a project.";
    }
    if (step === 3) {
      if (!draft.buildCmd.trim()) return "Build command is required.";
      if (!draft.outputDir.trim()) return "Output directory is required.";
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
    try {
      const rootDir = draft.rootDir.trim() === "/" ? "" : draft.rootDir.trim();
      const nodeImage =
        draft.nodeVersion === "20" ? "node:20-alpine" : "node:22-alpine";
      const service = await createService({
        project_id: draft.projectId,
        environment: draft.environment,
        kind: "static_site",
        name: draft.siteName.trim(),
        config: {
          source: {
            type: "github",
            repo: draft.repo.trim(),
            branch: draft.branch.trim(),
            rootDir,
          },
          autoDeploy: draft.deployOnPush,
          buildCmd: combineBuildCmd(draft.installCmd, draft.buildCmd),
          outputDir: draft.outputDir.trim(),
          nodeImage,
        },
      });
      const deployment = await createDeployment(service.id);
      const url =
        draft.customDomain.trim() ||
        defaultDeployUrl(draft.siteName.trim(), instance);
      onCreated({
        serviceId: service.id,
        deploymentId: deployment.id,
        siteName: draft.siteName.trim(),
        repo: draft.repo.trim(),
        branch: draft.branch.trim(),
        installCmd: draft.installCmd.trim() || "npm install",
        buildCmd: draft.buildCmd.trim() || "npm run build",
        outputDir: draft.outputDir.trim() || "dist",
        defaultUrl: url,
      });
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not create static site");
    } finally {
      setBusy(false);
    }
  }

  const nextLabel = step === STEPS.length ? "Create & deploy" : "Next";

  return (
    <WizardModalShell
      open={open}
      title="Create static site"
      subtitle="Five short steps, then deploy."
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
          onRepoChange={(repo) => {
            const match = repos.find((r) => r.full_name === repo);
            patch({
              repo,
              branch: match?.default_branch || draft.branch,
            });
          }}
        />
      ) : null}
      {step === 2 ? (
        <StepNameProject projects={projects} draft={draft} onChange={patch} />
      ) : null}
      {step === 3 ? (
        <StepBuild
          draft={draft}
          onChange={patch}
          onFramework={applyFramework}
        />
      ) : null}
      {step === 4 ? <StepEnvironment draft={draft} onChange={patch} /> : null}
      {step === 5 ? (
        <StepDomain draft={draft} defaultUrl={defaultUrl} onChange={patch} />
      ) : null}
    </WizardModalShell>
  );
}

function StepHeading({ title, hint }: { title: string; hint: string }) {
  return (
    <div className="mb-6">
      <h3 className="text-lg font-semibold tracking-[-0.02em]">{title}</h3>
      <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
        {hint}
      </p>
    </div>
  );
}

function StepSource({
  github,
  repos,
  draft,
  onChange,
  onRepoChange,
}: {
  github: GitHubStatus | null;
  repos: { full_name: string; default_branch: string }[];
  draft: Draft;
  onChange: (p: Partial<Draft>) => void;
  onRepoChange: (repo: string) => void;
}) {
  const account = github?.slug ?? github?.installed_on ?? "GitHub";
  const repoCount = github?.repository_count ?? repos.length;

  return (
    <>
      <StepHeading title="Source" hint="Where your code lives." />
      <div
        className="mb-6 flex flex-wrap items-center justify-between gap-3 rounded-[16px] border px-4 py-3"
        style={{ borderColor: "var(--border)", background: "#f4f8ff" }}
      >
        <div className="flex items-center gap-3">
          <span
            className="flex size-10 items-center justify-center rounded-xl"
            style={{ background: "#e3f0ff", color: "#0b3d91" }}
          >
            <GitBranch className="size-5" strokeWidth={1.8} aria-hidden />
          </span>
          <div>
            <p className="text-sm font-semibold">GitHub · {account}</p>
            <p className="text-[13px]" style={{ color: "var(--muted-text)" }}>
              {repoCount} repositories available
            </p>
          </div>
        </div>
        <Link
          href="/settings/server/integrations"
          className="text-sm font-medium hover:underline"
          style={{ color: "var(--link-blue)" }}
        >
          Change provider
        </Link>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <SearchSelectField
          label="Repository"
          placeholder="Choose a repository"
          searchPlaceholder="Search repositories…"
          value={draft.repo}
          onChange={onRepoChange}
          hint="Type to search your repositories."
          options={repos.map((r) => ({
            value: r.full_name,
            label: r.full_name,
          }))}
        />
        <Field
          label="Branch"
          value={draft.branch}
          onChange={(e) => onChange({ branch: e.target.value })}
          hint="Pushes to this branch deploy."
        />
      </div>
      <div className="mt-4">
        <Field
          label="Root directory"
          value={draft.rootDir}
          onChange={(e) => onChange({ rootDir: e.target.value })}
          hint="Only change this for a monorepo, e.g. apps/web."
        />
      </div>
    </>
  );
}

function StepNameProject({
  projects,
  draft,
  onChange,
}: {
  projects: Project[];
  draft: Draft;
  onChange: (p: Partial<Draft>) => void;
}) {
  return (
    <>
      <StepHeading title="Name & project" hint="Where this site shows up." />
      <div className="grid gap-4 md:grid-cols-3">
        <Field
          label="Site name"
          value={draft.siteName}
          onChange={(e) => onChange({ siteName: e.target.value })}
          placeholder="marketing"
          hint="Used in the default URL and across Dockyard."
        />
        <SelectField
          label="Project"
          placeholder="Select project"
          value={draft.projectId}
          onChange={(v) => onChange({ projectId: v })}
          options={projects.map((p) => ({ value: p.id, label: p.name }))}
        />
        <SelectField
          label="Environment"
          value={draft.environment}
          onChange={(v) => onChange({ environment: v })}
          options={[
            { value: "production", label: "Production" },
            { value: "staging", label: "Staging" },
          ]}
        />
      </div>
    </>
  );
}

function StepBuild({
  draft,
  onChange,
  onFramework,
}: {
  draft: Draft;
  onChange: (p: Partial<Draft>) => void;
  onFramework: (id: FrameworkPreset) => void;
}) {
  return (
    <>
      <StepHeading
        title="Build settings"
        hint="How Dockyard turns your code into a website."
      />
      <div className="mb-4 flex flex-wrap gap-2">
        {FRAMEWORKS.map((fw) => {
          const active = draft.framework === fw.id;
          return (
            <button
              key={fw.id}
              type="button"
              onClick={() => onFramework(fw.id)}
              className="rounded-md border px-3 py-1.5 text-sm font-medium transition-colors"
              style={
                active
                  ? {
                      borderColor: "var(--link-blue)",
                      color: "var(--link-blue)",
                      background: "#f4f8ff",
                    }
                  : {
                      borderColor: "var(--border)",
                      color: "var(--ink-navy)",
                      background: "var(--surface)",
                    }
              }
            >
              {fw.label}
            </button>
          );
        })}
      </div>
      <p className="mb-6 text-[13px]" style={{ color: "var(--muted-text)" }}>
        We detected {FRAMEWORKS.find((f) => f.id === draft.framework)?.label}{" "}
        from your repository. Adjust anything below.
      </p>
      <div className="grid gap-4 md:grid-cols-2">
        <Field
          label="Install command"
          value={draft.installCmd}
          onChange={(e) => onChange({ installCmd: e.target.value })}
        />
        <Field
          label="Build command"
          value={draft.buildCmd}
          onChange={(e) => onChange({ buildCmd: e.target.value })}
        />
        <Field
          label="Output directory"
          value={draft.outputDir}
          onChange={(e) => onChange({ outputDir: e.target.value })}
          hint="The folder that gets served."
        />
        <SelectField
          label="Node version"
          value={draft.nodeVersion}
          onChange={(v) => onChange({ nodeVersion: v })}
          options={[
            { value: "22", label: "22 LTS" },
            { value: "20", label: "20 LTS" },
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
        hint="Optional. Available while the site is built."
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
      <p className="mt-4 text-[13px]" style={{ color: "var(--muted-text)" }}>
        Variables are available at build time. Values are stored encrypted.
      </p>
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
      <StepHeading title="Domain & behavior" hint="How people reach it." />
      <div className="grid gap-4 md:grid-cols-2">
        <Field
          label="Default URL"
          value={defaultUrl}
          readOnly
          hint="Always works, no DNS needed."
        />
        <Field
          label="Custom domain (optional)"
          value={draft.customDomain}
          onChange={(e) => onChange({ customDomain: e.target.value })}
          placeholder="clientx.nl"
          hint="Point an A record to your server IP afterwards."
        />
      </div>
      <div className="mt-8 space-y-6">
        <div>
          <Switch
            label="Automatic HTTPS"
            checked={draft.autoHttps}
            onChange={(v) => onChange({ autoHttps: v })}
          />
          <p
            className="mt-1 pl-14 text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            Free certificate from Let's Encrypt, renewed automatically.
          </p>
        </div>
        <div>
          <Switch
            label="Deploy on push"
            checked={draft.deployOnPush}
            onChange={(v) => onChange({ deployOnPush: v })}
          />
          <p
            className="mt-1 pl-14 text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            Every push to main builds and publishes a new version.
          </p>
        </div>
        <div>
          <Switch
            label="Single-page app fallback"
            checked={draft.spaFallback}
            onChange={(v) => onChange({ spaFallback: v })}
          />
          <p
            className="mt-1 pl-14 text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            Serve index.html for unknown paths.
          </p>
        </div>
      </div>
    </>
  );
}
