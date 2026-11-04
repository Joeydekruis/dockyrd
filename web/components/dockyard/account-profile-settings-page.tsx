"use client";

import { useEffect, useState } from "react";
import {
  Owner2FARequiredBanner,
  SettingsErrorBanner,
  SettingsSection,
} from "@/components/dockyard/account-settings-shared";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { getMe, type MeResponse, patchAccount } from "@/lib/api";

export function AccountProfileSettingsPage() {
  const [user, setUser] = useState<MeResponse | null>(null);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    getMe()
      .then((me) => {
        setUser(me);
        setName(me.name);
        setEmail(me.email);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : "Failed to load profile");
      })
      .finally(() => setLoading(false));
  }, []);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setSaved(false);
    try {
      const updated = await patchAccount({
        name: name.trim(),
        email: email.trim(),
      });
      setUser(updated);
      setError(null);
      setSaved(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save profile");
    } finally {
      setSaving(false);
    }
  }

  const needs2FA = user?.role === "owner" && user.totp_enabled === false;

  return (
    <DashboardShell title="Profile" breadcrumb="Account settings">
      <SettingsPageBody>
        <Owner2FARequiredBanner show={!loading && !!needs2FA} />
        <SettingsErrorBanner message={error} />
        {saved ? (
          <p className="mb-4 text-sm text-green-700" role="status">
            Profile saved.
          </p>
        ) : null}
        <SettingsSection title="Profile">
          {loading ? (
            <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
          ) : (
            <form onSubmit={(e) => void onSubmit(e)} className="space-y-4">
              <label className="flex flex-col gap-1 text-sm">
                <span className="text-[var(--muted-foreground)]">Name</span>
                <input
                  className="field-input"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  autoComplete="name"
                  required
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="text-[var(--muted-foreground)]">Email</span>
                <input
                  className="field-input"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoComplete="email"
                  required
                />
              </label>
              <Button type="submit" disabled={saving}>
                {saving ? "Saving…" : "Save profile"}
              </Button>
            </form>
          )}
        </SettingsSection>
      </SettingsPageBody>
    </DashboardShell>
  );
}
