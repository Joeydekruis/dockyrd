"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/dockyrd/ui/button";
import { Field } from "@/components/dockyrd/ui/field";
import { Modal } from "@/components/dockyrd/ui/modal";
import { TextareaField } from "@/components/dockyrd/ui/textarea-field";
import { createProject, type Project } from "@/lib/api";

type CreateProjectModalProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (project: Project) => void;
};

export function CreateProjectModal({
  open,
  onClose,
  onCreated,
}: CreateProjectModalProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!open) return;
    setName("");
    setDescription("");
    setError(null);
  }, [open]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) {
      setError("Project name is required.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const project = await createProject({
        name: trimmed,
        description: description.trim() || undefined,
      });
      onCreated(project);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Modal
      open={open}
      title="Create project"
      description="A project groups apps, databases, static sites, and buckets on your server."
      onClose={onClose}
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={onClose}
            disabled={busy}
          >
            Cancel
          </Button>
          <Button type="submit" form="create-project-form" disabled={busy}>
            {busy ? "Creating…" : "Create project"}
          </Button>
        </>
      }
    >
      <form
        id="create-project-form"
        onSubmit={onSubmit}
        className="flex flex-col gap-4"
      >
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
        <Field
          label="Project name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Client X"
          hint="Shown on the dashboard and in URLs."
          autoFocus
        />
        <TextareaField
          label="Description (optional)"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Production workloads for the marketing site and API."
        />
      </form>
    </Modal>
  );
}
