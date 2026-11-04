"use client";

import { ChevronDown } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { Modal } from "@/components/dockyard/ui/modal";
import { SnackbarViewport } from "@/components/dockyard/ui/snackbar";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  type EmailLogEntry,
  getInstanceSettings,
  listEmailLog,
  patchInstanceSettings,
  testSMTP,
} from "@/lib/api";

type SnackbarState = {
  message: string;
  tone: "success" | "warn" | "info";
} | null;

const settingsCardClass =
  "mb-8 space-y-4 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6";

function emailKindLabel(kind: string): string {
  switch (kind) {
    case "smtp_test":
      return "SMTP test";
    case "password_reset":
      return "Password reset";
    case "team_invite":
      return "Team invite";
    default:
      return kind.replaceAll("_", " ");
  }
}

function formatLogTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString(undefined, {
      dateStyle: "medium",
      timeStyle: "short",
    });
  } catch {
    return iso;
  }
}

function EmailLogDetailModal({
  entry,
  onClose,
}: {
  entry: EmailLogEntry | null;
  onClose: () => void;
}) {
  const detailJson =
    entry?.detail && Object.keys(entry.detail).length > 0
      ? JSON.stringify(entry.detail, null, 2)
      : null;

  return (
    <Modal
      open={entry !== null}
      title="Email log entry"
      description={entry ? formatLogTime(entry.created_at_iso) : undefined}
      onClose={onClose}
      size="lg"
    >
      {entry ? (
        <dl className="mt-4 space-y-3 text-sm">
        <div>
          <dt className="text-[var(--muted-foreground)]">Type</dt>
          <dd className="font-medium">{emailKindLabel(entry.kind)}</dd>
        </div>
        <div>
          <dt className="text-[var(--muted-foreground)]">To</dt>
          <dd className="font-medium">{entry.to}</dd>
        </div>
        <div>
          <dt className="text-[var(--muted-foreground)]">From</dt>
          <dd>{entry.from || "—"}</dd>
        </div>
        <div>
          <dt className="text-[var(--muted-foreground)]">Subject</dt>
          <dd>{entry.subject || "—"}</dd>
        </div>
        <div>
          <dt className="text-[var(--muted-foreground)]">Status</dt>
          <dd>
            <StatusBadge
              label={entry.status === "sent" ? "Sent" : "Failed"}
              variant={entry.status === "sent" ? "ok" : "err"}
            />
          </dd>
        </div>
        <div>
          <dt className="text-[var(--muted-foreground)]">Response</dt>
          <dd className="whitespace-pre-wrap break-words font-mono text-xs">
            {entry.response || "—"}
          </dd>
        </div>
        {detailJson ? (
          <div>
            <dt className="text-[var(--muted-foreground)]">Log</dt>
            <dd>
              <pre className="mt-1 max-h-48 overflow-auto rounded-lg border border-[var(--border)] bg-[var(--background)] p-3 font-mono text-xs">
                {detailJson}
              </pre>
            </dd>
          </div>
        ) : null}
        </dl>
      ) : null}
    </Modal>
  );
}

export function EmailSettingsPage() {
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [instance, setInstance] = useState<{
    smtp: {
      host: string;
      port: number;
      username: string;
      from: string;
      use_tls: boolean;
      configured: boolean;
    };
  } | null>(null);
  const [smtpPassword, setSmtpPassword] = useState("");
  const [testTo, setTestTo] = useState("");
  const [snackbar, setSnackbar] = useState<SnackbarState>(null);
  const [logEntries, setLogEntries] = useState<EmailLogEntry[]>([]);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [detailEntry, setDetailEntry] = useState<EmailLogEntry | null>(null);

  const dismissSnackbar = useCallback(() => setSnackbar(null), []);

  useEffect(() => {
    if (!snackbar) return;
    const id = window.setTimeout(dismissSnackbar, 6000);
    return () => window.clearTimeout(id);
  }, [snackbar, dismissSnackbar]);

  function showSnackbar(
    message: string,
    tone: NonNullable<SnackbarState>["tone"],
  ) {
    setSnackbar({ message, tone });
  }

  const refreshLog = useCallback(async () => {
    try {
      const { entries } = await listEmailLog();
      setLogEntries(entries);
    } catch {
      setLogEntries([]);
    }
  }, []);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const inst = await getInstanceSettings().catch(() => null);
      setInstance(inst);
      setError(null);
      await refreshLog();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load settings");
    } finally {
      setLoading(false);
    }
  }, [refreshLog]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onSaveSMTP(e: React.FormEvent) {
    e.preventDefault();
    if (!instance) return;
    setSnackbar(null);
    try {
      const updated = await patchInstanceSettings({
        smtp: {
          host: instance.smtp.host,
          port: instance.smtp.port || 587,
          username: instance.smtp.username,
          from: instance.smtp.from,
          use_tls: instance.smtp.use_tls,
          password: smtpPassword || undefined,
        },
      });
      setInstance(updated);
      setSmtpPassword("");
      showSnackbar("SMTP settings saved.", "success");
    } catch (e) {
      showSnackbar(e instanceof Error ? e.message : "Could not save", "warn");
    }
  }

  async function onTestSMTP() {
    setSnackbar(null);
    try {
      await testSMTP(testTo.trim());
      showSnackbar("Test email sent.", "success");
      await refreshLog();
    } catch (e) {
      showSnackbar(e instanceof Error ? e.message : "SMTP test failed", "warn");
      await refreshLog();
    }
  }

  function toggleExpanded(id: string) {
    setExpandedId((prev) => (prev === id ? null : id));
  }

  return (
    <DashboardShell title="Email" breadcrumb="Server settings">
      <SettingsPageBody>
        <section aria-label="SMTP" className={settingsCardClass}>
          <h2 className="text-lg font-medium">SMTP</h2>
          <p className="text-sm text-[var(--muted-foreground)]">
            Outbound email for invites, password resets, and notifications.
            Passwords are stored encrypted on the server.
          </p>
          {instance ? (
            <>
              <form onSubmit={(e) => void onSaveSMTP(e)} className="space-y-3">
                <label className="block text-sm">
                  <span className="text-[var(--muted-foreground)]">
                    SMTP host
                  </span>
                  <input
                    className="field-input mt-1 w-full"
                    value={instance.smtp.host}
                    placeholder="smtp.example.com"
                    onChange={(e) =>
                      setInstance({
                        ...instance,
                        smtp: { ...instance.smtp, host: e.target.value },
                      })
                    }
                  />
                </label>
                <div className="grid grid-cols-2 gap-3">
                  <label className="block text-sm">
                    <span className="text-[var(--muted-foreground)]">Port</span>
                    <input
                      className="field-input mt-1 w-full"
                      type="number"
                      value={instance.smtp.port || 587}
                      onChange={(e) =>
                        setInstance({
                          ...instance,
                          smtp: {
                            ...instance.smtp,
                            port: Number(e.target.value) || 587,
                          },
                        })
                      }
                    />
                  </label>
                  <label className="flex items-end gap-2 pb-2 text-sm">
                    <input
                      type="checkbox"
                      checked={instance.smtp.use_tls}
                      onChange={(e) =>
                        setInstance({
                          ...instance,
                          smtp: {
                            ...instance.smtp,
                            use_tls: e.target.checked,
                          },
                        })
                      }
                    />
                    TLS (port 465)
                  </label>
                </div>
                <label className="block text-sm">
                  <span className="text-[var(--muted-foreground)]">
                    Username
                  </span>
                  <input
                    className="field-input mt-1 w-full"
                    value={instance.smtp.username}
                    placeholder="smtp-user"
                    onChange={(e) =>
                      setInstance({
                        ...instance,
                        smtp: { ...instance.smtp, username: e.target.value },
                      })
                    }
                  />
                </label>
                <label className="block text-sm">
                  <span className="text-[var(--muted-foreground)]">
                    Password{" "}
                    {instance.smtp.configured ? "(leave blank to keep)" : ""}
                  </span>
                  <input
                    className="field-input mt-1 w-full"
                    type="password"
                    autoComplete="new-password"
                    placeholder={
                      instance.smtp.configured ? "••••••••" : "SMTP password"
                    }
                    value={smtpPassword}
                    onChange={(e) => setSmtpPassword(e.target.value)}
                  />
                </label>
                <label className="block text-sm">
                  <span className="text-[var(--muted-foreground)]">From</span>
                  <input
                    className="field-input mt-1 w-full"
                    value={instance.smtp.from}
                    onChange={(e) =>
                      setInstance({
                        ...instance,
                        smtp: { ...instance.smtp, from: e.target.value },
                      })
                    }
                    placeholder="Dockyard <noreply@example.com>"
                  />
                </label>
                <Button type="submit">Save SMTP settings</Button>
              </form>
              {instance.smtp.configured ? (
                <div className="flex flex-wrap items-end gap-2 border-t border-[var(--border)] pt-4">
                  <label className="block flex-1 text-sm">
                    <span className="text-[var(--muted-foreground)]">
                      Send test email to
                    </span>
                    <input
                      className="field-input mt-1 w-full"
                      type="email"
                      placeholder="you@company.com"
                      value={testTo}
                      onChange={(e) => setTestTo(e.target.value)}
                    />
                  </label>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => void onTestSMTP()}
                  >
                    Test SMTP
                  </Button>
                </div>
              ) : null}
            </>
          ) : loading ? (
            <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
          ) : (
            <p className="text-sm text-[var(--muted-foreground)]">
              Could not load instance settings (admin only).
            </p>
          )}
        </section>

        <section aria-label="Email log" className={settingsCardClass}>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <h2 className="text-lg font-medium">Email log</h2>
            <Button type="button" variant="outline" onClick={() => void refreshLog()}>
              Refresh
            </Button>
          </div>
          <p className="text-sm text-[var(--muted-foreground)]">
            Outbound messages and delivery attempts, including failures.
          </p>
          {logEntries.length === 0 ? (
            <p className="text-sm text-[var(--muted-foreground)]">
              No email activity recorded yet.
            </p>
          ) : (
            <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--background)]">
              {logEntries.map((row) => {
                const expanded = expandedId === row.id;
                return (
                  <li key={row.id}>
                    <button
                      type="button"
                      className="flex w-full items-center gap-3 px-4 py-3 text-left text-sm hover:bg-[var(--surface)]"
                      aria-expanded={expanded}
                      onClick={() => toggleExpanded(row.id)}
                    >
                      <ChevronDown
                        className={`size-4 shrink-0 text-[var(--muted-foreground)] transition-transform ${expanded ? "rotate-180" : ""}`}
                        aria-hidden
                      />
                      <span className="min-w-0 flex-1 font-medium">
                        {emailKindLabel(row.kind)}
                      </span>
                      <span className="hidden truncate text-[var(--muted-foreground)] sm:inline">
                        {row.to}
                      </span>
                      <StatusBadge
                        label={row.status === "sent" ? "Sent" : "Failed"}
                        variant={row.status === "sent" ? "ok" : "err"}
                      />
                      <span className="shrink-0 text-xs text-[var(--muted-foreground)]">
                        {formatLogTime(row.created_at_iso)}
                      </span>
                    </button>
                    {expanded ? (
                      <div className="space-y-3 border-t border-[var(--border)] bg-[var(--surface)] px-4 py-3 text-sm">
                        <p>
                          <span className="text-[var(--muted-foreground)]">
                            To:{" "}
                          </span>
                          {row.to}
                        </p>
                        {row.subject ? (
                          <p>
                            <span className="text-[var(--muted-foreground)]">
                              Subject:{" "}
                            </span>
                            {row.subject}
                          </p>
                        ) : null}
                        {row.response ? (
                          <p className="line-clamp-3 font-mono text-xs text-[var(--muted-foreground)]">
                            {row.response}
                          </p>
                        ) : null}
                        <Button
                          type="button"
                          variant="outline"
                          onClick={() => setDetailEntry(row)}
                        >
                          View details
                        </Button>
                      </div>
                    ) : null}
                  </li>
                );
              })}
            </ul>
          )}
        </section>

        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
      </SettingsPageBody>

      <EmailLogDetailModal
        entry={detailEntry}
        onClose={() => setDetailEntry(null)}
      />

      <SnackbarViewport
        open={snackbar !== null}
        message={snackbar?.message ?? ""}
        tone={snackbar?.tone ?? "info"}
        onDismiss={dismissSnackbar}
      />
    </DashboardShell>
  );
}
