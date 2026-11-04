import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Box, Globe } from "lucide-react";
import { ServiceCard } from "@/components/dockyard/ui/service-card";

const iconArgType = {
  control: "select" as const,
  options: ["box", "globe"],
  mapping: { box: Box, globe: Globe },
};

const meta = {
  title: "Design/Components/Service card",
  component: ServiceCard,
  tags: ["autodocs"],
  argTypes: {
    name: { control: "text" },
    domain: { control: "text" },
    statusLabel: { control: "text" },
    statusVariant: {
      control: "select",
      options: ["ok", "warn", "err", "idle", "info"],
    },
    icon: iconArgType,
  },
  args: {
    name: "web",
    domain: "app.example.com",
    statusLabel: "Running",
    statusVariant: "ok",
    icon: Box,
  },
} satisfies Meta<typeof ServiceCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
