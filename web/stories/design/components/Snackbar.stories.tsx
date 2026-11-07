import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Snackbar, SnackbarViewport } from "@/components/dockyrd/ui/snackbar";

const meta = {
  title: "Design/Components/Snackbar",
  component: Snackbar,
  tags: ["autodocs"],
  argTypes: {
    tone: {
      control: "select",
      options: ["info", "success", "warning", "error"],
    },
  },
  args: {
    tone: "success",
    title: "Domain verified",
    description: "TLS certificate issued for marketing.example.com.",
    progress: 62,
    onDismiss: () => {},
  },
} satisfies Meta<typeof Snackbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

/** Design board: toast variants */
export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex max-w-md flex-col gap-6 p-6">
      <div>
        <p
          className="mb-2 font-mono text-[10px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          Toast · success
        </p>
        <Snackbar
          tone="success"
          title="Deploy finished"
          description="marketing is live on the default URL."
          progress={100}
          onDismiss={() => undefined}
        />
      </div>
      <div>
        <p
          className="mb-2 font-mono text-[10px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          Toast · error + action
        </p>
        <Snackbar
          tone="error"
          title="Deployment failed"
          description="Build exited with code 1."
          actionLabel="Retry"
          onAction={() => undefined}
          progress={48}
          onDismiss={() => undefined}
        />
      </div>
      <div>
        <p
          className="mb-2 font-mono text-[10px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          Toast · info / update
        </p>
        <Snackbar
          tone="info"
          title="Update available"
          description="dockyrd v0.0.28 is ready to install."
          actionLabel="View"
          onAction={() => undefined}
          progress={72}
          onDismiss={() => undefined}
        />
      </div>
    </div>
  ),
};

export const Viewport: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="relative min-h-[220px]">
      <SnackbarViewport
        open
        tone="success"
        title="URL settings saved"
        onDismiss={() => undefined}
      />
    </div>
  ),
};
