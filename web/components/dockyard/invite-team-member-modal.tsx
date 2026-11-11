"use client";

import { Check, X } from "lucide-react";
import { useCallback, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Modal } from "@/components/dockyard/ui/modal";
import { createInvite, type TeamInvite } from "@/lib/api";
import {
  type AssignableTeamRole,
  INVITE_ROLE_OPTIONS,
  TEAM_ROLE_META,
  teamRoleLabel,
} from "@/lib/team-roles";

type InviteTeamMemberModalProps = {
  open: boolean;
  onClose: () => void;
  onInvited: () => void | Promise<void>;
};

type SentRow = {
  email: string;
  token: string | null;
  emailSent: boolean;
};

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function InviteTeamMemberModal({
  open,
  onClose,
  onInvited,
}: InviteTeamMemberModalProps) {
  const [method, setMethod] = useState<"email" | "link">("email");
  const [emails, setEmails] = useState<string[]>([]);
  const [draft, setDraft] = useState("");
  const [role, setRole] = useState<AssignableTeamRole>("member");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState<SentRow[] | null>(null);
  const [linkInvite, setLinkInvite] = useState<TeamInvite | null>(null);

  const reset = useCallback(() => {
    setMethod("email");
    setEmails([]);
    setDraft("");
    setRole("member");
    setBusy(false);
    setError(null);
    setSent(null);
    setLinkInvite(null);
  }, []);

  function handleClose() {
    reset();
    onClose();
  }

  function addEmailsFromText(text: string) {
    const parts = text
      .split(/[\s,;]+/)
      .map((x) => x.trim().toLowerCase())
      .filter(Boolean);
    if (parts.length === 0) return;
    setEmails((prev) => {
      const next = [...prev];
      for (const p of parts) {
        if (!next.includes(p)) next.push(p);
      }
      return next;
    });
    setDraft("");
  }

  function inviteUrl(token: string | null | undefined) {
    if (!token) return null;
    if (typeof window !== "undefined") {
      return `${window.location.origin}/accept-invite?token=${token}`;
    }
    return `/accept-invite?token=${token}`;
  }

  async function submitEmailInvites() {
    const valid = emails.filter((e) => EMAIL_RE.test(e));
    if (valid.length === 0) {
      setError("Add at least one valid email address.");
      return;
    }
    setBusy(true);
    setError(null);
    const rows: SentRow[] = [];
    try {
      for (const email of valid) {
        const inv = await createInvite({ email, role });
        rows.push({
          email,
          token: inv.token ?? null,
          emailSent: Boolean(inv.email_sent),
        });
      }
      setSent(rows);
      await onInvited();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invite failed");
    } finally {
      setBusy(false);
    }
  }

  async function createLinkInvite() {
    setBusy(true);
    setError(null);
    try {
      const inv = await createInvite({ role });
      setLinkInvite(inv);
      await onInvited();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not create link");
    } finally {
      setBusy(false);
    }
  }

  async function copyText(text: string) {
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      /* ignore */
    }
  }

  const invalidEmails = emails.filter((e) => !EMAIL_RE.test(e));
  const roleSummary = `${teamRoleLabel(role)} · All projects`;

  if (sent) {
    return (
      <Modal
        open={open}
        size="lg"
        title={
          sent.every((r) => r.emailSent) ? "Email sent" : "Invites created"
        }
        description="Each link is valid for 7 days. You can resend or revoke it from this page."
        onClose={handleClose}
        footer={
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setSent(null);
                setEmails([]);
              }}
            >
              Invite more
            </Button>
            <Button type="button" onClick={handleClose}>
              Done
            </Button>
          </>
        }
      >
        <ul
          className="divide-y rounded-xl border"
          style={{ borderColor: "var(--border)" }}
        >
          {sent.map((row) => {
            const url = inviteUrl(row.token);
            return (
              <li
                key={row.email}
                className="flex flex-wrap items-center gap-3 px-3.5 py-3"
              >
                <div className="min-w-0 flex-1">
                  <p className="truncate font-semibold">{row.email}</p>
                  <p className="text-sm" style={{ color: "var(--muted-text)" }}>
                    {roleSummary}
                  </p>
                </div>
                <span
                  className="inline-flex items-center gap-1.5 text-xs font-medium"
                  style={{ color: row.emailSent ? "#176B3D" : "#8A5200" }}
                >
                  <span
                    className="size-1.5 rounded-full"
                    style={{
                      background: row.emailSent ? "#1A7F4B" : "#D98A00",
                    }}
                  />
                  {row.emailSent ? "Email sent" : "Link only"}
                </span>
                {url ? (
                  <Button
                    type="button"
                    variant="outline"
                    className="min-h-10 px-3 text-[11px]"
                    onClick={() => void copyText(url)}
                  >
                    Copy link
                  </Button>
                ) : null}
              </li>
            );
          })}
        </ul>
        <p className="mt-3 text-sm" style={{ color: "var(--muted-text)" }}>
          Not in their inbox? Copy the link and send it yourself.
        </p>
      </Modal>
    );
  }

  return (
    <Modal
      open={open}
      size="lg"
      title="Invite team member"
      description="They get access as soon as they accept."
      onClose={handleClose}
      footer={
        method === "link" && linkInvite ? (
          <Button type="button" onClick={handleClose}>
            Done
          </Button>
        ) : (
          <>
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy}
              onClick={() =>
                void (method === "email"
                  ? submitEmailInvites()
                  : createLinkInvite())
              }
            >
              {busy
                ? "Working…"
                : method === "email"
                  ? "Send invites"
                  : "Generate link"}
            </Button>
          </>
        )
      }
    >
      <fieldset
        className="mb-5 flex gap-1 rounded-xl p-1 border-0 m-0 min-w-0"
        style={{ background: "#F1F0ED" }}
      >
        <legend className="sr-only">Invite method</legend>
        {(["email", "link"] as const).map((key) => {
          const active = method === key;
          return (
            <button
              key={key}
              type="button"
              aria-pressed={active}
              onClick={() => {
                setMethod(key);
                setLinkInvite(null);
                setError(null);
              }}
              className="min-h-10 flex-1 rounded-lg border-0 font-medium transition-shadow"
              style={{
                background: active ? "#FFFFFF" : "transparent",
                boxShadow: active ? "0 1px 2px rgba(11,31,58,0.08)" : "none",
              }}
            >
              {key === "email" ? "By email" : "Invite link"}
            </button>
          );
        })}
      </fieldset>

      {error ? (
        <p className="mb-3 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      {method === "email" ? (
        <div className="mb-5 space-y-2">
          <p className="text-sm font-semibold">Email addresses</p>
          <div
            className="flex min-h-12 flex-wrap items-center gap-1.5 rounded-[10px] border px-1.5 py-1.5"
            style={{ borderColor: "var(--border)" }}
          >
            {emails.map((email) => {
              const ok = EMAIL_RE.test(email);
              return (
                <span
                  key={email}
                  className="inline-flex items-center rounded-lg pl-2.5 text-sm font-medium"
                  style={{
                    background: ok ? "#F1F0ED" : "#FBE8E6",
                    color: ok ? "#0B1F3A" : "#B3261E",
                  }}
                >
                  {email}
                  <button
                    type="button"
                    aria-label={`Remove ${email}`}
                    className="inline-flex size-8 items-center justify-center rounded-lg border-0 bg-transparent"
                    onClick={() =>
                      setEmails((prev) => prev.filter((e) => e !== email))
                    }
                  >
                    <X className="size-3.5" strokeWidth={2.2} />
                  </button>
                </span>
              );
            })}
            <input
              className="field-input min-h-9 min-w-[180px] flex-1 border-0 bg-transparent shadow-none focus:ring-0"
              type="email"
              aria-label="Add an email address"
              placeholder="Add an email, press Enter"
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" || e.key === ",") {
                  e.preventDefault();
                  addEmailsFromText(draft);
                }
              }}
              onBlur={() => addEmailsFromText(draft)}
            />
          </div>
          {invalidEmails.length > 0 ? (
            <p className="text-sm text-red-600">
              Fix invalid addresses before sending.
            </p>
          ) : null}
          <p className="text-sm" style={{ color: "var(--muted-text)" }}>
            Paste several addresses at once, separated by commas.
          </p>
        </div>
      ) : linkInvite ? (
        <div className="mb-5 space-y-2">
          <p className="text-sm font-semibold">Invite link</p>
          <div className="flex gap-2">
            <input
              className="field-input min-h-12 flex-1 font-mono text-[13px]"
              readOnly
              aria-label="Invite link"
              value={inviteUrl(linkInvite.token) ?? ""}
            />
            <Button
              type="button"
              iconLeft={Check}
              onClick={() => {
                const url = inviteUrl(linkInvite.token);
                if (url) void copyText(url);
              }}
            >
              Copy link
            </Button>
          </div>
          <p className="text-sm" style={{ color: "var(--muted-text)" }}>
            Works once. Anyone with this link can join with the role above, so
            share it only with the person you mean.
          </p>
        </div>
      ) : null}

      {!(method === "link" && linkInvite) ? (
        <>
          <p className="mb-2 text-sm font-semibold">Role</p>
          <div
            className="mb-4 flex flex-col gap-2"
            role="radiogroup"
            aria-label="Role"
          >
            {INVITE_ROLE_OPTIONS.map((key) => {
              const meta = TEAM_ROLE_META[key];
              const selected = role === key;
              return (
                // biome-ignore lint/a11y/useSemanticElements: custom role card selector
                <button
                  key={key}
                  type="button"
                  role="radio"
                  aria-checked={selected}
                  onClick={() => setRole(key)}
                  className="flex w-full items-start gap-3 rounded-xl border p-3 text-left"
                  style={{
                    borderColor: selected ? "#3B93F5" : "var(--border)",
                    background: selected ? "#F4F9FF" : "var(--surface)",
                  }}
                >
                  <span
                    className="mt-0.5 inline-flex size-[18px] shrink-0 items-center justify-center rounded-full border-2"
                    style={{
                      borderColor: selected ? "#3B93F5" : "#9AA3B2",
                    }}
                  >
                    <span
                      className="size-2 rounded-full"
                      style={{
                        background: selected ? "#3B93F5" : "transparent",
                      }}
                    />
                  </span>
                  <span>
                    <span className="block font-semibold">{meta.label}</span>
                    <span
                      className="mt-0.5 block text-sm"
                      style={{ color: "var(--muted-text)" }}
                    >
                      {meta.description}
                    </span>
                  </span>
                </button>
              );
            })}
          </div>
          {role === "admin" ? (
            <p
              className="rounded-xl px-3.5 py-3 text-sm"
              style={{ background: "#F1F0ED", color: "#3A4A63" }}
            >
              Admins always have access to every project and all server
              settings.
            </p>
          ) : null}
        </>
      ) : null}
    </Modal>
  );
}
