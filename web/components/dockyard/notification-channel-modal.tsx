"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { Modal } from "@/components/dockyard/ui/modal";
import { Button } from "@/components/dockyard/ui/button";
import {
  getNotificationSettings,
  type NotificationSettingsPayload,
} from "@/lib/api";

export type ChannelKind = "slack" | "discord" | "webhook" | "email";

type Props = {
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
  eventTypes: string[];
};

export function NotificationChannelModal({
  open,
  onClose,
  onSaved,
  eventTypes,
}: Props) {
  const [kind, setKind] = useState<ChannelKind>("slack");
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [signingSecret, setSigningSecret] = useState("");
  const [emailAdmins, setEmailAdmins] = useState(true);
  const [emailExtra, setEmailExtra] = useState("");
  const [smtpConfigured, setSmtpConfigured] = useState(true);
  const [minLevel, setMinLevel] = useState("warning");
  const [selectedTypes, setSelectedTypes] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [testResult, setTestResult] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setKind("slack");
    setName("");
    setUrl("");
    setSigningSecret("");
    setEmailAdmins(true);
    setEmailExtra("");
    setMinLevel("warning");
    setSelectedTypes(
      eventTypes.filter((t) => t !== "deploy.succeeded" && t !== "test.ping"),
    );
    setError(null);
    setTestResult(null);
    void getNotificationSettings().then((s: NotificationSettingsPayload) => {
      setSmtpConfigured(s.smtp_configured);
    });
  }, [open, eventTypes]);

  const toggleType = (t: string) => {
    setSelectedTypes((prev) =>
      prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t],
    );
  };

  const canSave =
    kind === "email"
      ? smtpConfigured && (emailAdmins || emailExtra.trim() !== "")
      : url.trim() !== "";

  const save = useCallback(async () => {
    setSaving(true);
    setError(null);
    try {
      const { createNotificationChannel, testNotificationChannel } =
        await import("@/lib/api");
      if (kind === "email") {
        const extra = emailExtra
          .split(/[\n,;]+/)
          .map((s) => s.trim())
          .filter(Boolean);
        const ch = await createNotificationChannel({
          kind: "email",
          name: name.trim() || "Email alerts",
          config: {
            recipients: {
              admins: emailAdmins,
              user_ids: [],
              extra,
            },
          },
          min_level: minLevel,
          event_types: selectedTypes,
        });
        const test = await testNotificationChannel(ch.id);
        if (test.ok) {
          setTestResult(`Delivered in ${test.latency_ms ?? 0} ms`);
        } else {
          setTestResult(test.error ?? "Test failed");
        }
        onSaved();
        onClose();
        return;
      }
      const secret: Record<string, unknown> = { url: url.trim() };
      if (kind === "webhook" && signingSecret.trim()) {
        secret.signing_secret = signingSecret.trim();
      }
      const ch = await createNotificationChannel({
        kind,
        name: name.trim() || kind,
        secret,
        min_level: minLevel,
        event_types: selectedTypes,
      });
      const test = await testNotificationChannel(ch.id);
      if (test.ok) {
        setTestResult(`Delivered in ${test.latency_ms ?? 0} ms`);
      } else {
        setTestResult(test.error ?? "Test failed");
      }
      onSaved();
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not save channel");
    } finally {
      setSaving(false);
    }
  }, [
    emailAdmins,
    emailExtra,
    kind,
    minLevel,
    name,
    onClose,
    onSaved,
    selectedTypes,
    signingSecret,
    url,
  ]);

  const kinds: ChannelKind[] = ["slack", "discord", "webhook", "email"];

  return (
    <Modal
      open={open}
      title="Add notification channel"
      description="Send deploy, backup, and server alerts outside the dashboard."
      onClose={onClose}
      size="lg"
    >
      <div className="mt-4 space-y-4 text-sm">
        <div className="flex flex-wrap gap-2">
          {kinds.map((k) => (
            <Button
              key={k}
              type="button"
              variant={kind === k ? "primary" : "outline"}
              onClick={() => setKind(k)}
            >
              {k}
            </Button>
          ))}
        </div>
        <label className="block space-y-1">
          <span className="text-[var(--muted-foreground)]">Name</span>
          <input
            className="w-full rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={kind === "email" ? "Ops email" : "Slack #ops"}
          />
        </label>
        {kind === "email" ? (
          <div className="space-y-3 rounded-lg border border-[var(--border)] p-4">
            {!smtpConfigured ? (
              <p className="text-[var(--destructive)]">
                Configure SMTP first under{" "}
                <Link
                  href="/settings/server/email"
                  className="underline"
                >
                  Server → Email
                </Link>
                .
              </p>
            ) : (
              <p className="text-[var(--muted-foreground)]">
                Uses instance SMTP. Deploy-failure inbox mail is deduplicated so
                admins are not emailed twice.
              </p>
            )}
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={emailAdmins}
                onChange={(e) => setEmailAdmins(e.target.checked)}
              />
              <span>All active owners and admins</span>
            </label>
            <label className="block space-y-1">
              <span className="text-[var(--muted-foreground)]">
                Extra addresses (one per line)
              </span>
              <textarea
                className="min-h-[72px] w-full rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2 font-mono text-xs"
                value={emailExtra}
                onChange={(e) => setEmailExtra(e.target.value)}
                placeholder="ops@example.com"
              />
            </label>
          </div>
        ) : (
          <>
            <label className="block space-y-1">
              <span className="text-[var(--muted-foreground)]">Webhook URL</span>
              <input
                className="w-full rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2 font-mono text-xs"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                autoComplete="off"
              />
            </label>
            {kind === "webhook" ? (
              <label className="block space-y-1">
                <span className="text-[var(--muted-foreground)]">
                  Signing secret
                </span>
                <input
                  className="w-full rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2 font-mono text-xs"
                  value={signingSecret}
                  onChange={(e) => setSigningSecret(e.target.value)}
                  autoComplete="off"
                />
              </label>
            ) : null}
          </>
        )}
        <label className="block space-y-1">
          <span className="text-[var(--muted-foreground)]">Minimum level</span>
          <select
            className="w-full rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
            value={minLevel}
            onChange={(e) => setMinLevel(e.target.value)}
          >
            <option value="info">Info</option>
            <option value="warning">Warning</option>
            <option value="error">Error</option>
          </select>
        </label>
        <fieldset>
          <legend className="mb-2 text-[var(--muted-foreground)]">
            Event types
          </legend>
          <div className="grid gap-2 sm:grid-cols-2">
            {eventTypes
              .filter((t) => t !== "test.ping")
              .map((t) => (
                <label key={t} className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={selectedTypes.includes(t)}
                    onChange={() => toggleType(t)}
                  />
                  <span className="font-mono text-xs">{t}</span>
                </label>
              ))}
          </div>
        </fieldset>
        {error ? (
          <p className="text-[var(--destructive)]" role="alert">
            {error}
          </p>
        ) : null}
        {testResult ? (
          <p className="text-[var(--muted-foreground)]">{testResult}</p>
        ) : null}
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="button" disabled={saving || !canSave} onClick={save}>
            Save channel
          </Button>
        </div>
      </div>
    </Modal>
  );
}
