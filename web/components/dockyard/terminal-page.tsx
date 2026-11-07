"use client";

import { FitAddon } from "@xterm/addon-fit";
import { Terminal as XTerm } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import Link from "next/link";
import { useCallback, useEffect, useId, useRef, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { TerminalSettingsDialog } from "@/components/dockyard/terminal-settings-dialog";
import { Button } from "@/components/dockyard/ui/button";
import { Chip } from "@/components/dockyard/ui/chip";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import {
  createTerminalSession,
  getMe,
  getTerminalMeta,
  getTerminalPreferences,
  listProjects,
  listTerminalTargets,
  patchTerminalPreferences,
  type MeResponse,
  type Project,
  type TerminalMeta,
  type TerminalTarget,
} from "@/lib/api";
import {
  loadTerminalThemeId,
  saveTerminalThemeId,
  stripAnsi,
  type TerminalThemeId,
  themeById,
} from "@/lib/terminal-themes";

type ConnState = "idle" | "connecting" | "connected" | "error";
type SessionKind = "container" | "host" | "test";

type TerminalTab = {
  id: string;
  kind: SessionKind;
  label: string;
  badge?: "ROOT" | "SANDBOX";
  target?: TerminalTarget;
  projectId?: string;
  mode: "shell" | "psql" | "valkey-cli";
  useRoot: boolean;
  connState: ConnState;
};

function targetKey(t: TerminalTarget): string {
  return `${t.service_id}\0${t.compose_service ?? ""}`;
}

function targetLabel(t: TerminalTarget): string {
  const suffix = t.compose_service ? ` · ${t.compose_service}` : "";
  return `${t.service_name}${suffix}`;
}

function newTabId(): string {
  return `t-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`;
}

export function TerminalPage() {
  const [user, setUser] = useState<MeResponse | null>(null);
  const [forbidden, setForbidden] = useState(false);
  const [targets, setTargets] = useState<TerminalTarget[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [meta, setMeta] = useState<TerminalMeta | null>(null);
  const [tabs, setTabs] = useState<TerminalTab[]>([]);
  const [activeId, setActiveId] = useState("");
  const [menuOpen, setMenuOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [themeId, setThemeId] = useState<TerminalThemeId>("dockyrd-dark");
  const [statusNote, setStatusNote] = useState<string | null>(null);
  const [hostMfaOpen, setHostMfaOpen] = useState(false);
  const [hostMfaCode, setHostMfaCode] = useState("");
  const mfaInputId = useId();

  const hostRef = useRef<HTMLDivElement | null>(null);
  const termRef = useRef<XTerm | null>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const activeTabRef = useRef("");
  const didAutoConnect = useRef(false);

  const activeTab = tabs.find((t) => t.id === activeId) ?? null;
  const shellTheme = themeById(themeId);

  useEffect(() => {
    setThemeId(loadTerminalThemeId());
  }, []);

  useEffect(() => {
    let cancelled = false;
    getMe()
      .then((me) => {
        if (cancelled) return;
        if (me.role !== "owner" && me.role !== "admin") {
          setForbidden(true);
          return;
        }
        setUser(me);
        return Promise.all([
          listTerminalTargets(),
          getTerminalMeta(),
          listProjects(),
          getTerminalPreferences().catch(() => ({ theme_id: loadTerminalThemeId() })),
        ]);
      })
      .then((res) => {
        if (cancelled || !res) return;
        const [targetRes, metaRes, projectRes, prefs] = res;
        setTargets(targetRes.targets);
        setMeta(metaRes);
        setProjects(projectRes.projects);
        if (prefs.theme_id) {
          setThemeId(prefs.theme_id as TerminalThemeId);
          saveTerminalThemeId(prefs.theme_id as TerminalThemeId);
        }
        if (targetRes.targets[0]) {
          const t = targetRes.targets[0];
          const id = newTabId();
          setTabs([
            {
              id,
              kind: "container",
              label: targetLabel(t),
              target: t,
              mode: "shell",
              useRoot: false,
              connState: "idle",
            },
          ]);
          setActiveId(id);
        }
      })
      .catch(() => {
        if (!cancelled) setForbidden(true);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const patchTab = useCallback((id: string, patch: Partial<TerminalTab>) => {
    setTabs((prev) => prev.map((t) => (t.id === id ? { ...t, ...patch } : t)));
  }, []);

  const teardown = useCallback(() => {
    wsRef.current?.close();
    wsRef.current = null;
    termRef.current?.dispose();
    termRef.current = null;
    fitRef.current = null;
  }, []);

  useEffect(() => () => teardown(), [teardown]);

  const applyTheme = useCallback((term: XTerm, id: TerminalThemeId) => {
    const th = themeById(id);
    term.options.theme = { ...th.xterm };
  }, []);

  const copyPlainSelection = useCallback(async (term?: XTerm) => {
    const t = term ?? termRef.current;
    if (!t) return;
    const raw = t.getSelection();
    if (!raw) return;
    const plain = stripAnsi(raw);
    try {
      await navigator.clipboard.writeText(plain);
      setStatusNote("Copied plain text (no styling).");
      window.setTimeout(() => setStatusNote(null), 2500);
    } catch {
      setStatusNote("Could not copy to clipboard.");
    }
  }, []);

  const mountTerm = useCallback(() => {
    if (!hostRef.current) return;
    teardown();
    const term = new XTerm({
      cursorBlink: true,
      fontFamily: "var(--font-mono, ui-monospace, monospace)",
      fontSize: 14,
      convertEol: true,
      theme: { ...shellTheme.xterm },
    });
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(hostRef.current);
    fit.fit();
    term.attachCustomKeyEventHandler((ev) => {
      if (
        ev.type === "keydown" &&
        ev.key === "c" &&
        (ev.ctrlKey || ev.metaKey) &&
        ev.shiftKey
      ) {
        void copyPlainSelection(term);
        return false;
      }
      return true;
    });
    termRef.current = term;
    fitRef.current = fit;
  }, [shellTheme.xterm, teardown, copyPlainSelection]);

  useEffect(() => {
    if (termRef.current) {
      applyTheme(termRef.current, themeId);
    }
  }, [themeId, applyTheme]);

  const sendResize = useCallback(() => {
    const term = termRef.current;
    const ws = wsRef.current;
    if (!term || !ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(
      JSON.stringify({
        type: "resize",
        cols: term.cols,
        rows: term.rows,
      }),
    );
  }, []);

  useEffect(() => {
    const onResize = () => {
      fitRef.current?.fit();
      sendResize();
    };
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, [sendResize]);

  const connectTab = useCallback(
    async (tab: TerminalTab, mfaCode?: string) => {
      activeTabRef.current = tab.id;
      mountTerm();
      patchTab(tab.id, { connState: "connecting" });
      setStatusNote(null);
      try {
        const session = await createTerminalSession({
          session_kind: tab.kind,
          project_id: tab.projectId,
          service_id: tab.target?.service_id,
          compose_service: tab.target?.compose_service,
          mode: tab.mode,
          exec_user: tab.useRoot ? "root" : "",
          mfa_code: mfaCode,
        });
        const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
        const ws = new WebSocket(
          `${proto}//${window.location.host}/api/v1/terminal/ws?ticket=${encodeURIComponent(session.ticket)}`,
        );
        ws.binaryType = "arraybuffer";
        wsRef.current = ws;
        const term = termRef.current;
        if (!term) return;

        ws.onopen = () => {
          if (activeTabRef.current !== tab.id) return;
          patchTab(tab.id, { connState: "connected" });
          fitRef.current?.fit();
          sendResize();
        };
        ws.onmessage = (ev) => {
          if (activeTabRef.current !== tab.id) return;
          if (typeof ev.data === "string") {
            term.write(ev.data);
            return;
          }
          term.write(new Uint8Array(ev.data as ArrayBuffer));
        };
        ws.onerror = () => {
          patchTab(tab.id, { connState: "error" });
          setStatusNote("Connection failed.");
        };
        ws.onclose = () => {
          if (activeTabRef.current === tab.id) {
            patchTab(tab.id, { connState: "idle" });
          }
        };

        term.onData((data) => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(new TextEncoder().encode(data));
          }
        });
      } catch (err) {
        patchTab(tab.id, { connState: "error" });
        setStatusNote(
          err instanceof Error ? err.message : "Could not start session.",
        );
      }
    },
    [mountTerm, patchTab, sendResize],
  );

  useEffect(() => {
    if (!activeTab || didAutoConnect.current) return;
    if (activeTab.connState === "idle") {
      didAutoConnect.current = true;
      void connectTab(activeTab);
    }
  }, [activeTab, connectTab]);

  const selectTab = (id: string) => {
    if (id === activeId) return;
    teardown();
    setActiveId(id);
    const tab = tabs.find((t) => t.id === id);
    if (tab) {
      void connectTab(tab);
    }
  };

  const closeTab = (id: string) => {
    setTabs((prev) => {
      const next = prev.filter((t) => t.id !== id);
      if (activeId === id) {
        teardown();
        const fallback = next[0]?.id ?? "";
        setActiveId(fallback);
        const tab = next[0];
        if (tab) void connectTab(tab);
      }
      return next.length ? next : prev;
    });
  };

  const addContainerTab = (target: TerminalTarget) => {
    setMenuOpen(false);
    const id = newTabId();
    const tab: TerminalTab = {
      id,
      kind: "container",
      label: targetLabel(target),
      target,
      mode: "shell",
      useRoot: false,
      connState: "idle",
    };
    setTabs((prev) => [...prev, tab]);
    teardown();
    setActiveId(id);
    void connectTab(tab);
  };

  const requestHostTab = () => {
    setMenuOpen(false);
    if (!meta?.session_kinds.host.available) {
      setStatusNote(
        meta?.session_kinds.host.allowed_by_instance
          ? "Only the server owner can open a host shell."
          : "Host terminal is disabled. Enable it under Server settings (owner only).",
      );
      return;
    }
    if (meta.session_kinds.host.requires_mfa_if_enabled) {
      setHostMfaCode("");
      setHostMfaOpen(true);
      return;
    }
    openHostTab();
  };

  const openTestTab = (project: Project) => {
    setMenuOpen(false);
    const id = newTabId();
    const tab: TerminalTab = {
      id,
      kind: "test",
      label: `Test · ${project.name}`,
      badge: "SANDBOX",
      projectId: project.id,
      mode: "shell",
      useRoot: false,
      connState: "idle",
    };
    setTabs((prev) => [...prev, tab]);
    teardown();
    setActiveId(id);
    void connectTab(tab);
  };

  const openHostTab = (mfaCode?: string) => {
    setHostMfaOpen(false);
    const id = newTabId();
    const tab: TerminalTab = {
      id,
      kind: "host",
      label: "Server",
      badge: "ROOT",
      mode: "shell",
      useRoot: false,
      connState: "idle",
    };
    setTabs((prev) => [...prev, tab]);
    teardown();
    setActiveId(id);
    void connectTab(tab, mfaCode);
  };

  const pickTheme = (id: TerminalThemeId) => {
    setThemeId(id);
    saveTerminalThemeId(id);
    void patchTerminalPreferences({ theme_id: id }).catch(() => undefined);
  };

  const clear = () => {
    termRef.current?.clear();
  };

  const warning = activeTab
    ? activeTab.kind === "host"
      ? {
          tone: "#C45E00",
          lead: "Root on the host.",
          text: "You have full server access. Prefer container or test shells when possible.",
        }
      : activeTab.kind === "test"
        ? {
            tone: "var(--status-ok-fg)",
            lead: "Sandbox.",
            text: "Throwaway container on the project network. It is removed when this session ends.",
          }
        : null
    : null;

  if (forbidden) {
    return (
      <DashboardShell title="Terminal">
        <EmptyState
          title="Terminal unavailable"
          description="Only server owners and admins can open shells."
        />
      </DashboardShell>
    );
  }

  const hostEnabled = meta?.session_kinds.host.allowed_by_instance;
  const hostAvailable = meta?.session_kinds.host.available;

  return (
    <DashboardShell title="Terminal" breadcrumb="Server">
      <PageHeader
        eyebrow="Server"
        title="Terminal"
        subtitle="Shells on your server and in running containers."
        action={<Chip>SESSION LOGGED</Chip>}
      />

      <section
        aria-label="Terminal"
        className="flex flex-col overflow-hidden rounded-[20px] border"
        style={{
          borderColor: shellTheme.border,
          background: shellTheme.surface,
        }}
      >
        <div
          className="flex flex-wrap items-center justify-between gap-2 border-b px-2 py-2"
          style={{
            background: shellTheme.bar,
            borderColor: shellTheme.border,
          }}
        >
          <div
            role="tablist"
            aria-label="Terminal sessions"
            className="flex min-w-0 flex-1 flex-wrap items-center gap-1.5"
          >
            {tabs.map((tab) => {
              const selected = tab.id === activeId;
              return (
                <div
                  key={tab.id}
                  className="inline-flex max-w-[240px] items-center rounded-[10px] border"
                  style={{
                    borderColor: selected ? shellTheme.border : "transparent",
                    background: selected ? shellTheme.surface : shellTheme.bar,
                  }}
                >
                  <button
                    type="button"
                    role="tab"
                    aria-selected={selected}
                    onClick={() => selectTab(tab.id)}
                    className="inline-flex min-h-11 min-w-0 items-center gap-2 px-3 text-sm font-medium"
                    style={{ color: "var(--text)" }}
                  >
                    <span
                      className="h-2 w-2 shrink-0 rounded-full"
                      style={{
                        background:
                          tab.connState === "connected"
                            ? "var(--status-ok-fg)"
                            : "var(--muted-foreground)",
                      }}
                    />
                    <span className="truncate">{tab.label}</span>
                    {tab.badge ? (
                      <span
                        className="shrink-0 rounded-md border px-1.5 font-mono text-[10px] uppercase"
                        style={{
                          borderColor:
                            tab.badge === "ROOT"
                              ? "#C45E00"
                              : "var(--status-ok-fg)",
                          color:
                            tab.badge === "ROOT"
                              ? "#C45E00"
                              : "var(--status-ok-fg)",
                        }}
                      >
                        {tab.badge}
                      </span>
                    ) : null}
                  </button>
                  {tabs.length > 1 ? (
                    <button
                      type="button"
                      aria-label={`Close ${tab.label}`}
                      onClick={() => closeTab(tab.id)}
                      className="inline-flex h-11 w-8 items-center justify-center text-[var(--muted-foreground)]"
                    >
                      ×
                    </button>
                  ) : null}
                </div>
              );
            })}
            <div className="relative">
              <button
                type="button"
                aria-label="New session"
                aria-haspopup="menu"
                aria-expanded={menuOpen}
                onClick={() => setMenuOpen((v) => !v)}
                className="flex h-11 w-11 items-center justify-center rounded-[10px] text-[var(--muted-foreground)]"
              >
                +
              </button>
              {menuOpen ? (
                <>
                  <button
                    type="button"
                    aria-label="Close menu"
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(false)}
                  />
                  <div
                    role="menu"
                    aria-label="New session"
                    className="absolute left-0 top-[52px] z-50 flex w-[min(340px,calc(100vw-3rem))] flex-col gap-0.5 rounded-[14px] border p-1.5 shadow-lg"
                    style={{
                      background: shellTheme.bar,
                      borderColor: shellTheme.border,
                    }}
                  >
                    <span className="px-2.5 pb-1 pt-2 font-mono text-[11px] uppercase tracking-wide text-[var(--muted-foreground)]">
                      Open a shell in
                    </span>
                    <span className="px-2.5 pt-2 font-mono text-[11px] uppercase tracking-wide text-[var(--muted-foreground)]">
                      Test
                    </span>
                    {meta?.session_kinds.test.available ? (
                      projects.length ? (
                        projects.map((p) => (
                          <button
                            key={p.id}
                            type="button"
                            role="menuitem"
                            onClick={() => openTestTab(p)}
                            className="rounded-lg px-2.5 py-2 text-left text-sm hover:bg-[var(--surface-2)]"
                          >
                            Test shell · {p.name}
                            <span className="block text-xs text-[var(--muted-foreground)]">
                              No host access · project network
                            </span>
                          </button>
                        ))
                      ) : (
                        <p className="px-2.5 py-2 text-sm text-[var(--muted-foreground)]">
                          No projects yet.
                        </p>
                      )
                    ) : (
                      <button
                        type="button"
                        role="menuitem"
                        disabled
                        className="rounded-lg px-2.5 py-2 text-left text-sm opacity-60"
                      >
                        Test shell unavailable
                      </button>
                    )}
                    <span className="px-2.5 pt-2 font-mono text-[11px] uppercase tracking-wide text-[var(--muted-foreground)]">
                      Containers
                    </span>
                    {targets.map((t) => (
                      <button
                        key={targetKey(t)}
                        type="button"
                        role="menuitem"
                        onClick={() => addContainerTab(t)}
                        className="rounded-lg px-2.5 py-2 text-left text-sm hover:bg-[var(--surface-2)]"
                      >
                        {targetLabel(t)}
                      </button>
                    ))}
                    <span className="px-2.5 pt-2 font-mono text-[11px] uppercase tracking-wide text-[var(--muted-foreground)]">
                      Advanced
                    </span>
                    <button
                      type="button"
                      role="menuitem"
                      disabled={!hostEnabled && user?.role !== "owner"}
                      onClick={() => requestHostTab()}
                      className="rounded-lg px-2.5 py-2 text-left text-sm disabled:opacity-50"
                    >
                      Server shell
                      <span className="block text-xs text-[var(--muted-foreground)]">
                        {hostAvailable
                          ? "Root on the host · owner only"
                          : hostEnabled
                            ? "Owner only · enable in settings"
                            : "Disabled · enable in Server settings"}
                      </span>
                    </button>
                    {!hostEnabled ? (
                      <p className="px-2.5 pb-2 text-xs text-[var(--muted-foreground)]">
                        <Link
                          href="/settings/server/general"
                          className="underline"
                        >
                          Instance settings
                        </Link>{" "}
                        → allow host terminal (owner, 2FA to enable).
                      </p>
                    ) : null}
                  </div>
                </>
              ) : null}
            </div>
          </div>
          <div className="flex items-center gap-1">
            <Button
              type="button"
              variant="outline"
              onClick={() => void copyPlainSelection()}
            >
              Copy plain
            </Button>
            <Button type="button" variant="outline" onClick={clear}>
              Clear
            </Button>
            <button
              type="button"
              aria-label="Terminal settings"
              onClick={() => setSettingsOpen(true)}
              className="flex h-11 w-11 items-center justify-center rounded-[10px] text-[var(--muted-foreground)]"
            >
              ⚙
            </button>
          </div>
        </div>

        {warning ? (
          <div
            role="note"
            className="flex items-start gap-2 border-b px-5 py-2.5 text-sm"
            style={{
              background: shellTheme.bar,
              borderColor: shellTheme.border,
              color: "var(--muted-foreground)",
            }}
          >
            <strong style={{ color: warning.tone }}>{warning.lead}</strong>
            <span>{warning.text}</span>
          </div>
        ) : null}

        {statusNote ? (
          <p className="px-5 py-2 text-sm text-[var(--danger)]" role="status">
            {statusNote}
          </p>
        ) : null}

        <div
          ref={hostRef}
          className="min-h-[420px] p-3"
          style={{ background: shellTheme.xterm.background }}
        />

        <div
          className="flex flex-wrap items-center justify-between gap-2 border-t px-5 py-2.5 font-mono text-xs text-[var(--muted-foreground)]"
          style={{
            background: shellTheme.bar,
            borderColor: shellTheme.border,
          }}
        >
          <span>
            {activeTab?.connState === "connected"
              ? "Connected"
              : (activeTab?.connState ?? "idle")}
            {activeTab?.kind === "host"
              ? " · root · bash"
              : activeTab?.kind === "test"
                ? " · sandbox · sh"
                : " · container"}
          </span>
          <span>
            {shellTheme.name} · UTF-8 · Ctrl+Shift+C copies plain text
          </span>
        </div>
      </section>

      <TerminalSettingsDialog
        open={settingsOpen}
        themeId={themeId}
        onPick={pickTheme}
        onClose={() => setSettingsOpen(false)}
      />

      {hostMfaOpen ? (
        <div className="fixed inset-0 z-[70] flex items-center justify-center p-4">
          <button
            type="button"
            className="absolute inset-0 bg-[rgba(11,31,58,0.45)]"
            aria-label="Close"
            onClick={() => setHostMfaOpen(false)}
          />
          <form
            className="relative w-full max-w-sm rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6"
            onSubmit={(e) => {
              e.preventDefault();
              openHostTab(hostMfaCode.trim());
            }}
          >
            <h2 className="mb-2 text-lg font-semibold">Confirm 2FA</h2>
            <p className="mb-4 text-sm text-[var(--muted-foreground)]">
              Enter your authenticator code to open a host shell.
            </p>
            <label className="mb-4 block text-sm" htmlFor={mfaInputId}>
              Code
              <input
                id={mfaInputId}
                className="field-input mt-1 w-full"
                inputMode="numeric"
                autoComplete="one-time-code"
                value={hostMfaCode}
                onChange={(e) => setHostMfaCode(e.target.value)}
              />
            </label>
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setHostMfaOpen(false)}
              >
                Cancel
              </Button>
              <Button type="submit" variant="primary">
                Open server shell
              </Button>
            </div>
          </form>
        </div>
      ) : null}
    </DashboardShell>
  );
}
