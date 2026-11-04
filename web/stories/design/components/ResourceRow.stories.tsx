import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Box, Database } from "lucide-react";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";

const iconArgType = {
  control: "select" as const,
  options: ["box", "database"],
  mapping: { box: Box, database: Database },
};

const meta = {
  title: "Design/Components/Resource row",
  component: ResourceRow,
  tags: ["autodocs"],
  argTypes: {
    name: { control: "text" },
    meta: { control: "text" },
    statusLabel: { control: "text" },
    statusVariant: {
      control: "select",
      options: ["ok", "warn", "err", "idle", "info"],
    },
    sha: { control: "text" },
    commitMessage: { control: "text" },
    icon: iconArgType,
  },
  args: {
    name: "web",
    meta: "256 MiB · 1 replica",
    statusLabel: "Running",
    statusVariant: "ok",
    sha: "c0ffee1",
    commitMessage: "chore: bump deps",
    icon: Box,
  },
} satisfies Meta<typeof ResourceRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
