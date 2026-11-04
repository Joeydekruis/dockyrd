import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Alert, type AlertTone } from "@/components/dockyard/ui/alert";
import { Button } from "@/components/dockyard/ui/button";

const tones: AlertTone[] = ["info", "success", "warning", "error"];

const meta = {
  title: "Design/Components/Alert",
  component: Alert,
  tags: ["autodocs"],
  argTypes: {
    tone: { control: "select", options: tones },
  },
  args: {
    tone: "info",
    title: "Build cache enabled",
    description: "Subsequent builds will reuse layers from the last deploy.",
    onDismiss: () => {},
  },
} satisfies Meta<typeof Alert>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

/** Design board: message banners */
export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex max-w-xl flex-col gap-4 p-6">
      <Alert
        tone="info"
        title="Build cache enabled"
        description="Subsequent builds reuse layers from the last deploy."
        onDismiss={() => undefined}
      />
      <Alert
        tone="success"
        title="Domain verified"
        description="TLS certificate issued for your custom hostname."
        onDismiss={() => undefined}
      />
      <Alert
        tone="warning"
        title="Disk almost full"
        description="Less than 10% free on the data volume."
        onDismiss={() => undefined}
      />
      <Alert
        tone="error"
        title="Deployment failed"
        description="Build exited with code 1 — check logs."
        onDismiss={() => undefined}
      />
    </div>
  ),
};

export const WithAction: Story = {
  args: {
    tone: "warning",
    title: "Certificate expires in 7 days",
    description: "Renew before traffic is interrupted.",
    action: <Button variant="outline">Renew</Button>,
  },
};
