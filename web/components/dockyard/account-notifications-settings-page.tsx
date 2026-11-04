"use client";

import { useCallback, useEffect, useState } from "react";
import {
  SettingsErrorBanner,
  SettingsSection,
} from "@/components/dockyard/account-settings-shared";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import {
  getNotificationPrefs,
  patchNotificationPrefs,
  type NotificationPrefs,
} from "@/lib/api";

const toggles: {
  key: keyof NotificationPrefs;
  label: string;
  description: string;
}[] = [
  {
    key: "deployments",
    label: "Deployments",
    description: "Build and deploy success or failure on projects you follow.",
  },
  {
    key: "security_alerts",
    label: "Security alerts",
    description: "New sign-ins, session revokes, and permission changes.",
  },
  {
    key: "product_updates",
    label: "Product updates",
    description: "Dockyard release notes and maintenance windows.",
  },
];

export function AccountNotificationsSettingsPage() {
  const [prefs, setPrefs] = useState<NotificationPrefs | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getNotificationPrefs();
      setPrefs(data);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load preferences");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onSave(e: React.FormEvent) {
    e.preventDefault();
    if (!prefs) return;
    setSaving(true);
    setSaved(false);
    try {
      const updated = await patchNotificationPrefs(prefs);
      setPrefs(updated);
      setError(null);
      setSaved(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save");
    } finally {
      setSaving(false);
    }
  }

  function setToggle(key: keyof NotificationPrefs, value: boolean) {
    setPrefs((prev) => (prev ? { ...prev, [key]: value } : prev));
  }

  return (
    <DashboardShell title="Notifications" breadcrumb="Account settings">
      <SettingsPageBody>
        <SettingsErrorBanner message={error} />
        {saved ? (
          <p className="mb-4 text-sm text-green-700" role="status">
            Notification preferences saved.
          </p>
        ) : null}
        <SettingsSection
          title="Email notifications"
          description="Choose what Dockyard sends to your account email when SMTP is configured on the server."
        >
          {loading || !prefs ? (
            <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
          ) : (
            <form onSubmit={(e) => void onSave(e)} className="space-y-4">
              {toggles.map((item) => (
                <label
                  key={item.key}
                  className="flex cursor-pointer items-start gap-3 rounded-lg border border-[var(--border)] p-4"
                >
                  <input
                    type="checkbox"
                    className="mt-1 size-4 shrink-0"
                    checked={prefs[item.key]}
                    onChange={(e) => setToggle(item.key, e.target.checked)}
                  />
                  <span>
                    <span className="block text-sm font-medium">
                      {item.label}
                    </span>
                    <span className="block text-sm text-[var(--muted-text)]">
                      {item.description}
                    </span>
                  </span>
                </label>
              ))}
              <Button type="submit" disabled={saving}>
                {saving ? "Saving…" : "Save preferences"}
              </Button>
            </form>
          )}
        </SettingsSection>
      </SettingsPageBody>
    </DashboardShell>
  );
}
