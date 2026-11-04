"use client";

import { ExternalLink, GitBranch } from "lucide-react";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { ConnectGitHubWizard } from "@/components/dockyard/connect-github-wizard";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { ConfirmDialog } from "@/components/dockyard/ui/confirm-dialog";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  disconnectGitHub,
  type GitHubStatus,
  getGitHubStatus,
  getInstanceSettings,
  getMe,
} from "@/lib/api";
import { openExternalUrl } from "@/lib/open-external-url";

type WizardStep = 1 | 2 | 3 | 4;

export function IntegrationsSettingsPage() {
  const search = useSearchParams();
  const router = useRouter();
  const githubQuery = search.get("github");

  const [status, setStatus] = useState<GitHubStatus | null>(null);
  const [instanceUrl, setInstanceUrl] = useState("");
  const [accountHint, setAccountHint] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [wizardOpen, setWizardOpen] = useState(false);
  const [wizardStep, setWizardStep] = useState<WizardStep>(1);
  const [disconnectOpen, setDisconnectOpen] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [st, inst, me] = await Promise.all([
        getGitHubStatus(),
        getInstanceSettings().catch(() => null),
        getMe().catch(() => null),
      ]);
      setStatus(st);
      const url =
        st.public_url?.trim() ||
        inst?.public_url?.trim() ||
        (typeof window !== "undefined" ? window.location.origin : "");
      setInstanceUrl(url);
      if (me?.name) {
        setAccountHint(me.name.split(/\s+/)[0] ?? me.email.split("@")[0] ?? "");
      } else if (me?.email) {
        setAccountHint(me.email.split("@")[0] ?? "");
      }
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load git providers");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (githubQuery === "app-created") {
      setWizardStep(3);
      setWizardOpen(true);
      router.replace("/settings/integrations", { scroll: false });
    }
  }, [githubQuery, router]);

  const connected =
    Boolean(status?.installation_id) &&
    Boolean(status?.app_configured ?? status?.linked);

  const setupIncomplete =
    Boolean(status?.app_configured ?? status?.linked) &&
    !status?.installation_id;

  function openWizard(step: WizardStep = 1) {
    setWizardStep(step);
    setWizardOpen(true);
  }

  async function onDisconnect() {
    try {
      await disconnectGitHub();
      setDisconnectOpen(false);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not disconnect GitHub");
    }
  }

  return (
    <DashboardShell title="Git providers" breadcrumb="Settings">
      <SettingsPageBody>
        <p className="mb-6 max-w-2xl text-sm text-[var(--muted-foreground)]">
          Connect source control to deploy from GitHub repositories and receive
          push webhooks on this Dockyard instance.
        </p>

        {error ? (
          <p className="mb-4 text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}

        <section aria-label="Connected providers" className="mb-8 space-y-3">
          <h2 className="text-sm font-medium">Connected</h2>
          {loading ? (
            <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
          ) : connected && status ? (
            <ConnectedGitHubCard
              status={status}
              onManage={() => {
                if (status.manage_url) openExternalUrl(status.manage_url);
              }}
              onDisconnect={() => setDisconnectOpen(true)}
            />
          ) : setupIncomplete && status ? (
            <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
              <p className="text-sm font-medium">
                {status.slug ?? "GitHub App"}
              </p>
              <p className="mt-1 text-sm text-[var(--muted-foreground)]">
                App created — finish by installing it on your repositories.
              </p>
              <Button
                type="button"
                className="mt-4"
                onClick={() => openWizard(3)}
              >
                Finish GitHub setup
              </Button>
            </div>
          ) : (
            <p className="rounded-xl border border-dashed border-[var(--border)] px-6 py-8 text-sm text-[var(--muted-foreground)]">
              No git providers connected yet.
            </p>
          )}
        </section>

        <section aria-label="Add a provider" className="space-y-3">
          <h2 className="text-sm font-medium">Add a provider</h2>
          <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
            <div className="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
              <div className="flex min-w-0 flex-1 gap-4">
                <span
                  className="flex size-12 shrink-0 items-center justify-center rounded-xl"
                  style={{ background: "#eceef1" }}
                >
                  <GitBranch className="size-6" strokeWidth={1.8} aria-hidden />
                </span>
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="text-base font-semibold">GitHub</h3>
                    <StatusBadge label="Recommended" variant="info" />
                  </div>
                  <p className="mt-1 max-w-2xl text-sm text-[var(--muted-foreground)]">
                    Deploy from github.com through a GitHub App. Auto-deploys on
                    push when enabled on a service.
                  </p>
                </div>
              </div>
              <div className="shrink-0 lg:pt-1">
                {!connected ? (
                  <Button type="button" onClick={() => openWizard(1)}>
                    Connect GitHub
                  </Button>
                ) : (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => openWizard(3)}
                  >
                    Manage access
                  </Button>
                )}
              </div>
            </div>
          </div>
        </section>

        <ConnectGitHubWizard
          open={wizardOpen}
          initialStep={wizardStep}
          instanceUrl={instanceUrl}
          accountHint={accountHint}
          onClose={() => setWizardOpen(false)}
          onComplete={() => void refresh()}
        />

        <ConfirmDialog
          open={disconnectOpen}
          title="Disconnect GitHub?"
          description="Dockyard will remove stored GitHub App credentials on this instance. Push webhooks and private repo access stop until you connect again."
          confirmLabel="Disconnect"
          cancelLabel="Cancel"
          destructive
          onCancel={() => setDisconnectOpen(false)}
          onConfirm={() => void onDisconnect()}
        />
      </SettingsPageBody>
    </DashboardShell>
  );
}

function ConnectedGitHubCard({
  status,
  onManage,
  onDisconnect,
}: {
  status: GitHubStatus;
  onManage: () => void;
  onDisconnect: () => void;
}) {
  const repoLine =
    status.repository_count != null
      ? `${status.repository_count} repositories`
      : "repositories";
  const installedOn = status.installed_on ?? "your account";

  return (
    <article className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex gap-4">
          <span
            className="flex size-12 shrink-0 items-center justify-center rounded-xl"
            style={{ background: "#e3f0ff", color: "#0b3d91" }}
          >
            <GitBranch className="size-6" strokeWidth={1.8} aria-hidden />
          </span>
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="font-medium">{status.slug ?? "GitHub App"}</h3>
              <StatusBadge label="Connected" variant="ok" />
            </div>
            <p className="mt-2 max-w-xl text-sm text-[var(--muted-foreground)]">
              GitHub App · installed on {installedOn} · {repoLine}. Manage opens
              the app on GitHub in a new tab, to change repository access or
              uninstall.
            </p>
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            variant="outline"
            iconRight={ExternalLink}
            onClick={onManage}
            disabled={!status.manage_url}
          >
            Manage on GitHub
          </Button>
          <Button type="button" variant="outline" onClick={onDisconnect}>
            Disconnect
          </Button>
        </div>
      </div>
    </article>
  );
}
