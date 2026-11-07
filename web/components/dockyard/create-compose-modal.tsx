"use client";

import { useEffect, useState } from "react";
import type { ResourceProvisionContext } from "@/components/dockyrd/resource-provisioning-modal";
import { Button } from "@/components/dockyrd/ui/button";
import { Field } from "@/components/dockyrd/ui/field";
import { Modal } from "@/components/dockyrd/ui/modal";
import { SelectField } from "@/components/dockyrd/ui/select-field";
import { createService, listProjects, type Project } from "@/lib/api";

const DEFAULT_YAML = `services:
  web:
    image: nginx:alpine
    expose:
      - "80"
`;

type CreateComposeModalProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (context: ResourceProvisionContext) => void;
  defaultProjectId?: string;
};

export function CreateComposeModal({
  open,
  onClose,
  onCreated,
  defaultProjectId,
}: CreateComposeModalProps) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [name, setName] = useState("");
  const [yaml, setYaml] = useState(DEFAULT_YAML);
  const [validateMsg, setValidateMsg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!open) return;
    setError(null);
    setValidateMsg(null);
    setName("");
    setYaml(DEFAULT_YAML);
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
        kind: "compose",
        name: name.trim(),
        config: { source: { type: "inline", yaml } },
      });
      onCreated({
        kind: "application",
        serviceId: service.id,
        name: name.trim(),
        label: "Pulling images and starting the Compose stack.",
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
      title="Create Compose stack"
      description="Paste a docker-compose.yml (images only in v1). Host ports are stripped — use Dockyrd domains for HTTP."
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
          <Button type="submit" form="create-compose-form" disabled={busy}>
            {busy ? "Creating…" : "Create stack"}
          </Button>
        </>
      }
    >
      <form id="create-compose-form" className="space-y-4" onSubmit={onSubmit}>
        {!defaultProjectId ? (
          <SelectField
            label="Project"
            value={projectId}
            onChange={setProjectId}
            options={projects.map((p) => ({ value: p.id, label: p.name }))}
            placeholder="Select project"
          />
        ) : null}
        <Field
          label="Stack name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="My stack"
          autoComplete="off"
        />
        <Field label="Compose YAML">
          <textarea
            className="field-input min-h-[220px] font-mono text-xs"
            value={yaml}
            onChange={(e) => setYaml(e.target.value)}
            spellCheck={false}
          />
        </Field>
        {validateMsg ? (
          <p className="text-sm" style={{ color: "var(--muted-text)" }}>
            {validateMsg}
          </p>
        ) : null}
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
      </form>
    </Modal>
  );
}
