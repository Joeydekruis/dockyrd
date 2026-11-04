import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Box, Home, Settings } from "lucide-react";
import { NavMenuItem } from "@/components/dockyard/ui/nav-menu-item";

const iconArgType = {
  control: "select" as const,
  options: ["home", "box", "settings"],
  mapping: { home: Home, box: Box, settings: Settings },
};

const meta = {
  title: "Design/Components/Nav menu item",
  component: NavMenuItem,
  tags: ["autodocs"],
  parameters: {
    nextjs: { appDirectory: true },
  },
  argTypes: {
    label: { control: "text" },
    href: { control: "text" },
    icon: iconArgType,
    active: { control: "boolean" },
  },
  args: {
    label: "Overview",
    href: "#",
    icon: Home,
    active: false,
  },
  decorators: [
    (Story) => (
      <nav className="w-56 p-4" style={{ background: "var(--surface)" }}>
        <Story />
      </nav>
    ),
  ],
} satisfies Meta<typeof NavMenuItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const Active: Story = {
  args: { active: true, label: "Services" },
};
