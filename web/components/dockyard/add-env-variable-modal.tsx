"use client";

import { useMemo, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import { Modal } from "@/components/dockyard/ui/modal";
import { Switch } from "@/components/dockyard/ui/switch";
import { Tabs } from "@/components/dockyard/ui/tabs";
import { parseEnvPaste } from "@/lib/parse-env";

export type NewEnvVariable = {
  key: string;
  value: string;
  is_secret: boolean;
  at_build: boolean;
  at_runtime: boolean;
};

type AddEnvVariableModalProps = {
  open: boolean;
  existingKeys: string[];
  onClose: () => void;
  onAdd: (entries: NewEnvVariable[]) => void | Promise<void>;
};

export function AddEnvVariableModal({
  open,
  existingKeys,
  onClose,
  onAdd,
}: AddEnvVariableModalProps) {
  const [mode, setMode] = useState<"single" | "paste">("single");
  const [key, setKey] = useState("");
  const [value, setValue] = useState("");
  const [isSecret, setIsSecret] = useState(false);
  const [paste, setPaste] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const existing = useMemo(
    () => new Set(existingKeys.map((k) => k.toUpperCase())),
    [existingKeys],
  );

  function reset() {
    setKey("");
    setValue("");
    setIsSecret(false);
    setPaste("");
    setError(null);
    setMode("single");
  }

  function handleClose() {
    reset();
    onClose();
  }

  async function submitSingle() {
    const k = key.trim().toUpperCase();
    if (!k) {
      setError("Key is required.");
      return;
    }
    if (existing.has(k)) {
      setError(`Variable ${k} already exists.`);
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await onAdd([
        {
          key: k,
          value,
          is_secret: isSecret,
          at_build: true,
          at_runtime: false,
        },
      ]);
      reset();
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not save variable");
    } finally {
      setBusy(false);
    }
  }

  async function submitPaste() {
    const parsed = parseEnvPaste(paste);
    if (parsed.length === 0) {
      setError("No valid KEY=VALUE lines found.");
      return;
    }
    const fresh = parsed.filter((p) => !existing.has(p.key));
    if (fresh.length === 0) {
      setError("All keys in the paste already exist.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await onAdd(
        fresh.map((p) => ({
          key: p.key,
          value: p.value,
          is_secret: p.isSecret,
          at_build: true,
          at_runtime: false,
        })),
      );
      reset();
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not save variables");
    } finally {
      setBusy(false);
    }
  }

  const previewCount =
    mode === "paste"
      ? parseEnvPaste(paste).filter((p) => !existing.has(p.key)).length
      : 0;

  return (
    <Modal
      open={open}
      title="Add variable"
      description="Build-time variables are injected during the static site build."
      onClose={handleClose}
      size="lg"
      footer={
        <>
          <Button type="button" variant="outline" onClick={handleClose}>
            Cancel
          </Button>
          <Button
            type="button"
            disabled={busy}
            onClick={() =>
              void (mode === "single" ? submitSingle() : submitPaste())
            }
          >
            {mode === "paste" && previewCount > 1
              ? `Add ${previewCount} variables`
              : "Add variable"}
          </Button>
        </>
      }
    >
      <Tabs
        activeId={mode}
        onChange={(id) => {
          setMode(id as "single" | "paste");
          setError(null);
        }}
        tabs={[
          { id: "single", label: "Single variable" },
          { id: "paste", label: "Paste .env" },
        ]}
      />
      {error ? (
        <p className="mt-3 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
      {mode === "single" ? (
        <div className="mt-4 space-y-4">
          <Field
            label="Key"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            placeholder="PUBLIC_API_URL"
          />
          <Field
            label="Value"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder="https://api.example.com"
          />
          <Switch
            label="Treat as secret"
            checked={isSecret}
            onChange={setIsSecret}
          />
        </div>
      ) : (
        <div className="mt-4 space-y-2">
          <label className="flex flex-col gap-1 text-sm">
            <span
              className="font-mono text-[11px] uppercase tracking-widest"
              style={{ color: "var(--muted-text)" }}
            >
              .env contents
            </span>
            <textarea
              className="field-input min-h-[160px] w-full font-mono text-[13px]"
              value={paste}
              onChange={(e) => setPaste(e.target.value)}
              placeholder={
                "PUBLIC_API_URL=https://api.example.com\nCMS_TOKEN=secret"
              }
            />
          </label>
          {previewCount > 0 ? (
            <p className="text-[13px]" style={{ color: "var(--muted-text)" }}>
              {previewCount} new variable{previewCount === 1 ? "" : "s"} will be
              added.
            </p>
          ) : null}
        </div>
      )}
    </Modal>
  );
}
