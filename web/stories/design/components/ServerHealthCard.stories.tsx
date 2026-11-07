import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ServerHealthCard } from "@/components/dockyard/ui/server-health-card";

/** **Source:** `web/components/dockyrd/ui/server-health-card.tsx` */
const meta = {
  title: "Design/Components/Server health card",
  component: ServerHealthCard,
  tags: ["autodocs"],
  argTypes: {
    healthy: { control: "boolean" },
    version: { control: "text" },
    statusLine: { control: "text" },
  },
  args: { healthy: true, version: "v0.1.0", statusLine: "Up to date" },
} satisfies Meta<typeof ServerHealthCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
