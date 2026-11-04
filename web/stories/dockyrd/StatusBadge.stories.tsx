import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";

const meta = {
  title: "Dockyrd/StatusBadge",
  component: StatusBadge,
  tags: ["autodocs"],
} satisfies Meta<typeof StatusBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Healthy: Story = {
  args: { label: "Healthy", variant: "ok" },
};

export const Deploying: Story = {
  args: { label: "Deploying", variant: "info" },
};

export const Failing: Story = {
  args: { label: "Failing", variant: "err" },
};

export const Stopped: Story = {
  args: { label: "Stopped", variant: "idle" },
};

export const Warning: Story = {
  args: { label: "Degraded", variant: "warn" },
};
