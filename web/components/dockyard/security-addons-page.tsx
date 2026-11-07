"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SecurityNav } from "@/components/dockyard/security-nav";
import { Button } from "@/components/dockyard/ui/button";
import { CodeBlock } from "@/components/dockyard/ui/code-block";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  installSecurityAddon,
  listSecurityAddons,
  planSecurityAddon,
  type SecurityAddon,
  uninstallSecurityAddon,
} from "@/lib/security-api";

const cardClass =
  "space-y-4 rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6";

function AddonRow({
  addon,
  onRefresh,
}: {
  addon: SecurityAddon;
  onRefresh: () => void;
}) {
  const [planJSON, setPlanJSON] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onPlan() {
    setBusy(true);
    try {
      const res = await planSecurityAddon(addon.id);
      setPlanJSON(JSON.stringify(res.plan, null, 2));
    } finally {
      setBusy(false);
    }
  }

  async function onInstall() {
    setBusy(true);
    try {
      await installSecurityAddon(addon.id);
      onRefresh();
    } finally {
      setBusy(false);
    }
  }

  async function onUninstall() {
    setBusy(true);
    try {
      await uninstallSecurityAddon(addon.id);
      onRefresh();
    } finally {
      setBusy(false);
    }
  }

  const stateVariant =
    addon.state === "installed"
      ? "ok"
      : addon.state === "failed"
        ? "err"
        : addon.state === "installing"
          ? "info"
          : "idle";

  return (
    <article className="flex flex-col gap-3 border-t border-[var(--border)] py-4 first:border-t-0">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="text-sm font-semibold">{addon.title}</h3>
            <StatusBadge
              variant={stateVariant}
              label={addon.state.replace("_", " ")}
            />
          </div>
          <p className="mt-1 text-sm text-[var(--muted-text)]">
            {addon.description}
          </p>
          {addon.last_error ? (
            <p className="mt-1 text-sm text-red-700">{addon.last_error}</p>
          ) : null}
          {addon.coming_soon ? (
            <p className="mt-1 text-sm text-[var(--muted-text)]">
              {addon.coming_soon}
            </p>
          ) : null}
        </div>
        <div className="flex flex-wrap gap-2">
          {addon.installable ? (
            <>
              <Button
                variant="outline"
                disabled={busy}
                onClick={() => void onPlan()}
              >
                Preview
              </Button>
              <Button
                disabled={busy || addon.state === "installing"}
                onClick={() => void onInstall()}
              >
                Install
              </Button>
              {addon.state === "installed" ? (
                <Button
                  variant="outline"
                  disabled={busy}
                  onClick={() => void onUninstall()}
                >
                  Uninstall
                </Button>
              ) : null}
            </>
          ) : (
            <Button variant="outline" disabled>
              Install
            </Button>
          )}
        </div>
      </div>
      {planJSON ? <CodeBlock label="Install preview" code={planJSON} /> : null}
    </article>
  );
}

export function SecurityAddonsPage() {
  const [addons, setAddons] = useState<SecurityAddon[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const res = await listSecurityAddons();
      setAddons(res.addons);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load add-ons");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
    const installing = addons.some((a) => a.state === "installing");
    if (!installing) return;
    const id = window.setInterval(() => void refresh(), 3000);
    return () => window.clearInterval(id);
  }, [refresh, addons]);

  return (
    <DashboardShell title="Security · Add-ons">
      <div className="flex flex-col gap-6 p-8">
        <PageHeader
          title="Security add-ons"
          subtitle="Optional protection installed via host actions on this machine. No separate privileged helper daemon is required in this MVP phase."
        />
        <SecurityNav active="addons" />
        {error ? <p className="text-sm text-red-700">{error}</p> : null}
        <section className={cardClass}>
          {loading ? (
            <p className="text-sm text-[var(--muted-text)]">Loading catalog…</p>
          ) : null}
          {addons.map((addon) => (
            <AddonRow
              key={addon.id}
              addon={addon}
              onRefresh={() => void refresh()}
            />
          ))}
        </section>
      </div>
    </DashboardShell>
  );
}
