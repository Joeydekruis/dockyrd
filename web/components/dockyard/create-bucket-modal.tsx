"use client";

import { useEffect, useState } from "react";
import type { ResourceProvisionContext } from "@/components/dockyard/resource-provisioning-modal";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import { Modal } from "@/components/dockyard/ui/modal";
import { SelectField } from "@/components/dockyard/ui/select-field";
import { createService, listProjects, type Project } from "@/lib/api";
import {
  NO_PROJECT_FIELD_HINT,
  projectSelectOptions,
} from "@/lib/create-project-field";
import { serviceDetailHref } from "@/lib/service-routes";

type CreateBucketModalProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (context: ResourceProvisionContext) => void;
  defaultProjectId?: string;
};

export function CreateBucketModal({
  open,
  onClose,
  onCreated,
  defaultProjectId,
}: CreateBucketModalProps) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [bucketName, setBucketName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!open) return;
    setError(null);
    setBucketName("");
    setProjectId(defaultProjectId ?? "");
    void listProjects().then(({ projects: ps }) => setProjects(ps));
  }, [open, defaultProjectId]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!bucketName.trim()) {
      setError("Bucket name is required.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const service = await createService({
        ...(projectId ? { project_id: projectId } : {}),
        kind: "bucket",
        name: bucketName.trim(),
        config: {
          bucket_name: bucketName.trim().toLowerCase().replace(/\s+/g, "-"),
        },
      });
      onCreated({
        kind: "bucket",
        serviceId: service.id,
        name: bucketName.trim(),
        label: "Registering bucket with object storage.",
        detailHref: serviceDetailHref(service.kind, service.slug),
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
      title="Create bucket"
      description="Add an S3-compatible bucket scoped to a project."
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
          <Button type="submit" form="create-bucket-form" disabled={busy}>
            {busy ? "Creating…" : "Create bucket"}
          </Button>
        </>
      }
    >
      <form
        id="create-bucket-form"
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
            placeholder="No project"
            value={projectId}
            onChange={setProjectId}
            options={projectSelectOptions(projects)}
            hint={NO_PROJECT_FIELD_HINT}
          />
        )}
        <Field
          label="Bucket name"
          value={bucketName}
          onChange={(e) => setBucketName(e.target.value)}
          placeholder="uploads"
          hint="Lowercase name used in the S3 API."
        />
      </form>
    </Modal>
  );
}
