"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import {
  type AuthSession,
  enableTotp,
  getAuthSecurity,
  listAuthSessions,
  revokeAuthSession,
  setupTotp,
} from "@/lib/api";

export function ProfileSettingsPage() {
  const [totpEnabled, setTotpEnabled] = useState(false);
  const [sessions, setSessions] = useState<AuthSession[]>([]);
  const [setupSecret, setSetupSecret] = useState<string | null>(null);
  const [setupURL, setSetupURL] = useState<string | null>(null);
  const [confirmCode, setConfirmCode] = useState("");
  const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [sec, sess] = await Promise.all([
        getAuthSecurity(),
        listAuthSessions(),
      ]);
      setTotpEnabled(sec.totp_enabled);
      setSessions(sess.sessions);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load profile");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onStartTotp() {
    try {
      const res = await setupTotp();
      setSetupSecret(res.secret);
      setSetupURL(res.otpauth_url);
      setRecoveryCodes(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "2FA setup failed");
    }
  }

  async function onEnableTotp(e: React.FormEvent) {
    e.preventDefault();
    if (!setupSecret || !confirmCode.trim()) return;
    try {
      const res = await enableTotp({
        secret: setupSecret,
        code: confirmCode.trim(),
      });
      setRecoveryCodes(res.recovery_codes ?? []);
      setSetupSecret(null);
      setSetupURL(null);
      setConfirmCode("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not enable 2FA");
    }
  }

  async function onRevoke(id: string) {
    try {
      await revokeAuthSession(id);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Revoke failed");
    }
  }

  return (
    <DashboardShell title="Profile & security" breadcrumb="Settings">
      {!loading && !totpEnabled ? (
        <p
          className="mb-4 rounded-lg border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm"
          role="status"
        >
          Owners must enable two-factor authentication before using the rest of
          the dashboard.
        </p>
      ) : null}

      {error ? (
        <p className="mb-4 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-700">
          {error}
        </p>
      ) : null}

      <section className="mb-8 max-w-xl space-y-4 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
        <h2 className="text-lg font-medium">Two-factor authentication</h2>
        {loading ? (
          <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
        ) : totpEnabled ? (
          <p className="text-sm text-green-700">
            Authenticator app is enabled for this account.
          </p>
        ) : setupSecret ? (
          <form onSubmit={onEnableTotp} className="space-y-3">
            {setupURL ? (
              <p className="break-all text-xs text-[var(--muted-foreground)]">
                Add this account in your app (or scan URL):{" "}
                <code className="rounded bg-[var(--muted)]/30 px-1">
                  {setupSecret}
                </code>
              </p>
            ) : null}
            <label className="flex flex-col gap-1 text-sm">
              <span className="text-[var(--muted-foreground)]">
                6-digit code
              </span>
              <input
                value={confirmCode}
                onChange={(e) => setConfirmCode(e.target.value)}
                className="field-input"
                inputMode="numeric"
                autoComplete="one-time-code"
                placeholder="000000"
              />
            </label>
            <Button type="submit">Confirm and enable</Button>
          </form>
        ) : (
          <Button
            type="button"
            variant="outline"
            onClick={() => void onStartTotp()}
          >
            Set up authenticator app
          </Button>
        )}
        {recoveryCodes ? (
          <div className="rounded-lg border border-amber-500/40 bg-amber-500/10 p-3 text-sm">
            <p className="font-medium">
              Save these recovery codes (shown once):
            </p>
            <pre className="mt-2 whitespace-pre-wrap font-mono text-xs">
              {recoveryCodes.join("\n")}
            </pre>
          </div>
        ) : null}
      </section>

      <section className="max-w-xl space-y-3">
        <h2 className="text-lg font-medium">Active sessions</h2>
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
                    {s.ip || "—"} · {s.current ? "This device" : "Other device"}
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
      </section>
    </DashboardShell>
  );
}
