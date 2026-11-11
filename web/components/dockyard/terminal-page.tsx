"use client";

import { FitAddon } from "@xterm/addon-fit";
import { Terminal as XTerm } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import { Settings } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { TerminalSettingsDialog } from "@/components/dockyard/terminal-settings-dialog";
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
  type MeResponse,
  type Project,
  patchTerminalPreferences,
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

type TabSession = {
  term: XTerm;
  fit: FitAddon;
  ws: WebSocket | null;
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

  const tabHostRefs = useRef(new Map<string, HTMLDivElement>());
  const tabSessions = useRef(new Map<string, TabSession>());
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
          getTerminalPreferences().catch(() => ({
            theme_id: loadTerminalThemeId(),
          })),
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
        const preferTest =
          !metaRes.session_kinds.host.available &&
          metaRes.session_kinds.test.available &&
          projectRes.projects.length > 0;
        const id = newTabId();
        if (preferTest) {
          const p = projectRes.projects[0];
          setTabs([
            {
              id,
              kind: "test",
              label: `Test · ${p.name}`,
              badge: "SANDBOX",
              projectId: p.id,
              mode: "shell",
              useRoot: false,
              connState: "idle",
            },
          ]);
          setActiveId(id);
        } else if (targetRes.targets[0]) {
          const t = targetRes.targets[0];
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
        } else if (metaRes.session_kinds.host.available) {
          setTabs([
            {
              id,
              kind: "host",
              label: "Server",
              badge: "ROOT",
              mode: "shell",
              useRoot: false,
              connState: "idle",
            },
          ]);
          setActiveId(id);
        } else if (
          metaRes.session_kinds.test.available &&
          projectRes.projects[0]
        ) {
          const p = projectRes.projects[0];
          setTabs([
            {
              id,
              kind: "test",
              label: `Test · ${p.name}`,
              badge: "SANDBOX",
              projectId: p.id,
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

  const teardownTab = useCallback((tabId: string) => {
    const sess = tabSessions.current.get(tabId);
    if (!sess) return;
    sess.ws?.close();
    sess.term.dispose();
    tabSessions.current.delete(tabId);
  }, []);

  const teardownAllTabs = useCallback(() => {
    for (const id of [...tabSessions.current.keys()]) {
      teardownTab(id);
    }
  }, [teardownTab]);

  useEffect(() => () => teardownAllTabs(), [teardownAllTabs]);

  const applyTheme = useCallback((term: XTerm, id: TerminalThemeId) => {
    const th = themeById(id);
    term.options.theme = { ...th.xterm };
  }, []);

  const copyPlainSelection = useCallback(
    async (term?: XTerm) => {
      const t =
        term ??
        (activeId ? tabSessions.current.get(activeId)?.term : undefined);
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
    },
    [activeId],
  );

  const ensureTabSession = useCallback(
    (tabId: string): TabSession | null => {
      const existing = tabSessions.current.get(tabId);
      if (existing) return existing;
      const host = tabHostRefs.current.get(tabId);
      if (!host) return null;
      const term = new XTerm({
        cursorBlink: true,
        fontFamily: "var(--font-mono, ui-monospace, monospace)",
        fontSize: 14,
        convertEol: true,
        theme: { ...shellTheme.xterm },
      });
      const fit = new FitAddon();
      term.loadAddon(fit);
      term.open(host);
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
      const sess: TabSession = { term, fit, ws: null };
      tabSessions.current.set(tabId, sess);
      return sess;
    },
    [shellTheme.xterm, copyPlainSelection],
  );

  useEffect(() => {
    for (const sess of tabSessions.current.values()) {
      applyTheme(sess.term, themeId);
    }
  }, [themeId, applyTheme]);

  const sendResizeFor = useCallback((sess: TabSession) => {
    if (!sess.ws || sess.ws.readyState !== WebSocket.OPEN) return;
    sess.ws.send(
      JSON.stringify({
        type: "resize",
        cols: sess.term.cols,
        rows: sess.term.rows,
      }),
    );
  }, []);

  useEffect(() => {
    const onResize = () => {
      const sess = activeId ? tabSessions.current.get(activeId) : undefined;
      if (!sess) return;
      sess.fit.fit();
      sendResizeFor(sess);
    };
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, [activeId, sendResizeFor]);

  useEffect(() => {
    const host = activeId ? tabHostRefs.current.get(activeId) : undefined;
    const pane = host?.closest(".terminal-xterm-pane");
    if (!pane || !(pane instanceof HTMLElement)) return;
    const ro = new ResizeObserver(() => {
      const sess = activeId ? tabSessions.current.get(activeId) : undefined;
      if (!sess) return;
      sess.fit.fit();
      sendResizeFor(sess);
    });
    ro.observe(pane);
    return () => ro.disconnect();
  }, [activeId, sendResizeFor]);

  const connectTab = useCallback(
    async (tab: TerminalTab) => {
      const existing = tabSessions.current.get(tab.id);
      if (existing?.ws?.readyState === WebSocket.OPEN) {
        return;
      }
      if (tab.connState === "connecting") {
        return;
      }
      const sess = ensureTabSession(tab.id);
      if (!sess) return;
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
        });
        const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
        const ws = new WebSocket(
          `${proto}//${window.location.host}/api/v1/terminal/ws?ticket=${encodeURIComponent(session.ticket)}`,
        );
        ws.binaryType = "arraybuffer";
        sess.ws = ws;
        const { term, fit } = sess;

        ws.onopen = () => {
          patchTab(tab.id, { connState: "connected" });
          if (activeId === tab.id) {
            fit.fit();
            sendResizeFor(sess);
          }
        };
        ws.onmessage = (ev) => {
          if (typeof ev.data === "string") {
            term.write(ev.data);
            return;
          }
          term.write(new Uint8Array(ev.data as ArrayBuffer));
        };
        ws.onerror = () => {
          patchTab(tab.id, { connState: "error" });
          if (activeId === tab.id) {
            setStatusNote("Connection failed.");
          }
        };
        ws.onclose = () => {
          sess.ws = null;
          patchTab(tab.id, { connState: "idle" });
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
    [activeId, ensureTabSession, patchTab, sendResizeFor],
  );

  useEffect(() => {
    if (!activeTab || didAutoConnect.current) return;
    if (activeTab.connState !== "idle") return;
    const id = activeTab.id;
    requestAnimationFrame(() => {
      if (!tabHostRefs.current.has(id)) return;
      didAutoConnect.current = true;
      void connectTab(activeTab);
    });
  }, [activeTab, connectTab]);

  const selectTab = (id: string) => {
    if (id === activeId) return;
    setActiveId(id);
    requestAnimationFrame(() => {
      const sess = tabSessions.current.get(id);
      if (sess) {
        sess.fit.fit();
        sendResizeFor(sess);
      }
      const tab = tabs.find((t) => t.id === id);
      if (tab && (tab.connState === "idle" || tab.connState === "error")) {
        void connectTab(tab);
      }
    });
  };

  const closeTab = (id: string) => {
    teardownTab(id);
    tabHostRefs.current.delete(id);
    setTabs((prev) => {
      const next = prev.filter((t) => t.id !== id);
      if (activeId === id) {
        const fallback = next[0]?.id ?? "";
        setActiveId(fallback);
        requestAnimationFrame(() => {
          const sess = tabSessions.current.get(fallback);
          if (sess) {
            sess.fit.fit();
            sendResizeFor(sess);
          }
        });
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
    setActiveId(id);
    void connectTab(tab);
  };

  const requestHostTab = () => {
    setMenuOpen(false);
    if (!meta?.session_kinds.host.available) {
      setStatusNote(
        meta?.session_kinds.host.allowed_by_instance
          ? "Only server owners and admins can open a host shell."
          : "Host terminal is disabled. Enable it under Server settings.",
      );
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
    setActiveId(id);
    void connectTab(tab);
  };

  const openHostTab = () => {
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
    setActiveId(id);
    void connectTab(tab);
  };

  const pickTheme = (id: TerminalThemeId) => {
    setThemeId(id);
    saveTerminalThemeId(id);
    void patchTerminalPreferences({ theme_id: id }).catch(() => undefined);
  };

  const clear = () => {
    tabSessions.current.get(activeId)?.term.clear();
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
      <div
        className="flex min-h-0 flex-col [&>header]:mb-4 [&>header]:shrink-0"
        style={{
          height:
            "calc(100dvh - var(--dockyrd-announcement-height, 44px) - 5.5rem - 3rem)",
          maxHeight:
            "calc(100dvh - var(--dockyrd-announcement-height, 44px) - 5.5rem - 3rem)",
        }}
      >
        <PageHeader
          eyebrow="Server"
          title="Terminal"
          subtitle="Shells on your server and in running containers."
          action={<Chip>SESSION LOGGED</Chip>}
        />

        <section
          aria-label="Terminal"
          className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-[20px] border"
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
                      background: selected
                        ? shellTheme.surface
                        : shellTheme.bar,
                    }}
                  >
                    <button
                      type="button"
                      role="tab"
                      aria-selected={selected}
                      onClick={() => selectTab(tab.id)}
                      className="inline-flex min-h-11 min-w-0 items-center gap-2 px-3 text-sm font-medium"
                      style={{ color: shellTheme.chrome.fg }}
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
                        className="inline-flex h-11 w-8 items-center justify-center"
                        style={{ color: shellTheme.chrome.muted }}
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
                  className="flex h-11 w-11 items-center justify-center rounded-[10px] text-lg"
                  style={{ color: shellTheme.chrome.muted }}
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
                      style={
                        {
                          background: shellTheme.bar,
                          borderColor: shellTheme.border,
                          color: shellTheme.chrome.fg,
                          ["--terminal-menu-hover" as string]:
                            shellTheme.surface,
                        } as React.CSSProperties
                      }
                    >
                      <span
                        className="px-2.5 pb-1 pt-2 font-mono text-[11px] uppercase tracking-wide"
                        style={{ color: shellTheme.chrome.muted }}
                      >
                        Open a shell in
                      </span>
                      <span
                        className="px-2.5 pt-2 font-mono text-[11px] uppercase tracking-wide"
                        style={{ color: shellTheme.chrome.muted }}
                      >
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
                              className="rounded-lg px-2.5 py-2 text-left text-sm hover:bg-[var(--terminal-menu-hover)]"
                            >
                              Test shell · {p.name}
                              <span
                                className="block text-xs"
                                style={{ color: shellTheme.chrome.muted }}
                              >
                                No host access · project network
                              </span>
                            </button>
                          ))
                        ) : (
                          <p
                            className="px-2.5 py-2 text-sm"
                            style={{ color: shellTheme.chrome.muted }}
                          >
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
                      <span
                        className="px-2.5 pt-2 font-mono text-[11px] uppercase tracking-wide"
                        style={{ color: shellTheme.chrome.muted }}
                      >
                        Containers
                      </span>
                      {targets.map((t) => (
                        <button
                          key={targetKey(t)}
                          type="button"
                          role="menuitem"
                          onClick={() => addContainerTab(t)}
                          className="rounded-lg px-2.5 py-2 text-left text-sm hover:bg-[var(--terminal-menu-hover)]"
                        >
                          {targetLabel(t)}
                        </button>
                      ))}
                      <span
                        className="px-2.5 pt-2 font-mono text-[11px] uppercase tracking-wide"
                        style={{ color: shellTheme.chrome.muted }}
                      >
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
                        <span
                          className="block text-xs"
                          style={{ color: shellTheme.chrome.muted }}
                        >
                          {hostAvailable
                            ? "Root on the host · owners and admins"
                            : "Disabled · enable in Server settings"}
                        </span>
                      </button>
                      {!hostEnabled ? (
                        <p
                          className="px-2.5 pb-2 text-xs"
                          style={{ color: shellTheme.chrome.muted }}
                        >
                          <Link
                            href="/settings/server/general"
                            className="underline"
                            style={{ color: shellTheme.chrome.fg }}
                          >
                            Instance settings
                          </Link>{" "}
                          → allow host terminal.
                        </p>
                      ) : null}
                    </div>
                  </>
                ) : null}
              </div>
            </div>
            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={() => void copyPlainSelection()}
                className="inline-flex min-h-11 items-center justify-center rounded-[10px] border px-3.5 font-mono text-[11px] font-medium uppercase tracking-wide hover:brightness-110"
                style={{
                  color: shellTheme.chrome.fg,
                  borderColor: shellTheme.chrome.border,
                  background: "transparent",
                }}
              >
                Copy plain
              </button>
              <button
                type="button"
                onClick={clear}
                className="inline-flex min-h-11 items-center justify-center rounded-[10px] border px-3.5 font-mono text-[11px] font-medium uppercase tracking-wide hover:brightness-110"
                style={{
                  color: shellTheme.chrome.fg,
                  borderColor: shellTheme.chrome.border,
                  background: "transparent",
                }}
              >
                Clear
              </button>
              <button
                type="button"
                aria-label="Terminal settings"
                onClick={() => setSettingsOpen(true)}
                className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-[10px] hover:brightness-110"
                style={{ color: shellTheme.chrome.muted }}
              >
                <Settings aria-hidden size={22} strokeWidth={1.75} />
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
                color: shellTheme.chrome.muted,
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
            className="terminal-xterm-pane relative min-h-0 flex-1 p-3"
            style={{ background: shellTheme.xterm.background }}
          >
            {tabs.map((tab) => (
              <div
                key={tab.id}
                ref={(el) => {
                  if (el) tabHostRefs.current.set(tab.id, el);
                  else tabHostRefs.current.delete(tab.id);
                }}
                className="absolute inset-3 overflow-hidden"
                style={{
                  visibility: tab.id === activeId ? "visible" : "hidden",
                  pointerEvents: tab.id === activeId ? "auto" : "none",
                }}
                aria-hidden={tab.id !== activeId}
              />
            ))}
          </div>

          <div
            className="flex flex-wrap items-center justify-between gap-2 border-t px-5 py-2.5 font-mono text-xs"
            style={{
              background: shellTheme.bar,
              borderColor: shellTheme.border,
              color: shellTheme.chrome.muted,
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
      </div>

      <TerminalSettingsDialog
        open={settingsOpen}
        themeId={themeId}
        onPick={pickTheme}
        onClose={() => setSettingsOpen(false)}
      />
    </DashboardShell>
  );
}
