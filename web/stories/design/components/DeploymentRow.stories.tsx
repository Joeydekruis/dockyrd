import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { DeploymentRow } from "@/components/dockyard/ui/deployment-row";

/** **Source:** `web/components/dockyrd/ui/deployment-row.tsx` */
const meta = {
  title: "Design/Components/Deployment row",
  component: DeploymentRow,
  tags: ["autodocs"],
  argTypes: {
    name: { control: "text" },
    statusLabel: { control: "text" },
    statusVariant: {
      control: "select",
      options: ["ok", "warn", "err", "idle", "info"],
    },
    sha: { control: "text" },
    commitMessage: { control: "text" },
    meta: { control: "text" },
  },
  args: {
    name: "Production deploy",
    statusLabel: "Live",
    statusVariant: "ok",
    sha: "f4e2a91",
    commitMessage: "feat: metrics export",
    meta: "2 min ago",
  },
} satisfies Meta<typeof DeploymentRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
