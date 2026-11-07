import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ChevronDown, Plus } from "lucide-react";
import type { ReactNode } from "react";
import { useState } from "react";
import {
  Alert,
  AuthField,
  Backdrop,
  Button,
  Checkbox,
  ConfirmDialog,
  Field,
  IconButton,
  Modal,
  PopoverMenu,
  ProgressBar,
  RadioGroup,
  SelectButton,
  SelectField,
  Skeleton,
  Snackbar,
  Spinner,
  SplitButton,
  Switch,
  TextareaField,
  Tooltip,
  UploadDropzone,
} from "@/components/dockyard/ui";
import { CatalogCanvas, CatalogRow } from "./catalog-layout";

const meta = {
  title: "Design/4 · Forms & overlays",
} satisfies Meta;

export default meta;

function Section({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) {
  return (
    <section
      className="flex flex-col gap-6 rounded-[20px] border p-6"
      style={{ borderColor: "var(--border)", background: "var(--surface)" }}
    >
      <div>
        <h2 className="text-2xl font-semibold tracking-tight">{title}</h2>
        {description ? (
          <p className="mt-1.5 text-sm" style={{ color: "var(--muted-text)" }}>
            {description}
          </p>
        ) : null}
      </div>
      {children}
    </section>
  );
}

export const SelectionAndChoice: StoryObj = {
  render: () => {
    const [segment, setSegment] = useState("production");
    const [plan, setPlan] = useState("standard");
    const [tags, setTags] = useState<string[]>(["web"]);
    const [notify, setNotify] = useState(true);
    const [autoDeploy, setAutoDeploy] = useState(false);
    return (
      <CatalogCanvas>
        <Section
          title="Selection & choice"
          description="SelectButton, radio, checkbox, select, autocomplete."
        >
          <SelectButton
            ariaLabel="Environment"
            options={[
              { value: "production", label: "Production" },
              { value: "staging", label: "Staging" },
            ]}
            value={segment}
            onChange={(v) => setSegment(v as string)}
          />
          <RadioGroup
            name="plan"
            legend="Plan"
            value={plan}
            onChange={setPlan}
            options={[
              {
                value: "standard",
                label: "Standard",
                description: "Single server",
              },
              {
                value: "ha",
                label: "High availability",
                description: "Multi-node",
              },
            ]}
          />
          <SelectButton
            ariaLabel="Tags"
            multi
            options={[
              { value: "web", label: "Web" },
              { value: "worker", label: "Worker" },
              { value: "cron", label: "Cron" },
            ]}
            value={tags}
            onChange={(v) => setTags(v as string[])}
          />
          <Checkbox
            label="Email me when deploy finishes"
            checked={notify}
            onChange={(e) => setNotify(e.target.checked)}
          />
          <Switch
            label="Auto-deploy on push"
            checked={autoDeploy}
            onChange={setAutoDeploy}
          />
          <SelectField
            label="Region"
            defaultValue="eu-central-1"
            options={[
              { value: "eu-central-1", label: "EU Central" },
              { value: "us-east-1", label: "US East" },
            ]}
            hint="Cannot be changed after first deploy."
          />
        </Section>
      </CatalogCanvas>
    );
  },
};

export const TextAndDateEntry: StoryObj = {
  render: () => {
    const [notes, setNotes] = useState("Rotate credentials after deploy.");
    return (
      <CatalogCanvas>
        <Section title="Text & date entry">
          <div className="max-w-md space-y-4">
            <AuthField
              label="Email"
              type="email"
              defaultValue="you@example.com"
            />
            <Field label="Service name" defaultValue="marketing" />
            <Field
              label="Deploy at"
              type="datetime-local"
              defaultValue="2026-11-04T14:00"
            />
            <TextareaField
              label="Release notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              maxLength={280}
              hint="Shown on the deployment timeline."
            />
          </div>
        </Section>
      </CatalogCanvas>
    );
  },
};

export const ButtonsWithMenus: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Buttons with menus">
        <CatalogRow>
          <SplitButton iconLeft={Plus} onMenuClick={() => undefined}>
            Deploy
          </SplitButton>
          <PopoverMenu
            items={[
              { id: "redeploy", label: "Redeploy" },
              { id: "rollback", label: "Rollback" },
              { id: "delete", label: "Delete service", destructive: true },
            ]}
            trigger={<IconButton icon={ChevronDown} label="Actions" />}
          />
        </CatalogRow>
      </Section>
    </CatalogCanvas>
  ),
};

export const Upload: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Upload">
        <div className="max-w-lg space-y-4">
          <UploadDropzone onBrowse={() => undefined} />
          <UploadDropzone state="uploading" />
          <UploadDropzone state="done" />
          <UploadDropzone state="error" errorMessage="File type not allowed." />
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const MessagesAndToasts: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Messages & toasts">
        <div className="max-w-xl space-y-6">
          <Alert
            tone="success"
            title="Domain verified"
            description="TLS certificate issued for marketing.example.com."
            onDismiss={() => undefined}
          />
          <Alert
            tone="error"
            title="Deployment failed"
            description="Build exited with code 1 — check logs for Dockerfile line 42."
            onDismiss={() => undefined}
          />
          <Snackbar
            tone="success"
            title="Settings saved"
            description="General settings were updated."
            progress={100}
            onDismiss={() => undefined}
          />
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const ProgressAndLoading: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Progress & loading">
        <div className="max-w-md space-y-6">
          <ProgressBar label="Deploy progress" value={62} />
          <ProgressBar label="Complete" variant="complete" />
          <ProgressBar label="Failed step" variant="error" value={40} />
          <CatalogRow>
            <Spinner label="Building image" />
            <Skeleton lines={2} />
          </CatalogRow>
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const Overlays: StoryObj = {
  render: () => {
    const [modalOpen, setModalOpen] = useState(true);
    const [confirmOpen, setConfirmOpen] = useState(false);
    return (
      <CatalogCanvas>
        <Section title="Overlays">
          <CatalogRow>
            <Tooltip
              placement="top"
              trigger={<Button variant="outline">Hover me</Button>}
              content="Restart container"
            />
            <Button variant="outline" onClick={() => setConfirmOpen(true)}>
              Confirm dialog
            </Button>
            <Button variant="outline" onClick={() => setModalOpen((v) => !v)}>
              Toggle modal
            </Button>
          </CatalogRow>
          {modalOpen ? (
            <>
              <Backdrop onClick={() => setModalOpen(false)} />
              <Modal
                open
                title="Update available"
                description="Running v0.0.27. Latest release: v0.0.28."
                onClose={() => setModalOpen(false)}
                footer={
                  <>
                    <Button>Go to Updates</Button>
                    <Button
                      variant="outline"
                      onClick={() => setModalOpen(false)}
                    >
                      Close
                    </Button>
                  </>
                }
              />
            </>
          ) : null}
          <ConfirmDialog
            open={confirmOpen}
            title="Delete service?"
            description="This stops the container and removes its volume."
            destructive
            confirmLabel="Delete"
            onConfirm={() => setConfirmOpen(false)}
            onCancel={() => setConfirmOpen(false)}
          />
        </Section>
      </CatalogCanvas>
    );
  },
  parameters: { layout: "fullscreen" },
};
