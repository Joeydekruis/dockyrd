"use client";

import { ExternalLink } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { SettingsCopyPill } from "@/components/dockyard/ui/settings-detail-block";
import { Switch } from "@/components/dockyard/ui/switch";
import { Tabs } from "@/components/dockyard/ui/tabs";
import {
  getGitHubStatus,
  getInstanceSettings,
  getMe,
  type InstanceSettings,
  patchInstanceSettings,
} from "@/lib/api";
import { openExternalUrl } from "@/lib/open-external-url";
import { PUBLIC_SITE_ORIGIN } from "@/lib/public-links";

const CONNECT_TABS = [
  { id: "claude", label: "Claude" },
  { id: "cursor", label: "Cursor" },
] as const;

type ConnectTabId = (typeof CONNECT_TABS)[number]["id"];

function buildMcpUrl(inst: InstanceSettings | null, origin: string): string {
  const fromApi = inst?.mcp_url?.trim();
  if (fromApi) {
    return fromApi.replace(/\/$/, "");
  }
  const base = origin.trim() || "https://your-panel";
  return `${base.replace(/\/$/, "")}/mcp`;
}

function ConnectInstructions({ client }: { client: ConnectTabId }) {
  if (client === "claude") {
    return (
      <ol className="list-decimal space-y-2 pl-5 text-sm text-[var(--muted-foreground)]">
        <li>
          In{" "}
          <strong className="font-medium text-[var(--foreground)]">
            Claude
          </strong>
          , open{" "}
          <strong className="font-medium text-[var(--foreground)]">
            Settings
          </strong>{" "}
          →{" "}
          <strong className="font-medium text-[var(--foreground)]">
            Connectors
          </strong>{" "}
          → add a{" "}
          <strong className="font-medium text-[var(--foreground)]">
            custom connector
          </strong>{" "}
          (remote MCP).
        </li>
        <li>
          Paste the{" "}
          <strong className="font-medium text-[var(--foreground)]">
            MCP server URL
          </strong>{" "}
          above. Use the same HTTPS host as this panel (not{" "}
          <code className="text-xs">localhost</code> unless you develop on the
          server).
        </li>
        <li>
          On first use, Claude opens this Dockyrd instance to sign in (OAuth).
          Approve the scopes your role allows.
        </li>
        <li>
          If connection fails, remove the connector and add it again after
          confirming MCP is enabled below and your instance is on the latest
          version.
        </li>
      </ol>
    );
  }
  return (
    <ol className="list-decimal space-y-2 pl-5 text-sm text-[var(--muted-foreground)]">
      <li>
        In{" "}
        <strong className="font-medium text-[var(--foreground)]">Cursor</strong>
        , open{" "}
        <strong className="font-medium text-[var(--foreground)]">
          Settings
        </strong>{" "}
        → <strong className="font-medium text-[var(--foreground)]">MCP</strong>{" "}
        → add an{" "}
        <strong className="font-medium text-[var(--foreground)]">HTTP</strong>{" "}
        server.
      </li>
      <li>
        Set the server URL to the copied MCP endpoint (Streamable HTTP). Cursor
        discovers OAuth from the panel when you connect.
      </li>
      <li>
        Complete sign-in in the browser when prompted, including 2FA if your
        instance requires it.
      </li>
      <li>
        Prefer OAuth over pasting API tokens into project files. For{" "}
        <code className="text-xs">dockyrd mcp --stdio</code> on the host only,
        use a <code className="text-xs">dyk_*</code> token from Account
        settings.
      </li>
    </ol>
  );
}

export function McpSettingsPage() {
  const [instSettings, setInstSettings] = useState<InstanceSettings | null>(
    null,
  );
  const [instanceUrl, setInstanceUrl] = useState("");
  const [isAdmin, setIsAdmin] = useState(false);
  const [mcpSaving, setMcpSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [connectTab, setConnectTab] = useState<ConnectTabId>("claude");

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

  const mcpUrl = useMemo(
    () => buildMcpUrl(instSettings, instanceUrl),
    [instSettings, instanceUrl],
  );

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
          Connect Claude or Cursor to debug deploys, read logs, and trigger
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
              <McpUrlBlock
                mcpUrl={mcpUrl}
                mcpDisabled={Boolean(instSettings?.mcp_disabled)}
              />

              <div className="mt-8">
                <h2 className="mb-3 text-sm font-medium">How to connect</h2>
                <Tabs
                  tabs={[...CONNECT_TABS]}
                  activeId={connectTab}
                  onChange={(id) => setConnectTab(id as ConnectTabId)}
                />
                <div
                  className="mt-4"
                  role="tabpanel"
                  aria-label={
                    connectTab === "claude" ? "Claude setup" : "Cursor setup"
                  }
                >
                  <ConnectInstructions client={connectTab} />
                </div>
              </div>

              {isAdmin && instSettings ? (
                <div className="mt-8 space-y-4 border-t border-[var(--border)] pt-6">
                  <h2 className="text-sm font-medium">Instance controls</h2>
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
              <div className="mt-6 flex flex-wrap gap-2">
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

function McpUrlBlock({
  mcpUrl,
  mcpDisabled,
}: {
  mcpUrl: string;
  mcpDisabled: boolean;
}) {
  return (
    <>
      <h2 className="text-sm font-medium">MCP server URL</h2>
      <p className="mt-1 max-w-2xl text-sm text-[var(--muted-foreground)]">
        Remote clients use this Streamable HTTP endpoint with OAuth 2.1.
      </p>
      <div className="mt-3 max-w-full">
        <SettingsCopyPill value={mcpUrl} />
      </div>
      {mcpDisabled ? (
        <p className="mt-3 text-sm text-amber-700" role="status">
          MCP is disabled on this instance. Re-enable below before connecting
          clients.
        </p>
      ) : null}
    </>
  );
}
