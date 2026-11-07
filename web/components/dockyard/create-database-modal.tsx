"use client";

import { useEffect, useState } from "react";
import type { ResourceProvisionContext } from "@/components/dockyrd/resource-provisioning-modal";
import { Button } from "@/components/dockyrd/ui/button";
import { Field } from "@/components/dockyrd/ui/field";
import { Modal } from "@/components/dockyrd/ui/modal";
import { SelectField } from "@/components/dockyrd/ui/select-field";
import { createService, listProjects, type Project } from "@/lib/api";

type CreateDatabaseModalProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (context: ResourceProvisionContext) => void;
  defaultProjectId?: string;
};

export function CreateDatabaseModal({
  open,
  onClose,
  onCreated,
  defaultProjectId,
}: CreateDatabaseModalProps) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [name, setName] = useState("");
  const [kind, setKind] = useState<"postgres" | "redis">("postgres");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!open) return;
    setError(null);
    setName("");
    setProjectId(defaultProjectId ?? "");
    setKind("postgres");
    void listProjects().then(({ projects: ps }) => setProjects(ps));
  }, [open, defaultProjectId]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!projectId || !name.trim()) {
      setError("Project and name are required.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const service = await createService({
        project_id: projectId,
        kind,
        name: name.trim(),
        config: kind === "postgres" ? { version: 17 } : { version: 8 },
      });
      onCreated({
        kind: "database",
        serviceId: service.id,
        name: name.trim(),
        label:
          kind === "postgres"
            ? "Postgres container starting on this server."
            : "Redis container starting on this server.",
      });
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
      title="Create database"
      description="Provision Postgres or Redis for a project."
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
          <Button type="submit" form="create-database-form" disabled={busy}>
            {busy ? "Creating…" : "Create database"}
          </Button>
        </>
      }
    >
      <form
        id="create-database-form"
        onSubmit={onSubmit}
        className="flex flex-col gap-4"
      >
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
        {defaultProjectId ? null : (
          <SelectField
            label="Project"
            placeholder="Select project"
            value={projectId}
            onChange={setProjectId}
            options={projects.map((p) => ({ value: p.id, label: p.name }))}
          />
        )}
        <SelectField
          label="Engine"
          value={kind}
          onChange={(v) => setKind(v as "postgres" | "redis")}
          options={[
            { value: "postgres", label: "Postgres 17" },
            { value: "redis", label: "Redis (Valkey 8)" },
          ]}
        />
        <Field
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Primary database"
          hint="Shown in your project and on this overview."
        />
      </form>
    </Modal>
  );
}
