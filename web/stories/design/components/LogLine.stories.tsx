import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { LogLine } from "@/components/dockyard/ui/log-line";

/** **Source:** `web/components/dockyrd/ui/log-line.tsx` */
const meta = {
  title: "Design/Components/Log line",
  component: LogLine,
  tags: ["autodocs"],
  argTypes: {
    timestamp: { control: "text" },
    status: { control: "number" },
    method: { control: "text" },
    path: { control: "text" },
  },
  args: {
    timestamp: "2026-11-04T10:00:00Z",
    status: 200,
    method: "GET",
    path: "/api/v1/health",
  },
} satisfies Meta<typeof LogLine>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
