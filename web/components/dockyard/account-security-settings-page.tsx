"use client";

import { useCallback, useEffect, useState } from "react";
import {
  SettingsErrorBanner,
  SettingsSection,
} from "@/components/dockyard/account-settings-shared";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import {
  type AuthSession,
  listAuthSessions,
  revokeAuthSession,
} from "@/lib/api";

export function AccountActiveSessionsSection() {
  const [sessions, setSessions] = useState<AuthSession[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const sess = await listAuthSessions();
      setSessions(sess.sessions);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load sessions");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onRevoke(id: string) {
    try {
      await revokeAuthSession(id);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Revoke failed");
    }
  }

  return (
    <>
      <SettingsErrorBanner message={error} />
      <SettingsSection title="Active sessions">
        {loading ? (
          <DockyrdPageLoading className="py-2" cubeSize={28} />
        ) : (
          <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
            {sessions.length === 0 ? (
              <li className="px-4 py-4 text-sm text-[var(--muted-foreground)]">
                No sessions.
              </li>
            ) : (
              sessions.map((s) => (
                <li
                  key={s.id}
                  className="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-sm"
                >
                  <div>
                    <p>{s.user_agent || "Unknown device"}</p>
                    <p className="text-xs text-[var(--muted-foreground)]">
                      {s.ip || "—"} ·{" "}
                      {s.current ? "This device" : "Other device"}
                    </p>
                  </div>
                  {!s.current ? (
                    <Button
                      type="button"
                      variant="link"
                      className="text-xs"
                      style={{ color: "var(--danger)" }}
                      onClick={() => void onRevoke(s.id)}
                    >
                      Revoke
                    </Button>
                  ) : null}
                </li>
              ))
            )}
          </ul>
        )}
      </SettingsSection>
    </>
  );
}

export function AccountSecuritySettingsPage() {
  return (
    <DashboardShell title="Profile & security" breadcrumb="Account settings">
      <SettingsPageBody>
        <AccountActiveSessionsSection />
      </SettingsPageBody>
    </DashboardShell>
  );
}
