import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { StatCard } from "@/components/dockyard/ui/stat-card";

const meta = {
  title: "Dockyrd/StatCard",
  component: StatCard,
  tags: ["autodocs"],
} satisfies Meta<typeof StatCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Applications: Story = {
  args: {
    label: "Applications",
    value: 9,
    hint: "Across 4 projects",
  },
};

export const Databases: Story = {
  args: {
    label: "Databases",
    value: 5,
  },
};
