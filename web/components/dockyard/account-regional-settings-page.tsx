"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  SettingsErrorBanner,
  SettingsSection,
} from "@/components/dockyard/account-settings-shared";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import {
  browserTimezone,
  formatClockTime,
  listTimeZones,
  type TimeFormat,
  type UserLocalePrefs,
} from "@/lib/user-locale";
import { useUserLocale } from "@/lib/user-locale-context";

const timeFormatOptions: { value: TimeFormat; label: string; hint: string }[] =
  [
    {
      value: "24h",
      label: "24-hour",
      hint: "14:00, 16:30",
    },
    {
      value: "12h",
      label: "12-hour",
      hint: "2PM, 4:30AM",
    },
  ];

export function AccountRegionalSettingsSection() {
  const { prefs: saved, loading, save } = useUserLocale();
  const [draft, setDraft] = useState<UserLocalePrefs | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [savedMsg, setSavedMsg] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!loading && saved) {
      setDraft(saved);
    }
  }, [loading, saved]);

  const timeZones = useMemo(() => listTimeZones(), []);

  const previewTs = Math.floor(Date.now() / 1000) - 3600;
  const preview = draft
    ? formatClockTime(previewTs, draft)
    : formatClockTime(previewTs, saved);

  const useBrowserTz = useCallback(() => {
    setDraft((prev) =>
      prev ? { ...prev, timezone: browserTimezone() } : prev,
    );
  }, []);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!draft) return;
    setSaving(true);
    setSavedMsg(false);
    try {
      await save(draft);
      setError(null);
      setSavedMsg(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save");
    } finally {
      setSaving(false);
    }
  }

  return (
    <>
      <SettingsErrorBanner message={error} />
      {savedMsg ? (
        <p className="mb-4 text-sm text-green-700" role="status">
          Date and time preferences saved. Charts update immediately.
        </p>
      ) : null}
      <SettingsSection
        title="Date & time"
        description="Timezone and clock format for analytics charts and timestamps in the dashboard."
      >
        {loading || !draft ? (
          <DockyrdPageLoading className="py-2" cubeSize={28} />
        ) : (
          <form onSubmit={(e) => void onSubmit(e)} className="space-y-6">
            <label className="flex flex-col gap-1 text-sm">
              <span className="text-[var(--muted-foreground)]">Timezone</span>
              <div className="flex flex-wrap gap-2">
                <select
                  className="field-input min-w-[min(100%,20rem)] flex-1"
                  value={draft.timezone}
                  onChange={(e) =>
                    setDraft({ ...draft, timezone: e.target.value })
                  }
                >
                  {timeZones.map((tz) => (
                    <option key={tz} value={tz}>
                      {tz}
                    </option>
                  ))}
                </select>
                <Button type="button" variant="outline" onClick={useBrowserTz}>
                  Use browser timezone
                </Button>
              </div>
            </label>

            <fieldset className="space-y-3">
              <legend className="text-sm text-[var(--muted-foreground)]">
                Time format
              </legend>
              {timeFormatOptions.map((opt) => (
                <label
                  key={opt.value}
                  className="flex cursor-pointer items-start gap-3 rounded-lg border border-[var(--border)] p-3"
                >
                  <input
                    type="radio"
                    name="time_format"
                    className="mt-1"
                    checked={draft.time_format === opt.value}
                    onChange={() =>
                      setDraft({ ...draft, time_format: opt.value })
                    }
                  />
                  <span>
                    <span className="font-medium">{opt.label}</span>
                    <span className="mt-0.5 block text-[13px] text-[var(--muted-text)]">
                      Example: {opt.hint}
                    </span>
                  </span>
                </label>
              ))}
            </fieldset>

            <p className="text-sm text-[var(--muted-text)]">
              Preview (one hour ago):{" "}
              <span className="font-mono tabular-nums">{preview}</span>
            </p>

            <Button type="submit" disabled={saving}>
              {saving ? "Saving…" : "Save preferences"}
            </Button>
          </form>
        )}
      </SettingsSection>
    </>
  );
}

export function AccountRegionalSettingsPage() {
  return (
    <DashboardShell title="Profile & security" breadcrumb="Account settings">
      <SettingsPageBody>
        <AccountRegionalSettingsSection />
      </SettingsPageBody>
    </DashboardShell>
  );
}
