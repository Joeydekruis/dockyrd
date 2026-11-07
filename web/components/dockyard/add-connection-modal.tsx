"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import { Modal } from "@/components/dockyard/ui/modal";
import { SearchSelect } from "@/components/dockyard/ui/search-select";
import { Switch } from "@/components/dockyard/ui/switch";
import {
  createServiceLink,
  type LinkableCandidate,
  listLinkableServices,
} from "@/lib/api";

type AddConnectionModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  serviceId: string;
  providerMode?: boolean;
  onCreated?: () => void;
};

export function AddConnectionModal({
  open,
  onOpenChange,
  serviceId,
  providerMode = false,
  onCreated,
}: AddConnectionModalProps) {
  const [candidates, setCandidates] = useState<LinkableCandidate[]>([]);
  const [targetId, setTargetId] = useState("");
  const [addVars, setAddVars] = useState(true);
  const [enabledKeys, setEnabledKeys] = useState<Record<string, boolean>>({});
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    const res = await listLinkableServices(serviceId);
    setCandidates(res.candidates);
    if (res.candidates[0]) {
      setTargetId(res.candidates[0].service.id);
    }
  }, [serviceId]);

  useEffect(() => {
    if (!open) return;
    void load().catch(() => setCandidates([]));
  }, [open, load]);

  const selected = useMemo(
    () => candidates.find((c) => c.service.id === targetId),
    [candidates, targetId],
  );

  useEffect(() => {
    if (!selected) return;
    const next: Record<string, boolean> = {};
    for (const v of selected.suggested_variables) {
      next[v.key] = v.default_on;
    }
    setEnabledKeys(next);
  }, [selected]);

  const options = useMemo(
    () =>
      candidates.map((c) => ({
        value: c.service.id,
        label: `${c.service.name} · ${c.internal_host}`,
      })),
    [candidates],
  );

  async function onSubmit() {
    if (!selected) return;
    setBusy(true);
    setError(null);
    try {
      const variables = addVars
        ? selected.suggested_variables
            .filter((v) => enabledKeys[v.key])
            .map((v) => ({
              key: v.key,
              ref: v.ref,
              at_build: v.at_build,
              at_runtime: v.at_runtime,
              is_secret: v.is_secret,
            }))
        : [];
      await createServiceLink(
        serviceId,
        providerMode
          ? {
              from_service_id: selected.service.id,
              variables,
            }
          : {
              to_service_id: selected.service.id,
              variables,
            },
      );
      onOpenChange(false);
      onCreated?.();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not create connection");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Modal
      open={open}
      onClose={() => onOpenChange(false)}
      title="Add connection"
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            type="button"
            disabled={busy || !selected}
            onClick={() => void onSubmit()}
          >
            Add connection
          </Button>
        </>
      }
    >
      {candidates.length === 0 ? (
        <p className="text-[13px]" style={{ color: "var(--muted-text)" }}>
          No services available to link. Create a database, cache, or bucket in
          this project first.
        </p>
      ) : (
        <div className="space-y-4">
          <Field label="Service">
            <SearchSelect
              value={targetId}
              onChange={setTargetId}
              options={options}
              placeholder="Choose a service"
            />
          </Field>
          <div className="flex items-center justify-between gap-3">
            <span className="text-[13px]">
              Add environment variables to the application
            </span>
            <Switch checked={addVars} onChange={setAddVars} />
          </div>
          {addVars && selected ? (
            <ul
              className="space-y-2 rounded-md border p-3 text-[13px]"
              style={{ borderColor: "var(--border)" }}
            >
              {selected.suggested_variables.map((v) => (
                <li key={v.key} className="flex flex-wrap items-center gap-2">
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={enabledKeys[v.key] ?? false}
                      onChange={(e) =>
                        setEnabledKeys((prev) => ({
                          ...prev,
                          [v.key]: e.target.checked,
                        }))
                      }
                    />
                    <span className="font-mono text-[12px]">{v.key}</span>
                  </label>
                  <span
                    className="truncate font-mono text-[11px]"
                    style={{ color: "var(--muted-text)" }}
                  >
                    {v.ref}
                  </span>
                </li>
              ))}
            </ul>
          ) : null}
          <p className="text-[12px]" style={{ color: "var(--muted-text)" }}>
            Redeploy after saving so the app receives new variables.
          </p>
          {error ? (
            <p className="text-sm text-red-600" role="alert">
              {error}
            </p>
          ) : null}
        </div>
      )}
    </Modal>
  );
}
