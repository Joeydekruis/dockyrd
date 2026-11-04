"use client";

import { useEffect, useState } from "react";
import type { ResourceProvisionContext } from "@/components/dockyard/resource-provisioning-modal";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import { Modal } from "@/components/dockyard/ui/modal";
import { SelectField } from "@/components/dockyard/ui/select-field";
import { createService, listProjects, type Project } from "@/lib/api";

type CreateApplicationModalProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (context: ResourceProvisionContext) => void;
  /** Pre-select project; hides the project picker when set. */
  defaultProjectId?: string;
};

export function CreateApplicationModal({
  open,
  onClose,
  onCreated,
  defaultProjectId,
}: CreateApplicationModalProps) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!open) return;
    setError(null);
    setName("");
    setProjectId(defaultProjectId ?? "");
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
        kind: "app",
        name: name.trim(),
        config: {
          source: { type: "image", ref: "nginx:alpine" },
          port: 80,
          autoDeploy: false,
        },
      });
      onCreated({
        kind: "application",
        serviceId: service.id,
        name: name.trim(),
        label: "Pulling image and starting the container.",
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
      title="Create application"
      description="Start with a container app in a project. Wire GitHub or a Dockerfile from the project page."
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
          <Button type="submit" form="create-application-form" disabled={busy}>
            {busy ? "Creating…" : "Create application"}
          </Button>
        </>
      }
    >
      <form
        id="create-application-form"
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
        <Field
          label="Application name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="web"
          hint="Used in URLs and the applications list."
        />
      </form>
    </Modal>
  );
}
