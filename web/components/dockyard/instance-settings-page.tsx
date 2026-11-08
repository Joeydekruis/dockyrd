"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { AppsOnYourDomainSection } from "@/components/dockyard/apps-on-your-domain-section";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { SnackbarViewport } from "@/components/dockyard/ui/snackbar";
import { Switch } from "@/components/dockyard/ui/switch";
import {
  getInstanceSettings,
  getMe,
  type InstanceSettings,
  patchInstanceSettings,
  postServerEdgeReload,
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
  const [instance, setInstance] = useState<InstanceSettings | null>(null);
  const [isOwner, setIsOwner] = useState(false);
  const [terminalHostBusy, setTerminalHostBusy] = useState(false);
  const [edgeRepairBusy, setEdgeRepairBusy] = useState(false);
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
      const [inst, me] = await Promise.all([
        getInstanceSettings().catch(() => null),
        getMe().catch(() => null),
      ]);
      setInstance(inst);
      setIsOwner(me?.role === "owner");
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

  async function setTerminalAllowHost(next: boolean) {
    if (!instance || terminalHostBusy) return;
    const prev = Boolean(instance.terminal_allow_host);
    setInstance({ ...instance, terminal_allow_host: next });
    setTerminalHostBusy(true);
    try {
      const updated = await patchInstanceSettings({
        terminal_allow_host: next,
      });
      setInstance(updated);
      showSnackbar("Terminal access settings saved.", "success");
    } catch (e) {
      setInstance({ ...instance, terminal_allow_host: prev });
      showSnackbar(e instanceof Error ? e.message : "Could not save", "warn");
    } finally {
      setTerminalHostBusy(false);
    }
  }

  async function repairEdge() {
    if (edgeRepairBusy) return;
    setEdgeRepairBusy(true);
    try {
      const res = await postServerEdgeReload();
      if (res.healthy_after) {
        showSnackbar(
          res.was_unhealthy
            ? "Edge proxy restarted; HTTPS looks healthy."
            : "Edge proxy restarted and routes reloaded.",
          "success",
        );
      } else {
        showSnackbar(
          "Edge proxy restarted. Certificates may take a minute — confirm DNS and ACME email.",
          "warn",
        );
      }
    } catch (e) {
      showSnackbar(
        e instanceof Error ? e.message : "Edge repair failed",
        "warn",
      );
    } finally {
      setEdgeRepairBusy(false);
    }
  }

  async function onSavePublicUrl(e: React.FormEvent) {
    e.preventDefault();
    if (!instance) return;
    setSnackbar(null);
    try {
      const updated = await patchInstanceSettings({
        public_url: instance.public_url,
        acme_email: instance.acme_email,
        require_2fa_all: instance.require_2fa_all,
      });
      setInstance(updated);
      showSnackbar("URL settings saved.", "success");
    } catch (e) {
      showSnackbar(e instanceof Error ? e.message : "Could not save", "warn");
    }
  }

  return (
    <DashboardShell title="General" breadcrumb="Server settings">
      <SettingsPageBody>
        <section aria-label="Public URL and TLS" className={settingsCardClass}>
          <h2 className="text-lg font-medium">Dashboard URL &amp; TLS</h2>
          <p className="text-sm text-[var(--muted-foreground)]">
            HTTPS address of this Dockyrd dashboard (via Caddy). Used in
            password-reset and invite links. Configure outbound email under{" "}
            <Link href="/settings/server/email" className="underline">
              Email
            </Link>
            . Set an ACME email for Let&apos;s Encrypt certificates. DNS must
            point at this server; open ports 80 and 443 on the firewall.
          </p>
          {instance ? (
            <form
              onSubmit={(e) => void onSavePublicUrl(e)}
              className="space-y-3"
            >
              <label className="block text-sm">
                <span className="text-[var(--muted-foreground)]">
                  Dashboard URL
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
              <div className="flex flex-wrap items-center gap-3">
                <Button type="submit">Save URL settings</Button>
                {isOwner ? (
                  <Button
                    type="button"
                    variant="outline"
                    disabled={edgeRepairBusy}
                    onClick={() => void repairEdge()}
                  >
                    {edgeRepairBusy ? "Repairing edge…" : "Repair edge (Caddy)"}
                  </Button>
                ) : null}
              </div>
              <p className="text-xs text-[var(--muted-foreground)]">
                Dockyrd also restarts the edge proxy automatically when HTTPS or
                certificate issuance looks broken — no SSH required.
              </p>
              <AppsOnYourDomainSection
                instance={instance}
                onInstanceChange={setInstance}
                onMessage={showSnackbar}
              />
            </form>
          ) : loading ? (
            <DockyrdPageLoading className="py-2" cubeSize={28} />
          ) : (
            <p className="text-sm text-[var(--muted-foreground)]">
              Could not load instance settings (admin only).
            </p>
          )}
        </section>

        {instance && isOwner ? (
          <section aria-label="Web terminal" className={settingsCardClass}>
            <h2 className="text-lg font-medium">Host terminal</h2>
            <p className="text-sm text-[var(--muted-foreground)]">
              When enabled, the server owner can open a root shell on the host
              from the Terminal page.
            </p>
            <Switch
              label="Allow host terminal (owner only)"
              checked={Boolean(instance.terminal_allow_host)}
              disabled={terminalHostBusy}
              onChange={(v) => void setTerminalAllowHost(v)}
            />
          </section>
        ) : null}

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
