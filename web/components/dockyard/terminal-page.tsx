"use client";

import { FitAddon } from "@xterm/addon-fit";
import { Terminal as XTerm } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import { useCallback, useEffect, useRef, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { Chip } from "@/components/dockyard/ui/chip";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { Select } from "@/components/dockyard/ui/select";
import {
  createTerminalSession,
  getMe,
  listTerminalTargets,
  type MeResponse,
  type TerminalTarget,
} from "@/lib/api";

type ConnState = "idle" | "connecting" | "connected" | "error";

function targetKey(t: TerminalTarget): string {
  return `${t.service_id}\0${t.compose_service ?? ""}`;
}

function targetLabel(t: TerminalTarget): string {
  const suffix = t.compose_service ? ` · ${t.compose_service}` : "";
  return `${t.service_name}${suffix}`;
}

export function TerminalPage() {
  const [user, setUser] = useState<MeResponse | null>(null);
  const [forbidden, setForbidden] = useState(false);
  const [targets, setTargets] = useState<TerminalTarget[]>([]);
  const [selectedKey, setSelectedKey] = useState("");
  const [mode, setMode] = useState<"shell" | "psql" | "valkey-cli">("shell");
  const [useRoot, setUseRoot] = useState(false);
  const [connState, setConnState] = useState<ConnState>("idle");
  const [statusNote, setStatusNote] = useState<string | null>(null);

  const hostRef = useRef<HTMLDivElement | null>(null);
  const termRef = useRef<XTerm | null>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const selected = targets.find((t) => targetKey(t) === selectedKey) ?? null;

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
        return listTerminalTargets();
      })
      .then((res) => {
        if (cancelled || !res) return;
        setTargets(res.targets);
        if (res.targets[0]) {
          setSelectedKey(targetKey(res.targets[0]));
        }
      })
      .catch(() => {
        if (!cancelled) setForbidden(true);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const teardown = useCallback(() => {
    wsRef.current?.close();
    wsRef.current = null;
    termRef.current?.dispose();
    termRef.current = null;
    fitRef.current = null;
    setConnState("idle");
  }, []);

  useEffect(() => () => teardown(), [teardown]);

  const mountTerm = useCallback(() => {
    if (!hostRef.current || termRef.current) return;
    const term = new XTerm({
      cursorBlink: true,
      fontFamily: "var(--font-mono, ui-monospace, monospace)",
      fontSize: 13,
      theme: {
        background: "#FFFFFF",
        foreground: "#0B1F3A",
        cursor: "#1A7F4B",
      },
      convertEol: true,
    });
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(hostRef.current);
    fit.fit();
    termRef.current = term;
    fitRef.current = fit;
  }, []);

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

  const connect = useCallback(async () => {
    if (!selected) return;
    teardown();
    mountTerm();
    setConnState("connecting");
    setStatusNote(null);
    try {
      const session = await createTerminalSession({
        service_id: selected.service_id,
        compose_service: selected.compose_service,
        mode,
        exec_user: useRoot ? "root" : "",
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
        setConnState("connected");
        fitRef.current?.fit();
        sendResize();
      };
      ws.onmessage = (ev) => {
        if (typeof ev.data === "string") {
          term.write(ev.data);
          return;
        }
        const buf = new Uint8Array(ev.data as ArrayBuffer);
        term.write(buf);
      };
      ws.onerror = () => {
        setConnState("error");
        setStatusNote("Connection failed.");
      };
      ws.onclose = () => {
        setConnState("idle");
      };

      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(new TextEncoder().encode(data));
        }
      });
    } catch (err) {
      setConnState("error");
      setStatusNote(
        err instanceof Error ? err.message : "Could not start session.",
      );
    }
  }, [selected, mode, useRoot, teardown, mountTerm, sendResize]);

  const clear = () => {
    termRef.current?.clear();
  };

  if (forbidden) {
    return (
      <DashboardShell title="Terminal">
        <EmptyState
          title="Terminal unavailable"
          description="Only server owners and admins can open container shells."
        />
      </DashboardShell>
    );
  }

  const headerAction = (
    <div className="flex flex-wrap items-center gap-2">
      <Chip>SESSION LOGGED</Chip>
    </div>
  );

  const selectOptions = targets.map((t) => ({
    value: targetKey(t),
    label: targetLabel(t),
  }));

  return (
    <DashboardShell title="Terminal" breadcrumb="Server">
      <PageHeader
        eyebrow="Server"
        title="Terminal"
        subtitle="Interactive shell inside a running workload container."
        action={headerAction}
      />

      <div
        className="rounded-[20px] border p-4"
        style={{ borderColor: "var(--border)", background: "var(--surface)" }}
      >
        <div className="mb-4 flex flex-wrap items-end gap-3">
          <div className="min-w-[240px] flex-1">
            <Select
              aria-label="Container"
              value={selectedKey}
              options={selectOptions}
              onChange={setSelectedKey}
              placeholder="No running containers"
            />
          </div>
          {selected?.service_kind === "postgres" ? (
            <Button
              type="button"
              variant={mode === "psql" ? "primary" : "outline"}
              onClick={() => setMode("psql")}
            >
              Open psql
            </Button>
          ) : null}
          {selected?.service_kind === "redis" ? (
            <Button
              type="button"
              variant={mode === "valkey-cli" ? "primary" : "outline"}
              onClick={() => setMode("valkey-cli")}
            >
              Open valkey-cli
            </Button>
          ) : null}
          {mode !== "shell" ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => setMode("shell")}
            >
              Shell
            </Button>
          ) : null}
          {user?.role === "owner" ? (
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={useRoot}
                onChange={(e) => setUseRoot(e.target.checked)}
              />
              Run as root in container
            </label>
          ) : null}
          <div className="ml-auto flex flex-wrap gap-2">
            <span
              className="font-mono text-[11px] uppercase tracking-wide"
              style={{ color: "var(--muted-text)" }}
            >
              {connState === "connected" ? (
                <span style={{ color: "var(--status-ok-fg)" }}>Connected</span>
              ) : (
                connState
              )}
            </span>
            <Button
              type="button"
              variant="outline"
              onClick={clear}
              disabled={!termRef.current}
            >
              Clear
            </Button>
            <Button
              type="button"
              variant="primary"
              onClick={() => void connect()}
              disabled={!selected || connState === "connecting"}
            >
              {connState === "connected" ? "Reconnect" : "Connect"}
            </Button>
          </div>
        </div>

        {statusNote ? (
          <p className="mb-2 text-sm" style={{ color: "var(--danger)" }}>
            {statusNote}
          </p>
        ) : null}

        <section
          ref={hostRef}
          aria-label="Terminal output"
          className="min-h-[420px] overflow-hidden rounded-[12px] border p-2"
          style={{ borderColor: "var(--border)", background: "#fff" }}
        />
      </div>
    </DashboardShell>
  );
}
