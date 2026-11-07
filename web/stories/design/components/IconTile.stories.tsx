import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Box, Settings } from "lucide-react";
import { IconTile } from "@/components/dockyrd/ui/icon-tile";

const iconArgType = {
  control: "select" as const,
  options: ["box", "settings"],
  mapping: { box: Box, settings: Settings },
};

const meta = {
  title: "Design/Components/Icon tile",
  component: IconTile,
  tags: ["autodocs"],
  argTypes: {
    icon: iconArgType,
    label: { control: "text" },
    size: { control: "select", options: ["sm", "md"] },
  },
  args: { icon: Box, label: "API service", size: "md" },
} satisfies Meta<typeof IconTile>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
