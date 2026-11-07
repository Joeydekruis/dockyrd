"use client";

import { useState } from "react";
import { Button } from "@/components/dockyrd/ui/button";
import { Modal } from "@/components/dockyrd/ui/modal";
import { Switch } from "@/components/dockyrd/ui/switch";

type RemoveConnectionModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  peerName: string;
  variablesCount: number;
  onConfirm: (removeLinkedVariables: boolean) => Promise<void>;
};

export function RemoveConnectionModal({
  open,
  onOpenChange,
  peerName,
  variablesCount,
  onConfirm,
}: RemoveConnectionModalProps) {
  const [removeVars, setRemoveVars] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit() {
    setBusy(true);
    setError(null);
    try {
      await onConfirm(removeVars);
      onOpenChange(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not remove connection");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Modal
      open={open}
      title="Remove connection"
      description={`Stop linking ${peerName} to this application.`}
      onClose={() => onOpenChange(false)}
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={busy}
          >
            Cancel
          </Button>
          <Button type="button" disabled={busy} onClick={() => void onSubmit()}>
            {busy ? "Removing…" : "Remove connection"}
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Peers will no longer be reachable via internal DNS from this app.
          Deploy again after removing references from environment variables.
        </p>
        {variablesCount > 0 ? (
          <Switch
            label={`Remove ${variablesCount === 1 ? "1 linked variable" : `${variablesCount} linked variables`}`}
            checked={removeVars}
            onChange={setRemoveVars}
          />
        ) : null}
      </div>
    </Modal>
  );
}
