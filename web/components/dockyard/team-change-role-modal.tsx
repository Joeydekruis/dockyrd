"use client";

import { useEffect, useState } from "react";
import { TeamRoleBadge } from "@/components/dockyard/team-role-badge";
import { Button } from "@/components/dockyard/ui/button";
import { Modal } from "@/components/dockyard/ui/modal";
import { patchTeamUser, type TeamUser } from "@/lib/api";
import { initialsFromName } from "@/lib/team-format";
import {
  type AssignableTeamRole,
  INVITE_ROLE_OPTIONS,
  TEAM_ROLE_META,
} from "@/lib/team-roles";

type TeamChangeRoleModalProps = {
  open: boolean;
  user: TeamUser | null;
  onClose: () => void;
  onSaved: () => void | Promise<void>;
};

export function TeamChangeRoleModal({
  open,
  user,
  onClose,
  onSaved,
}: TeamChangeRoleModalProps) {
  const [role, setRole] = useState<AssignableTeamRole>("member");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (user && INVITE_ROLE_OPTIONS.includes(user.role as AssignableTeamRole)) {
      setRole(user.role as AssignableTeamRole);
    } else {
      setRole("member");
    }
    setError(null);
  }, [user]);

  async function save() {
    if (!user) return;
    setBusy(true);
    setError(null);
    try {
      await patchTeamUser(user.id, { role });
      await onSaved();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save role");
    } finally {
      setBusy(false);
    }
  }

  if (!user) return null;

  return (
    <Modal
      open={open}
      size="lg"
      title={`Change role for ${user.name}`}
      description={user.email}
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            type="button"
            disabled={busy || user.role === "owner"}
            onClick={() => void save()}
          >
            {busy ? "Saving…" : "Save changes"}
          </Button>
        </>
      }
    >
      {error ? (
        <p className="mb-3 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
      <div className="mb-4 flex items-center gap-3">
        <span
          className="inline-flex size-11 shrink-0 items-center justify-center rounded-full text-sm font-semibold"
          style={{ background: "#E3F0FF", color: "#0B57C9" }}
        >
          {initialsFromName(user.name)}
        </span>
        <TeamRoleBadge role={user.role} />
      </div>
      <p className="mb-2 text-sm font-semibold">Role</p>
      <div className="flex flex-col gap-2" role="radiogroup" aria-label="Role">
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
              className="flex w-full items-start gap-3 rounded-xl border p-3 text-left transition-colors"
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
          className="mt-4 rounded-xl px-3.5 py-3 text-sm"
          style={{ background: "#F1F0ED", color: "#3A4A63" }}
        >
          Admins always have access to every project and all server settings.
        </p>
      ) : null}
    </Modal>
  );
}
