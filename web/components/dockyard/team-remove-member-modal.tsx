"use client";

import { useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Modal } from "@/components/dockyard/ui/modal";
import { TeamRoleBadge } from "@/components/dockyard/team-role-badge";
import { patchTeamUser, type TeamUser } from "@/lib/api";

type TeamRemoveMemberModalProps = {
  open: boolean;
  user: TeamUser | null;
  onClose: () => void;
  onRemoved: () => void | Promise<void>;
};

export function TeamRemoveMemberModal({
  open,
  user,
  onClose,
  onRemoved,
}: TeamRemoveMemberModalProps) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function confirm() {
    if (!user) return;
    setBusy(true);
    setError(null);
    try {
      await patchTeamUser(user.id, { disabled: true });
      await onRemoved();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not remove member");
    } finally {
      setBusy(false);
    }
  }

  if (!user) return null;

  return (
    <Modal
      open={open}
      title={`Remove ${user.name}?`}
      description={user.email}
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            type="button"
            variant="danger"
            disabled={busy}
            onClick={() => void confirm()}
          >
            {busy ? "Removing…" : "Remove member"}
          </Button>
        </>
      }
    >
      {error ? (
        <p className="mb-3 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
      <div className="mb-4">
        <TeamRoleBadge role={user.role} />
      </div>
      <ul className="list-none space-y-2.5 text-sm" style={{ color: "#3A4A63" }}>
        <li className="flex gap-2.5">
          <span
            className="mt-1.5 size-1.5 shrink-0 rounded-full"
            style={{ background: "#B3261E" }}
          />
          They lose access straight away and all their sessions and API tokens
          are revoked.
        </li>
        <li className="flex gap-2.5">
          <span
            className="mt-1.5 size-1.5 shrink-0 rounded-full"
            style={{ background: "#6B7686" }}
          />
          Their deployments, apps and databases stay as they are.
        </li>
        <li className="flex gap-2.5">
          <span
            className="mt-1.5 size-1.5 shrink-0 rounded-full"
            style={{ background: "#6B7686" }}
          />
          You can invite them again at any time.
        </li>
      </ul>
    </Modal>
  );
}
