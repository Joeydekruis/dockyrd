"use client";

import { Link2, Settings } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import { Modal } from "@/components/dockyard/ui/modal";
import {
  SettingsCopyPill,
  SettingsDetailBlock,
  SettingsEnabledStatus,
  SettingsGroupHeading,
} from "@/components/dockyard/ui/settings-detail-block";
import { Switch } from "@/components/dockyard/ui/switch";
import { TextareaField } from "@/components/dockyard/ui/textarea-field";
import { type Project, patchProject } from "@/lib/api";
import { useSnackbar } from "@/lib/snackbar-context";

type EditModal = null | "name" | "description" | "links";

function cardClassName() {
  return "rounded-[20px] border px-6 pb-2 pt-4";
}

type ProjectSettingsTabProps = {
  project: Project;
  onUpdated: () => void;
  onError: (message: string | null) => void;
};

export function ProjectSettingsTab({
  project,
  onUpdated,
  onError,
}: ProjectSettingsTabProps) {
  const { showSnackbar } = useSnackbar();
  const strict = project.links_mode === "strict";
  const [modal, setModal] = useState<EditModal>(null);
  const [busy, setBusy] = useState(false);

  const [name, setName] = useState(project.name);
  const [description, setDescription] = useState(project.description ?? "");
  const [strictLinks, setStrictLinks] = useState(strict);

  useEffect(() => {
    setName(project.name);
  }, [project.name]);

  useEffect(() => {
    setDescription(project.description ?? "");
  }, [project.description]);

  useEffect(() => {
    setStrictLinks(project.links_mode === "strict");
  }, [project.links_mode]);

  function closeModal() {
    if (busy) return;
    setModal(null);
  }

  async function save(
    body: Parameters<typeof patchProject>[1],
    successMessage: string,
  ) {
    setBusy(true);
    onError(null);
    try {
      await patchProject(project.id, body);
      onUpdated();
      showSnackbar(successMessage, "success");
      setModal(null);
    } catch (e) {
      onError(
        e instanceof Error ? e.message : "Could not save project settings",
      );
    } finally {
      setBusy(false);
    }
  }

  const descriptionDisplay = project.description?.trim()
    ? project.description.trim()
    : "—";

  return (
    <div className="space-y-8">
      <div>
        <SettingsGroupHeading icon={Settings} label="General" />
        <section
          className={cardClassName()}
          style={{
            borderColor: "var(--border)",
            background: "var(--surface)",
          }}
          aria-label="General"
        >
          <SettingsDetailBlock
            title="Details"
            description="How the project appears in the dashboard and resource lists."
            actionLabel="Update name"
            rows={[
              { label: "Name", value: project.name },
              {
                label: "Slug",
                value: <SettingsCopyPill value={project.slug} />,
              },
              {
                label: "Project ID",
                value: <SettingsCopyPill value={project.id} />,
                span: 2,
              },
            ]}
            onAction={() => setModal("name")}
          />
          <SettingsDetailBlock
            title="Description"
            description="Optional note for your team. Not shown on deployed resources."
            actionLabel="Update description"
            rows={[
              { label: "Description", value: descriptionDisplay, span: 3 },
            ]}
            onAction={() => setModal("description")}
          />
        </section>
      </div>

      <div>
        <SettingsGroupHeading icon={Link2} label="Networking" />
        <section
          className={cardClassName()}
          style={{
            borderColor: "var(--border)",
            background: "var(--surface)",
          }}
          aria-label="Service links"
        >
          <SettingsDetailBlock
            title="Service links"
            description="In strict mode, only linked services can reach each other on the Docker network. Reference variables always require an explicit link."
            actionLabel="Update link policy"
            rows={[
              {
                label: "Strict links",
                value: <SettingsEnabledStatus enabled={strict} />,
              },
              {
                label: "Mode",
                value: (
                  <span className="font-mono text-[13px]">
                    {strict
                      ? "Enforced"
                      : "Not enforced (legacy open project network)"}
                  </span>
                ),
                span: 2,
              },
            ]}
            onAction={() => setModal("links")}
          />
        </section>
      </div>

      <Modal
        open={modal === "name"}
        title="Update project name"
        description="Shown in the sidebar, project list, and breadcrumbs. The URL slug does not change."
        onClose={closeModal}
        footer={
          <>
            <Button type="button" variant="outline" onClick={closeModal}>
              Cancel
            </Button>
            <Button
              type="button"
              loading={busy}
              disabled={!name.trim() || name.trim() === project.name}
              onClick={() =>
                void save({ name: name.trim() }, "Project name saved.")
              }
            >
              Save
            </Button>
          </>
        }
      >
        <Field
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          autoFocus
          maxLength={100}
        />
      </Modal>

      <Modal
        open={modal === "description"}
        title="Update description"
        description="Optional internal note for this project."
        onClose={closeModal}
        footer={
          <>
            <Button type="button" variant="outline" onClick={closeModal}>
              Cancel
            </Button>
            <Button
              type="button"
              loading={busy}
              onClick={() =>
                void save(
                  { description: description.trim() },
                  "Description saved.",
                )
              }
            >
              Save
            </Button>
          </>
        }
      >
        <TextareaField
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={4}
          placeholder="What this project is for…"
        />
      </Modal>

      <Modal
        open={modal === "links"}
        title="Update link policy"
        description="Controls whether unlinked services on the project network can talk to each other."
        onClose={closeModal}
        footer={
          <>
            <Button type="button" variant="outline" onClick={closeModal}>
              Cancel
            </Button>
            <Button
              type="button"
              loading={busy}
              disabled={strictLinks === strict}
              onClick={() =>
                void save(
                  { links_mode: strictLinks ? "strict" : "open" },
                  "Link policy saved.",
                )
              }
            >
              Save
            </Button>
          </>
        }
      >
        <Switch
          label="Strict links (network enforced)"
          checked={strictLinks}
          onChange={setStrictLinks}
        />
      </Modal>
    </div>
  );
}
