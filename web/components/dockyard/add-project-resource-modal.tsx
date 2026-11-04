"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Modal } from "@/components/dockyard/ui/modal";
import { SearchSelectField } from "@/components/dockyard/ui/search-select-field";
import { moveServiceToProject, type Project } from "@/lib/api";
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
};

export function AddProjectResourceModal({
  open,
  project,
  onClose,
  onChanged,
}: AddProjectResourceModalProps) {
  const [kind, setKind] = useState<ProjectResourceKind>("app");
  const [candidates, setCandidates] = useState<
    Awaited<ReturnType<typeof loadMovableResources>>
  >([]);
  const [loadingCandidates, setLoadingCandidates] = useState(false);
  const [selectedId, setSelectedId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

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

  return (
    <Modal
      open={open}
      title="Add resource"
      description="Add an existing resource from another project to this project."
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
    </Modal>
  );
}
