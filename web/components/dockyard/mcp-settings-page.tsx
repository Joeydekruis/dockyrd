"use client";

import { ExternalLink } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { Switch } from "@/components/dockyard/ui/switch";
import {
  getGitHubStatus,
  getInstanceSettings,
  getMe,
  type InstanceSettings,
  patchInstanceSettings,
} from "@/lib/api";
import { openExternalUrl } from "@/lib/open-external-url";
import { PUBLIC_SITE_ORIGIN } from "@/lib/public-links";

export function McpSettingsPage() {
  const [instSettings, setInstSettings] = useState<InstanceSettings | null>(
    null,
  );
  const [instanceUrl, setInstanceUrl] = useState("");
  const [isAdmin, setIsAdmin] = useState(false);
  const [mcpSaving, setMcpSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [inst, st, me] = await Promise.all([
        getInstanceSettings().catch(() => null),
        getGitHubStatus().catch(() => null),
        getMe().catch(() => null),
      ]);
      setInstSettings(inst);
      const url =
        inst?.mcp_url?.replace(/\/mcp$/, "") ||
        st?.public_url?.trim() ||
        inst?.public_url?.trim() ||
        (typeof window !== "undefined" ? window.location.origin : "");
      setInstanceUrl(url);
      setIsAdmin(me?.role === "owner" || me?.role === "admin");
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load MCP settings");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onMcpToggle(
    field: "mcp_disabled" | "mcp_allow_api_token",
    value: boolean,
  ) {
    setMcpSaving(true);
    try {
      const updated = await patchInstanceSettings({ [field]: value });
      setInstSettings(updated);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not save MCP settings");
    } finally {
      setMcpSaving(false);
    }
  }

  return (
    <DashboardShell title="AI assistants (MCP)" breadcrumb="Server settings">
      <SettingsPageBody>
        <p className="mb-6 max-w-2xl text-sm text-[var(--muted-foreground)]">
          Connect Cursor or Claude to debug deploys, read logs, and trigger
          releases on this Dockyrd instance.
        </p>

        {error ? (
          <p className="mb-4 text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}

        <section aria-label="MCP connection" className="space-y-3">
          {loading ? (
            <DockyrdPageLoading className="py-8" cubeSize={28} />
          ) : (
            <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
              <p className="max-w-2xl text-sm text-[var(--muted-foreground)]">
                Remote clients use OAuth 2.1 at{" "}
                <code className="text-xs">
                  {instSettings?.mcp_url ??
                    `${instanceUrl || "https://your-panel"}/mcp`}
                </code>
                .
              </p>
              {instSettings?.mcp_disabled ? (
                <p className="mt-2 text-sm text-amber-700" role="status">
                  MCP is disabled on this instance. Re-enable below to connect
                  clients.
                </p>
              ) : null}
              <ul className="mt-4 list-disc space-y-1 pl-5 text-sm text-[var(--muted-foreground)]">
                <li>
                  Add an HTTP MCP server in Cursor or Claude pointing at your
                  panel URL.
                </li>
                <li>
                  Sign in when prompted (PKCE). Scopes: read, deploy, write.
                </li>
                <li>
                  On the server only:{" "}
                  <code className="text-xs">dockyrd mcp --stdio</code> with a{" "}
                  <code className="text-xs">dyk_</code> API token.
                </li>
              </ul>
              {isAdmin && instSettings ? (
                <div className="mt-6 space-y-4 border-t border-[var(--border)] pt-4">
                  <div className="flex items-center justify-between gap-4">
                    <span className="text-sm">
                      <span className="font-medium">Allow MCP</span>
                      <span className="mt-0.5 block text-[var(--muted-foreground)]">
                        When off, <code className="text-xs">POST /mcp</code>{" "}
                        returns 404.
                      </span>
                    </span>
                    <Switch
                      checked={!instSettings.mcp_disabled}
                      disabled={mcpSaving}
                      label="Allow MCP"
                      onChange={(on) => void onMcpToggle("mcp_disabled", !on)}
                    />
                  </div>
                  <div className="flex items-center justify-between gap-4">
                    <span className="text-sm">
                      <span className="font-medium">
                        Allow API tokens on MCP
                      </span>
                      <span className="mt-0.5 block text-[var(--muted-foreground)]">
                        Optional: accept <code className="text-xs">dyk_*</code>{" "}
                        bearer tokens (not OAuth).
                      </span>
                    </span>
                    <Switch
                      checked={Boolean(instSettings.mcp_allow_api_token)}
                      disabled={mcpSaving || instSettings.mcp_disabled}
                      label="Allow API tokens on MCP"
                      onChange={(on) =>
                        void onMcpToggle("mcp_allow_api_token", on)
                      }
                    />
                  </div>
                </div>
              ) : null}
              <div className="mt-4 flex flex-wrap gap-2">
                <Button
                  type="button"
                  variant="outline"
                  iconRight={ExternalLink}
                  onClick={() =>
                    openExternalUrl(`${PUBLIC_SITE_ORIGIN}/docs/mcp.html`)
                  }
                >
                  MCP setup guide
                </Button>
              </div>
            </div>
          )}
        </section>
      </SettingsPageBody>
    </DashboardShell>
  );
}
