import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Button } from "@/components/dockyard/ui/button";

const meta = {
  title: "Dockyrd/Button",
  component: Button,
  tags: ["autodocs"],
  argTypes: {
    variant: { control: "select", options: ["primary", "outline"] },
  },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: {
    children: "Create application",
  },
};

export const Outline: Story = {
  args: {
    variant: "outline",
    children: "Docs",
  },
};

export const Disabled: Story = {
  args: {
    children: "Saving…",
    disabled: true,
  },
};

export const AsLink: Story = {
  args: {
    href: "https://github.com/Joeydekruis/dockyrd",
    variant: "outline",
    children: "View on GitHub",
  },
};
