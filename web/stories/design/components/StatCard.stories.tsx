import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { StatCard } from "@/components/dockyard/ui/stat-card";

/** **Source:** `web/components/dockyrd/ui/stat-card.tsx` */
const meta = {
  title: "Design/Components/Stat card",
  component: StatCard,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    value: { control: "text" },
    hint: { control: "text" },
  },
  args: { label: "Requests", value: "12.4k", hint: "+8% vs last week" },
} satisfies Meta<typeof StatCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
