import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";

const meta = {
  title: "Dockyrd/ResourceRow",
  component: ResourceRow,
  tags: ["autodocs"],
} satisfies Meta<typeof ResourceRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    name: "marketing",
    meta: "marketing.example.com",
    statusLabel: "Live",
    statusVariant: "ok",
    sha: "a1b2c3d4e5f6",
    commitMessage: "Fix cache headers",
  },
};

export const Deploying: Story = {
  args: {
    name: "api-worker",
    meta: "api.example.com",
    statusLabel: "Deploying",
    statusVariant: "info",
  },
};
