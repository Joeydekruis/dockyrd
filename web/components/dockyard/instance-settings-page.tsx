"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { SnackbarViewport } from "@/components/dockyard/ui/snackbar";
import {
  getInstanceSettings,
  patchInstanceSettings,
  testSMTP,
} from "@/lib/api";

type SnackbarState = {
  message: string;
  tone: "success" | "warn" | "info";
} | null;

const settingsCardClass =
  "mb-8 space-y-4 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6";

export function InstanceSettingsPage() {
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [instance, setInstance] = useState<{
    public_url: string;
    acme_email: string;
    wildcard_domain: string;
    require_2fa_all: boolean;
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

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const inst = await getInstanceSettings().catch(() => null);
      setInstance(inst);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load settings");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onSavePublicUrl(e: React.FormEvent) {
    e.preventDefault();
    if (!instance) return;
    setSnackbar(null);
    try {
      const updated = await patchInstanceSettings({
        public_url: instance.public_url,
        acme_email: instance.acme_email,
        wildcard_domain: instance.wildcard_domain,
        require_2fa_all: instance.require_2fa_all,
      });
      setInstance(updated);
      showSnackbar("URL settings saved.", "success");
    } catch (e) {
      showSnackbar(e instanceof Error ? e.message : "Could not save", "warn");
    }
  }

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
    } catch (e) {
      showSnackbar(e instanceof Error ? e.message : "SMTP test failed", "warn");
    }
  }

  return (
    <DashboardShell title="General" breadcrumb="Server settings">
      <SettingsPageBody>
        <section aria-label="Public URL and TLS" className={settingsCardClass}>
          <h2 className="text-lg font-medium">Public URL &amp; TLS</h2>
          <p className="text-sm text-[var(--muted-foreground)]">
            HTTPS address of this Dockyrd dashboard (via Caddy). Used in
            password-reset and invite links. Set an ACME email for Let&apos;s
            Encrypt certificates. DNS must point at this server; open ports 80
            and 443 on the firewall.
          </p>
          {instance ? (
            <form
              onSubmit={(e) => void onSavePublicUrl(e)}
              className="space-y-3"
            >
              <label className="block text-sm">
                <span className="text-[var(--muted-foreground)]">
                  Public URL
                </span>
                <input
                  className="field-input mt-1 w-full"
                  value={instance.public_url}
                  onChange={(e) =>
                    setInstance({ ...instance, public_url: e.target.value })
                  }
                  placeholder="https://panel.example.com"
                />
              </label>
              <label className="block text-sm">
                <span className="text-[var(--muted-foreground)]">
                  ACME email
                </span>
                <input
                  className="field-input mt-1 w-full"
                  type="email"
                  value={instance.acme_email}
                  placeholder="admin@example.com"
                  onChange={(e) =>
                    setInstance({ ...instance, acme_email: e.target.value })
                  }
                />
              </label>
              <label className="block text-sm">
                <span className="text-[var(--muted-foreground)]">
                  Wildcard domain
                </span>
                <input
                  className="field-input mt-1 w-full"
                  value={instance.wildcard_domain}
                  onChange={(e) =>
                    setInstance({
                      ...instance,
                      wildcard_domain: e.target.value,
                    })
                  }
                  placeholder="*.apps.example.com"
                />
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={instance.require_2fa_all}
                  onChange={(e) =>
                    setInstance({
                      ...instance,
                      require_2fa_all: e.target.checked,
                    })
                  }
                />
                Require 2FA for all users (owners always require 2FA)
              </label>
              <Button type="submit">Save URL settings</Button>
            </form>
          ) : loading ? (
            <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
          ) : (
            <p className="text-sm text-[var(--muted-foreground)]">
              Could not load instance settings (admin only).
            </p>
          )}
        </section>

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
          ) : null}
        </section>

        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
      </SettingsPageBody>

      <SnackbarViewport
        open={snackbar !== null}
        message={snackbar?.message ?? ""}
        tone={snackbar?.tone ?? "info"}
        onDismiss={dismissSnackbar}
      />
    </DashboardShell>
  );
}
