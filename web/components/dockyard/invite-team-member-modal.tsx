"use client";

import { useState } from "react";
import { Button } from "@/components/dockyrd/ui/button";
import { Field } from "@/components/dockyrd/ui/field";
import { Modal } from "@/components/dockyrd/ui/modal";
import { SelectField } from "@/components/dockyrd/ui/select-field";
import { createInvite } from "@/lib/api";

type InviteTeamMemberModalProps = {
  open: boolean;
  onClose: () => void;
  onInvited: () => void | Promise<void>;
};

type SuccessState = {
  token: string | null;
  emailSent: boolean;
  email: string;
};

export function InviteTeamMemberModal({
  open,
  onClose,
  onInvited,
}: InviteTeamMemberModalProps) {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("member");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<SuccessState | null>(null);

  function reset() {
    setEmail("");
    setRole("member");
    setBusy(false);
    setError(null);
    setSuccess(null);
  }

  function handleClose() {
    reset();
    onClose();
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = email.trim();
    if (!trimmed) {
      setError("Email is required.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const inv = await createInvite({ email: trimmed, role });
      setSuccess({
        token: inv.token ?? null,
        emailSent: Boolean(inv.email_sent),
        email: trimmed,
      });
      await onInvited();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invite failed");
    } finally {
      setBusy(false);
    }
  }

  const inviteUrl =
    success?.token && typeof window !== "undefined"
      ? `${window.location.origin}/accept-invite?token=${success.token}`
      : success?.token
        ? `/accept-invite?token=${success.token}`
        : null;

  const title = success ? "Invite sent" : "Invite team member";
  const description = success
    ? success.emailSent
      ? `We emailed an invite to ${success.email}.`
      : `Invite created for ${success.email}. Share the link below if email is not configured.`
    : "Send an invite link. They choose a password when accepting.";

  return (
    <Modal
      open={open}
      title={title}
      description={description}
      onClose={handleClose}
      footer={
        success ? (
          <Button type="button" onClick={handleClose}>
            Done
          </Button>
        ) : (
          <>
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              form="invite-team-member-form"
              disabled={busy}
            >
              {busy ? "Sending…" : "Send invite"}
            </Button>
          </>
        )
      }
    >
      {success ? (
        <div className="space-y-3">
          {success.emailSent ? (
            <p className="text-sm text-green-700" role="status">
              Invite email sent via SMTP.
            </p>
          ) : (
            <p className="text-sm" style={{ color: "var(--muted-text)" }}>
              SMTP not configured — copy the link below or configure outbound
              email under Settings → General.
            </p>
          )}
          {inviteUrl ? (
            <p className="text-sm" style={{ color: "var(--muted-text)" }}>
              Accept invite URL:{" "}
              <code className="break-all rounded bg-[var(--muted)]/30 px-1">
                {inviteUrl}
              </code>
            </p>
          ) : null}
        </div>
      ) : (
        <form
          id="invite-team-member-form"
          className="space-y-4"
          onSubmit={(e) => void submit(e)}
        >
          {error ? (
            <p className="text-sm text-red-600" role="alert">
              {error}
            </p>
          ) : null}
          <Field
            label="Email"
            type="email"
            required
            autoComplete="email"
            placeholder="colleague@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <SelectField
            label="Role"
            value={role}
            onChange={setRole}
            options={[
              { value: "admin", label: "Admin" },
              { value: "member", label: "Member" },
              { value: "viewer", label: "Viewer" },
            ]}
          />
        </form>
      )}
    </Modal>
  );
}
