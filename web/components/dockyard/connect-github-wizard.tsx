"use client";

import { Check, ExternalLink, GitBranch, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import {
  type GitHubStatus,
  getGitHubStatus,
  startGitHubAppLink,
} from "@/lib/api";
import { openExternalUrl } from "@/lib/open-external-url";

const WIZARD_STORAGE_KEY = "dockyrd.githubWizard.v1";

type WizardDraft = {
  ownerKind: "personal" | "organization";
  appName: string;
  instanceUrl: string;
};

type Step = 1 | 2 | 3 | 4;

const STEPS: { n: Step; title: string; hint: string }[] = [
  { n: 1, title: "Prepare", hint: "Name and owner" },
  { n: 2, title: "Create", hint: "On GitHub" },
  { n: 3, title: "Install", hint: "Pick repositories" },
  { n: 4, title: "Done", hint: "Ready to deploy" },
];

const PERMISSIONS: { name: string; level: string }[] = [
  { name: "Contents", level: "Read" },
  { name: "Metadata", level: "Read" },
  { name: "Pull requests", level: "Read" },
  { name: "Webhook events", level: "Push" },
];

function suggestAppName(instanceUrl: string, accountHint: string) {
  let host = "dockyrd";
  try {
    const u = new URL(instanceUrl);
    host = u.hostname.split(".")[0] || host;
  } catch {
    /* keep default */
  }
  const suffix = accountHint
    .toLowerCase()
    .replace(/[^a-z0-9-]+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 30);
  const base = suffix ? `dockyrd-${suffix}` : `dockyrd-${host}`;
  return base.slice(0, 34);
}

function loadDraft(): WizardDraft | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = sessionStorage.getItem(WIZARD_STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as WizardDraft;
  } catch {
    return null;
  }
}

function saveDraft(draft: WizardDraft) {
  sessionStorage.setItem(WIZARD_STORAGE_KEY, JSON.stringify(draft));
}

type ConnectGitHubWizardProps = {
  open: boolean;
  initialStep?: Step;
  instanceUrl: string;
  accountHint: string;
  onClose: () => void;
  onComplete: () => void;
};

export function ConnectGitHubWizard({
  open,
  initialStep = 1,
  instanceUrl,
  accountHint,
  onClose,
  onComplete,
}: ConnectGitHubWizardProps) {
  const [step, setStep] = useState<Step>(initialStep);
  const [draft, setDraft] = useState<WizardDraft>(() => ({
    ownerKind: "personal",
    appName: suggestAppName(instanceUrl, accountHint),
    instanceUrl,
  }));
  const [status, setStatus] = useState<GitHubStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const refreshStatus = useCallback(async () => {
    const st = await getGitHubStatus();
    setStatus(st);
    return st;
  }, []);

  useEffect(() => {
    if (!open) return;
    setStep(initialStep);
    const stored = loadDraft();
    setDraft(
      stored ?? {
        ownerKind: "personal",
        appName: suggestAppName(instanceUrl, accountHint),
        instanceUrl: instanceUrl.trim(),
      },
    );
    setError(null);
    void refreshStatus();
  }, [open, initialStep, instanceUrl, accountHint, refreshStatus]);

  useEffect(() => {
    if (!open) return;
    setDraft((d) => ({
      ...d,
      instanceUrl: instanceUrl.trim() || d.instanceUrl,
    }));
  }, [instanceUrl, open]);

  const updateDraft = (patch: Partial<WizardDraft>) => {
    setDraft((prev) => {
      const next = { ...prev, ...patch };
      saveDraft(next);
      return next;
    });
  };

  async function openCreateOnGitHub() {
    setBusy(true);
    setError(null);
    try {
      const { url } = await startGitHubAppLink({
        app_name: draft.appName.trim(),
        public_url: draft.instanceUrl.trim(),
      });
      openExternalUrl(url);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not start GitHub setup");
    } finally {
      setBusy(false);
    }
  }

  async function confirmAppCreated() {
    setBusy(true);
    setError(null);
    try {
      const st = await refreshStatus();
      if (!st.app_configured && !st.linked) {
        setError(
          "GitHub app not detected yet. Finish creating the app on GitHub, then try again.",
        );
        return;
      }
      setStep(3);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not verify GitHub app");
    } finally {
      setBusy(false);
    }
  }

  async function openInstallOnGitHub() {
    const st = status ?? (await refreshStatus());
    const url = st.install_url;
    if (!url) {
      setError(
        "Install link is not ready yet. Confirm the app was created first.",
      );
      return;
    }
    openExternalUrl(url);
  }

  async function confirmInstalled() {
    setBusy(true);
    setError(null);
    try {
      const st = await refreshStatus();
      if (!st.installation_id) {
        setError(
          "Installation not detected yet. Install the app on GitHub and select repositories, then try again.",
        );
        return;
      }
      setStep(4);
    } catch (e) {
      setError(
        e instanceof Error ? e.message : "Could not verify installation",
      );
    } finally {
      setBusy(false);
    }
  }

  function finish() {
    sessionStorage.removeItem(WIZARD_STORAGE_KEY);
    onComplete();
    onClose();
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        className="absolute inset-0 border-0 bg-black/40"
        aria-label="Close dialog"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="connect-github-title"
        className="relative z-10 flex max-h-[90vh] w-full max-w-[720px] flex-col overflow-hidden rounded-[20px] border shadow-lg"
        style={{ background: "var(--surface)", borderColor: "var(--border)" }}
      >
        <header className="flex items-start gap-3 border-b border-[var(--border)] px-6 py-5">
          <span
            className="flex size-10 shrink-0 items-center justify-center rounded-xl"
            style={{ background: "#e3f0ff", color: "#0b3d91" }}
          >
            <GitBranch strokeWidth={1.8} className="size-5" aria-hidden />
          </span>
          <div className="min-w-0 flex-1">
            <h2 id="connect-github-title" className="text-lg font-semibold">
              Connect GitHub
            </h2>
            <p className="text-sm text-[var(--muted-foreground)]">
              Set up a GitHub App in about two minutes
            </p>
          </div>
          <button
            type="button"
            className="btn-ghost rounded-lg p-2"
            aria-label="Close"
            onClick={onClose}
          >
            <X className="size-5" strokeWidth={1.8} />
          </button>
        </header>

        <div className="flex min-h-0 flex-1 flex-col sm:flex-row">
          <nav
            aria-label="Setup progress"
            className="shrink-0 border-b border-[var(--border)] px-6 py-4 sm:w-[200px] sm:border-r sm:border-b-0"
          >
            <ol className="space-y-3">
              {STEPS.map(({ n, title, hint }) => {
                const done = n < step;
                const active = n === step;
                return (
                  <li key={n} className="flex gap-3">
                    <span
                      className="flex size-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold"
                      style={
                        done
                          ? { background: "#e6f4ec", color: "#176b3d" }
                          : active
                            ? { background: "#102a43", color: "#fff" }
                            : {
                                background: "var(--muted)",
                                color: "var(--muted-foreground)",
                              }
                      }
                    >
                      {done ? (
                        <Check className="size-4" strokeWidth={2.5} />
                      ) : (
                        n
                      )}
                    </span>
                    <div>
                      <p
                        className="text-sm font-medium"
                        style={
                          active
                            ? { color: "var(--foreground)" }
                            : { color: "var(--muted-foreground)" }
                        }
                      >
                        {title}
                      </p>
                      <p className="text-xs text-[var(--muted-foreground)]">
                        {hint}
                      </p>
                    </div>
                  </li>
                );
              })}
            </ol>
          </nav>

          <div className="min-h-0 flex-1 overflow-y-auto px-6 py-5">
            {error ? (
              <p className="mb-4 text-sm text-red-600" role="alert">
                {error}
              </p>
            ) : null}

            {step === 1 ? (
              <StepPrepare
                draft={draft}
                accountHint={accountHint}
                onChange={updateDraft}
              />
            ) : null}
            {step === 2 ? (
              <StepCreate
                appName={draft.appName}
                onOpenGitHub={() => void openCreateOnGitHub()}
              />
            ) : null}
            {step === 3 ? (
              <StepInstall
                slug={status?.slug}
                onInstall={() => void openInstallOnGitHub()}
              />
            ) : null}
            {step === 4 ? <StepDone status={status} /> : null}
          </div>
        </div>

        <footer className="flex flex-wrap items-center justify-between gap-3 border-t border-[var(--border)] px-6 py-4">
          <p className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-foreground)]">
            Step {step} of 4
          </p>
          <div className="flex flex-wrap gap-2">
            {step > 1 && step < 4 ? (
              <Button
                type="button"
                variant="outline"
                onClick={() => setStep((s) => (s > 1 ? ((s - 1) as Step) : s))}
              >
                Back
              </Button>
            ) : null}
            {step === 1 ? (
              <Button
                type="button"
                onClick={() => {
                  if (!draft.instanceUrl.trim()) {
                    setError("Instance URL is required.");
                    return;
                  }
                  if (!draft.appName.trim()) {
                    setError("App name is required.");
                    return;
                  }
                  saveDraft(draft);
                  setStep(2);
                  setError(null);
                }}
              >
                Continue
              </Button>
            ) : null}
            {step === 2 ? (
              <>
                <Button
                  type="button"
                  variant="outline"
                  iconRight={ExternalLink}
                  disabled={busy}
                  onClick={() => void openCreateOnGitHub()}
                >
                  Open GitHub
                </Button>
                <Button
                  type="button"
                  disabled={busy}
                  onClick={() => void confirmAppCreated()}
                >
                  I created the app
                </Button>
              </>
            ) : null}
            {step === 3 ? (
              <>
                <Button
                  type="button"
                  variant="outline"
                  iconRight={ExternalLink}
                  onClick={() => void openInstallOnGitHub()}
                >
                  Install on GitHub
                </Button>
                <Button
                  type="button"
                  disabled={busy}
                  onClick={() => void confirmInstalled()}
                >
                  I installed it
                </Button>
              </>
            ) : null}
            {step === 4 ? (
              <Button type="button" onClick={finish}>
                Done
              </Button>
            ) : null}
          </div>
        </footer>
      </div>
    </div>
  );
}

function StepPrepare({
  draft,
  accountHint,
  onChange,
}: {
  draft: WizardDraft;
  accountHint: string;
  onChange: (patch: Partial<WizardDraft>) => void;
}) {
  return (
    <div className="space-y-5">
      <div>
        <h3 className="text-base font-semibold">Choose where the app lives</h3>
        <p className="mt-1 text-sm text-[var(--muted-foreground)]">
          Dockyrd registers a GitHub App for you in one click. You only decide
          who owns it.
        </p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        {(
          [
            {
              id: "personal" as const,
              title: "Personal account",
              hint: "Only your own repositories.",
            },
            {
              id: "organization" as const,
              title: "Organization",
              hint: "Repositories of a GitHub organization.",
            },
          ] as const
        ).map((opt) => (
          <label
            key={opt.id}
            className="flex cursor-pointer gap-3 rounded-xl border p-4 has-checked:border-[#102a43]"
            style={{
              borderColor:
                draft.ownerKind === opt.id ? "#102a43" : "var(--border)",
            }}
          >
            <input
              type="radio"
              name="github-owner"
              className="mt-1"
              checked={draft.ownerKind === opt.id}
              onChange={() => onChange({ ownerKind: opt.id })}
            />
            <span>
              <span className="block text-sm font-medium">{opt.title}</span>
              <span className="text-xs text-[var(--muted-foreground)]">
                {opt.hint}
              </span>
            </span>
          </label>
        ))}
      </div>
      <label className="block text-sm">
        <span className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-foreground)]">
          App name
        </span>
        <input
          className="field-input mt-1 w-full"
          value={draft.appName}
          onChange={(e) => onChange({ appName: e.target.value })}
        />
        <span className="mt-1 block text-xs text-[var(--muted-foreground)]">
          GitHub app names are unique. We added your account name
          {accountHint ? ` (${accountHint})` : ""} to avoid a clash.
        </span>
      </label>
      <label className="block text-sm">
        <span className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-foreground)]">
          Instance URL
        </span>
        <input
          className="field-input mt-1 w-full"
          value={draft.instanceUrl}
          onChange={(e) => onChange({ instanceUrl: e.target.value })}
          placeholder="https://deploy.example.com"
        />
        <span className="mt-1 block text-xs text-[var(--muted-foreground)]">
          GitHub sends push events to this address, so it must be reachable over
          HTTPS.
        </span>
      </label>
      <div
        className="rounded-xl border px-4 py-3 text-sm"
        style={{
          borderColor: "#f0d090",
          background: "#fbf8ef",
          color: "#6b4e00",
        }}
      >
        Running Dockyrd locally? Webhooks need a public URL. Use a tunnel or
        your server domain, then come back to this step.
      </div>
    </div>
  );
}

function StepCreate({
  appName,
  onOpenGitHub,
}: {
  appName: string;
  onOpenGitHub: () => void;
}) {
  return (
    <div className="space-y-5">
      <div>
        <h3 className="text-base font-semibold">Create the app on GitHub</h3>
        <p className="mt-1 text-sm text-[var(--muted-foreground)]">
          Click the button below. GitHub opens in a new tab with everything
          pre-filled — you only confirm.
        </p>
      </div>
      <div className="rounded-xl border border-[var(--border)] p-4">
        <p className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-foreground)]">
          What Dockyrd asks for
        </p>
        <ul className="mt-3 space-y-2">
          {PERMISSIONS.map((p) => (
            <li
              key={p.name}
              className="flex items-center justify-between text-sm"
            >
              <span>{p.name}</span>
              <span className="font-mono text-xs text-[var(--muted-foreground)]">
                {p.level}
              </span>
            </li>
          ))}
        </ul>
      </div>
      <Button type="button" iconRight={ExternalLink} onClick={onOpenGitHub}>
        Open GitHub
      </Button>
      <div
        className="rounded-xl px-4 py-3 text-sm text-[var(--muted-foreground)]"
        style={{ background: "var(--muted)" }}
      >
        On GitHub you will see a green button like{" "}
        <span className="font-medium text-[var(--foreground)]">
          Create GitHub App for {appName || "dockyrd"}
        </span>
        .
      </div>
    </div>
  );
}

function StepInstall({
  slug,
  onInstall,
}: {
  slug?: string;
  onInstall: () => void;
}) {
  return (
    <div className="space-y-5">
      <div>
        <h3 className="text-base font-semibold">
          Install it on your repositories
        </h3>
        <p className="mt-1 text-sm text-[var(--muted-foreground)]">
          GitHub may bring you back here. Now give the app access to the code it
          should deploy.
        </p>
      </div>
      {slug ? (
        <p
          className="rounded-xl border px-4 py-3 text-sm text-green-800"
          style={{ borderColor: "#b8e0c8", background: "#e6f4ec" }}
          role="status"
        >
          App <span className="font-mono">{slug}</span> was created.
        </p>
      ) : null}
      <ol className="list-decimal space-y-2 pl-5 text-sm text-[var(--muted-foreground)]">
        <li>Click Install on GitHub below (opens a new tab).</li>
        <li>Pick your account or organization.</li>
        <li>
          Choose <strong>Only select repositories</strong> and tick the ones you
          want, or choose all repositories.
        </li>
        <li>Click Install. You can change this later on GitHub.</li>
      </ol>
      <Button type="button" iconRight={ExternalLink} onClick={onInstall}>
        Install on GitHub
      </Button>
    </div>
  );
}

function StepDone({ status }: { status: GitHubStatus | null }) {
  return (
    <div className="space-y-5 text-center sm:text-left">
      <div
        className="mx-auto flex size-14 items-center justify-center rounded-full sm:mx-0"
        style={{ background: "#e6f4ec", color: "#176b3d" }}
      >
        <Check className="size-7" strokeWidth={2} />
      </div>
      <div>
        <h3 className="text-base font-semibold">GitHub is connected</h3>
        <p className="mt-1 text-sm text-[var(--muted-foreground)]">
          Dockyrd can now see your repositories and will deploy when you push.
        </p>
      </div>
      <dl className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] text-sm">
        <Row label="App" value={status?.slug?.toUpperCase() ?? "—"} />
        <Row
          label="Installed on"
          value={status?.installed_on?.toUpperCase() ?? "—"}
        />
        <Row
          label="Repositories"
          value={
            status?.repository_count != null
              ? `${status.repository_count} selected`
              : "—"
          }
        />
        <Row
          label="Webhook"
          value={status?.webhook_active ? "Active" : "Pending"}
        />
      </dl>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 px-4 py-3">
      <dt className="text-[var(--muted-foreground)]">{label}</dt>
      <dd className="font-mono text-xs uppercase tracking-wide">{value}</dd>
    </div>
  );
}
