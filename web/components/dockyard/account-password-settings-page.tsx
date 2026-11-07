"use client";

import { useCallback, useEffect, useState } from "react";
import {
  Owner2FARequiredBanner,
  SettingsErrorBanner,
  SettingsSection,
} from "@/components/dockyrd/account-settings-shared";
import { DashboardShell } from "@/components/dockyrd/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyrd/dockyrd-page-loading";
import { SettingsPageBody } from "@/components/dockyrd/settings-page-body";
import { Button } from "@/components/dockyrd/ui/button";
import {
  changePassword,
  enableTotp,
  getAuthSecurity,
  getMe,
  type MeResponse,
  setupTotp,
} from "@/lib/api";

export function AccountPasswordSettingsPage() {
  const [user, setUser] = useState<MeResponse | null>(null);
  const [totpEnabled, setTotpEnabled] = useState(false);
  const [setupSecret, setSetupSecret] = useState<string | null>(null);
  const [setupURL, setSetupURL] = useState<string | null>(null);
  const [confirmCode, setConfirmCode] = useState("");
  const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [passwordSaved, setPasswordSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const [changingPassword, setChangingPassword] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [me, sec] = await Promise.all([getMe(), getAuthSecurity()]);
      setUser(me);
      setTotpEnabled(sec.totp_enabled);
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

  async function onChangePassword(e: React.FormEvent) {
    e.preventDefault();
    if (newPassword !== confirmPassword) {
      setError("New passwords do not match");
      return;
    }
    setChangingPassword(true);
    setPasswordSaved(false);
    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      setError(null);
      setPasswordSaved(true);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not change password",
      );
    } finally {
      setChangingPassword(false);
    }
  }

  const needs2FA =
    user?.role === "owner" && user.totp_enabled === false && !totpEnabled;

  return (
    <DashboardShell title="Password & 2FA" breadcrumb="Account settings">
      <SettingsPageBody>
        <Owner2FARequiredBanner show={!loading && !!needs2FA} />
        <SettingsErrorBanner message={error} />

        <SettingsSection title="Password">
          {passwordSaved ? (
            <p className="text-sm text-green-700" role="status">
              Password updated.
            </p>
          ) : null}
          <form
            onSubmit={(e) => void onChangePassword(e)}
            className="space-y-3"
          >
            <label className="flex flex-col gap-1 text-sm">
              <span className="text-[var(--muted-foreground)]">
                Current password
              </span>
              <input
                type="password"
                className="field-input"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                autoComplete="current-password"
                required
              />
            </label>
            <label className="flex flex-col gap-1 text-sm">
              <span className="text-[var(--muted-foreground)]">
                New password
              </span>
              <input
                type="password"
                className="field-input"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                autoComplete="new-password"
                minLength={12}
                required
              />
            </label>
            <label className="flex flex-col gap-1 text-sm">
              <span className="text-[var(--muted-foreground)]">
                Confirm new password
              </span>
              <input
                type="password"
                className="field-input"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                autoComplete="new-password"
                minLength={12}
                required
              />
            </label>
            <Button type="submit" disabled={changingPassword}>
              {changingPassword ? "Updating…" : "Update password"}
            </Button>
          </form>
        </SettingsSection>

        <SettingsSection title="Two-factor authentication">
          {loading ? (
            <DockyrdPageLoading className="py-2" cubeSize={28} />
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
        </SettingsSection>
      </SettingsPageBody>
    </DashboardShell>
  );
}
