import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { EmptyState } from "@/components/dockyard/ui/empty-state";

/** **Source:** `web/components/dockyrd/ui/empty-state.tsx` */
const meta = {
  title: "Design/Components/Empty state",
  component: EmptyState,
  tags: ["autodocs"],
  argTypes: {
    title: { control: "text" },
    description: { control: "text" },
    actionLabel: { control: "text" },
    actionHref: { control: "text" },
  },
  args: {
    title: "No services yet",
    description: "Create your first service from Git or a template.",
    actionLabel: "Create service",
    actionHref: "#",
  },
} satisfies Meta<typeof EmptyState>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
