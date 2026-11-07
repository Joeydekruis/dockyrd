import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { LogStatusSummary } from "@/components/dockyard/ui/log-status-summary";

const meta = {
  title: "Design/Components/Log status summary",
  component: LogStatusSummary,
  tags: ["autodocs"],
  args: {
    counts: { "2xx": 1240, "4xx": 12, "5xx": 2 },
  },
} satisfies Meta<typeof LogStatusSummary>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
