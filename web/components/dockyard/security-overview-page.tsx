"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SecurityNav } from "@/components/dockyard/security-nav";
import { Button } from "@/components/dockyard/ui/button";
import { CodeBlock } from "@/components/dockyard/ui/code-block";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { StatusRow } from "@/components/dockyard/ui/status-row";
import {
  confirmHostAction,
  fixSecurityCheck,
  getSecurityChecks,
  revertHostAction,
  runSecurityChecks,
  type SecurityCheck,
  type SecurityChecksReport,
} from "@/lib/security-api";

const cardClass =
  "space-y-2 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6";

function CheckActions({
  check,
  onRefresh,
}: {
  check: SecurityCheck;
  onRefresh: () => void;
}) {
  const [preview, setPreview] = useState<string | null>(null);
  const [actionID, setActionID] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onFix() {
    setBusy(true);
    try {
      const res = await fixSecurityCheck(check.id);
      setActionID(res.host_action_id);
      setPreview(JSON.stringify(res, null, 2));
      onRefresh();
    } finally {
      setBusy(false);
    }
  }

  async function onConfirm() {
    if (!actionID) return;
    setBusy(true);
    try {
      await confirmHostAction(actionID);
      setActionID(null);
      onRefresh();
    } finally {
      setBusy(false);
    }
  }

  async function onRevert() {
    if (!actionID) return;
    setBusy(true);
    try {
      await revertHostAction(actionID);
      setActionID(null);
      onRefresh();
    } finally {
      setBusy(false);
    }
  }

  if (check.fix === "link" && check.fix_href) {
    return (
      <Button variant="outline" href={check.fix_href}>
        Open settings
      </Button>
    );
  }
  if (check.fix !== "host_action") return null;

  return (
    <div className="flex max-w-md flex-col items-end gap-2">
      <div className="flex flex-wrap justify-end gap-2">
        <Button variant="outline" disabled={busy} onClick={() => void onFix()}>
          Apply fix
        </Button>
        {actionID ? (
          <>
            <Button disabled={busy} onClick={() => void onConfirm()}>
              Keep change
            </Button>
            <Button
              variant="outline"
              disabled={busy}
              onClick={() => void onRevert()}
            >
              Revert now
            </Button>
          </>
        ) : null}
      </div>
      {actionID ? (
        <p className="text-xs text-[var(--muted-text)]">
          Confirm within 120 seconds or the change reverts automatically.
        </p>
      ) : null}
      {preview ? <CodeBlock label="Host action" code={preview} /> : null}
    </div>
  );
}

export function SecurityOverviewPage() {
  const [report, setReport] = useState<SecurityChecksReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [rerunning, setRerunning] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      setReport(await getSecurityChecks());
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load checks");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onRerun() {
    setRerunning(true);
    try {
      setReport(await runSecurityChecks());
    } catch (e) {
      setError(e instanceof Error ? e.message : "Run failed");
    } finally {
      setRerunning(false);
    }
  }

  return (
    <DashboardShell title="Security">
      <div className="flex flex-col gap-6 p-8">
        <PageHeader
          title="Security"
          subtitle="Baseline checks and host actions. Changes apply on the server where dockyrd runs — confirm SSH and firewall fixes before the lockout timer expires."
        />
        <SecurityNav active="overview" />
        <div className="flex flex-wrap gap-2">
          <Button disabled={rerunning} onClick={() => void onRerun()}>
            Run checks again
          </Button>
          <Button variant="outline" href="/server/security/addons">
            Protection add-ons
          </Button>
        </div>
        {error ? <p className="text-sm text-red-700">{error}</p> : null}
        <section className={cardClass}>
          <h2 className="text-base font-semibold">Baseline checks</h2>
          {loading ? (
            <p className="text-sm text-[var(--muted-text)]">Loading…</p>
          ) : null}
          {report?.checks.map((check) => (
            <StatusRow
              key={check.id}
              title={check.title}
              status={check.status}
              detail={check.detail}
              actions={
                <CheckActions check={check} onRefresh={() => void refresh()} />
              }
            />
          ))}
        </section>
        <section className={cardClass}>
          <h2 className="text-base font-semibold">Protection add-ons</h2>
          <p className="text-sm text-[var(--muted-text)]">
            Install host firewall and unattended updates from the Add-ons tab.
            CrowdSec and proxy WAF/rate limit are catalog-only in this release.
          </p>
        </section>
      </div>
    </DashboardShell>
  );
}
