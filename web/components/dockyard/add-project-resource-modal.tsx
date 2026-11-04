"use client";

import { Plus } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CreateStaticSiteWizard } from "@/components/dockyard/create-static-site-wizard";
import type { StaticSiteCreateResult } from "@/components/dockyard/static-site-create-progress-modal";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import { Modal } from "@/components/dockyard/ui/modal";
import { SearchSelectField } from "@/components/dockyard/ui/search-select-field";
import { createService, moveServiceToProject, type Project } from "@/lib/api";
import {
  loadMovableResources,
  movableToSelectOptions,
  PROJECT_RESOURCE_KINDS,
  type ProjectResourceKind,
} from "@/lib/project-resource-catalog";

type AddProjectResourceModalProps = {
  open: boolean;
  project: Project | null;
  onClose: () => void;
  onChanged: () => void;
  onStaticSiteCreated?: (result: StaticSiteCreateResult) => void;
};

export function AddProjectResourceModal({
  open,
  project,
  onClose,
  onChanged,
  onStaticSiteCreated,
}: AddProjectResourceModalProps) {
  const [kind, setKind] = useState<ProjectResourceKind>("app");
  const [candidates, setCandidates] = useState<
    Awaited<ReturnType<typeof loadMovableResources>>
  >([]);
  const [loadingCandidates, setLoadingCandidates] = useState(false);
  const [selectedId, setSelectedId] = useState("");
  const [newName, setNewName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [staticWizardOpen, setStaticWizardOpen] = useState(false);

  const refreshCandidates = useCallback(async () => {
    if (!project) {
      setCandidates([]);
      return;
    }
    setLoadingCandidates(true);
    try {
      const rows = await loadMovableResources(kind, project.id);
      setCandidates(rows);
      setSelectedId("");
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load resources");
      setCandidates([]);
    } finally {
      setLoadingCandidates(false);
    }
  }, [kind, project]);

  useEffect(() => {
    if (!open) return;
    setKind("app");
    setNewName("");
    setSelectedId("");
    setError(null);
  }, [open]);

  useEffect(() => {
    if (!open || !project) return;
    void refreshCandidates();
  }, [open, project, refreshCandidates]);

  const selectOptions = useMemo(
    () => movableToSelectOptions(candidates),
    [candidates],
  );

  const kindDef = PROJECT_RESOURCE_KINDS.find((k) => k.id === kind);

  async function onMoveExisting() {
    if (!project || !selectedId) return;
    setBusy(true);
    setError(null);
    try {
      await moveServiceToProject(selectedId, project.id);
      onChanged();
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not add resource");
    } finally {
      setBusy(false);
    }
  }

  async function onCreateNew(e: React.FormEvent) {
    e.preventDefault();
    if (!project) return;
    if (kind === "static_site") {
      setStaticWizardOpen(true);
      return;
    }
    const name = newName.trim();
    if (!name) {
      setError("Name is required.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const config =
        kind === "app"
          ? {
              source: { type: "image", ref: "nginx:alpine" },
              port: 80,
              autoDeploy: false,
            }
          : kind === "postgres"
            ? { version: 17 }
            : { bucket_name: name, public_read: false };
      await createService({
        project_id: project.id,
        kind:
          kind === "postgres"
            ? "postgres"
            : kind === "bucket"
              ? "bucket"
              : "app",
        name,
        config,
      });
      onChanged();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <Modal
        open={open && !staticWizardOpen}
        title="Add resource"
        description="Move an existing resource into this project or create a new one."
        onClose={onClose}
        footer={
          <>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy || !selectedId}
              onClick={() => void onMoveExisting()}
            >
              Add selected
            </Button>
          </>
        }
      >
        <fieldset className="border-0 p-0">
          <legend
            className="mb-3 font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Resource type
          </legend>
          <div className="flex flex-wrap gap-2">
            {PROJECT_RESOURCE_KINDS.map((item) => {
              const selected = item.id === kind;
              const Icon = item.icon;
              return (
                <button
                  key={item.id}
                  type="button"
                  aria-pressed={selected}
                  aria-label={item.label}
                  onClick={() => setKind(item.id)}
                  className="flex size-11 items-center justify-center rounded-full border-2 transition-colors"
                  style={{
                    borderColor: selected
                      ? "var(--accent-blue)"
                      : "var(--border)",
                    background: selected ? "var(--brand-soft)" : "transparent",
                    color: selected
                      ? "var(--link-blue)"
                      : "var(--secondary-text)",
                  }}
                >
                  <Icon strokeWidth={1.8} className="size-5" aria-hidden />
                </button>
              );
            })}
          </div>
          <p className="mt-2 text-sm" style={{ color: "var(--muted-text)" }}>
            {kindDef?.label}
          </p>
        </fieldset>

        {error ? (
          <p className="mt-4 text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}

        <div className="mt-6">
          <SearchSelectField
            label="Existing resource"
            options={selectOptions}
            value={selectedId}
            onChange={setSelectedId}
            placeholder={
              loadingCandidates
                ? "Loading…"
                : selectOptions.length
                  ? "Search resources…"
                  : "No matching resources elsewhere"
            }
            searchPlaceholder="Filter by name or project"
            emptyMessage="No resources in other projects"
            disabled={loadingCandidates || selectOptions.length === 0}
          />
        </div>

        <form
          className="mt-8 border-t pt-6"
          style={{ borderColor: "var(--border)" }}
          onSubmit={onCreateNew}
        >
          <p
            className="mb-3 font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Or create new
          </p>
          {kind === "static_site" ? (
            <Button
              type="submit"
              iconLeft={Plus}
              disabled={busy}
              className="w-full sm:w-auto"
            >
              Open static site wizard
            </Button>
          ) : (
            <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
              <Field
                label="Name"
                name="resource-name"
                value={newName}
                onChange={(ev) => setNewName(ev.target.value)}
                placeholder={kind === "postgres" ? "Primary database" : "api"}
                className="flex-1"
              />
              <Button type="submit" disabled={busy} iconLeft={Plus}>
                Create
              </Button>
            </div>
          )}
        </form>
      </Modal>

      <CreateStaticSiteWizard
        open={staticWizardOpen}
        defaultProjectId={project?.id}
        onClose={() => {
          setStaticWizardOpen(false);
          onClose();
        }}
        onCreated={(result) => {
          setStaticWizardOpen(false);
          onStaticSiteCreated?.(result);
          onChanged();
          onClose();
        }}
      />
    </>
  );
}
