"use client";

import { useCallback, useEffect, useState } from "react";
import {
  SettingsErrorBanner,
  SettingsSection,
} from "@/components/dockyrd/account-settings-shared";
import { DashboardShell } from "@/components/dockyrd/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyrd/dockyrd-page-loading";
import { SettingsPageBody } from "@/components/dockyrd/settings-page-body";
import { Button } from "@/components/dockyrd/ui/button";
import { Field } from "@/components/dockyrd/ui/field";
import {
  type APITokenEntry,
  createAPIToken,
  listAPITokens,
  listProjects,
  type Project,
  revokeAPIToken,
} from "@/lib/api";

const SCOPES = [
  { value: "read", label: "Read — list projects, services, deployments, logs" },
  { value: "deploy", label: "Deploy — read plus trigger deploys" },
  { value: "write", label: "Write — deploy plus service mutations" },
] as const;

export function AccountAPITokensSettingsPage() {
  const [tokens, setTokens] = useState<APITokenEntry[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [name, setName] = useState("");
  const [scope, setScope] = useState<(typeof SCOPES)[number]["value"]>("read");
  const [projectID, setProjectID] = useState("");
  const [createdPlain, setCreatedPlain] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [tok, proj] = await Promise.all([listAPITokens(), listProjects()]);
      setTokens(tok.tokens);
      setProjects(proj.projects);
      if (!projectID && proj.projects[0]?.id) {
        setProjectID(proj.projects[0].id);
      }
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load tokens");
    } finally {
      setLoading(false);
    }
  }, [projectID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setCreatedPlain(null);
    try {
      const res = await createAPIToken({ name, scope, project_id: projectID });
      setCreatedPlain(res.token);
      setName("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setSaving(false);
    }
  }

  async function onRevoke(id: string) {
    try {
      await revokeAPIToken(id);
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Revoke failed");
    }
  }

  return (
    <DashboardShell title="API tokens" breadcrumb="Account settings">
      <SettingsPageBody>
        <SettingsErrorBanner message={error} />
        {createdPlain ? (
          <p className="mb-4 rounded-lg border border-[var(--border)] bg-[var(--surface)] px-4 py-3 font-mono text-sm">
            Copy this token now — it will not be shown again:
            <br />
            <span className="break-all">{createdPlain}</span>
          </p>
        ) : null}
        <SettingsSection
          title="Create token"
          description="Tokens authenticate the dy CLI and automation. They cannot perform admin actions."
        >
          <form className="space-y-4" onSubmit={(e) => void onCreate(e)}>
            <div>
              <Field
                id="token-name"
                label="Name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="CI deploy"
                required
              />
            </div>
            <div>
              <label className="text-sm font-medium" htmlFor="token-project">
                Project
              </label>
              <select
                id="token-project"
                className="mt-1 w-full rounded-lg border border-[var(--border)] bg-[var(--surface)] px-3 py-2 text-sm"
                value={projectID}
                onChange={(e) => setProjectID(e.target.value)}
                required
              >
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <span className="text-sm font-medium">Scope</span>
              <ul className="mt-2 space-y-2">
                {SCOPES.map((s) => (
                  <li key={s.value}>
                    <label className="flex cursor-pointer items-start gap-2 text-sm">
                      <input
                        type="radio"
                        name="scope"
                        value={s.value}
                        checked={scope === s.value}
                        onChange={() => setScope(s.value)}
                      />
                      <span>{s.label}</span>
                    </label>
                  </li>
                ))}
              </ul>
            </div>
            <Button type="submit" disabled={saving || !name || !projectID}>
              {saving ? "Creating…" : "Create token"}
            </Button>
          </form>
        </SettingsSection>
        <SettingsSection title="Active tokens">
          {loading ? (
            <DockyrdPageLoading className="py-2" cubeSize={28} />
          ) : (
            <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
              {tokens.length === 0 ? (
                <li className="px-4 py-4 text-sm text-[var(--muted-foreground)]">
                  No API tokens yet.
                </li>
              ) : (
                tokens.map((t) => (
                  <li
                    key={t.id}
                    className="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-sm"
                  >
                    <div>
                      <p className="font-medium">{t.name}</p>
                      <p className="text-xs text-[var(--muted-foreground)]">
                        {t.scope} · project {t.project_id}
                        {t.last_used_at
                          ? ` · last used ${new Date(t.last_used_at * 1000).toLocaleString()}`
                          : ""}
                      </p>
                    </div>
                    <Button
                      type="button"
                      variant="link"
                      className="text-xs"
                      style={{ color: "var(--danger)" }}
                      onClick={() => void onRevoke(t.id)}
                    >
                      Revoke
                    </Button>
                  </li>
                ))
              )}
            </ul>
          )}
        </SettingsSection>
      </SettingsPageBody>
    </DashboardShell>
  );
}
