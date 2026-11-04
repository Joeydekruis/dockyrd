import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ArrowRight } from "lucide-react";
import { InlineLink } from "@/components/dockyard/ui/inline-link";

const iconArgType = {
  control: "select" as const,
  options: ["none", "arrowRight"],
  mapping: { none: undefined, arrowRight: ArrowRight },
};

const meta = {
  title: "Design/Components/Inline link",
  component: InlineLink,
  tags: ["autodocs"],
  argTypes: {
    href: { control: "text" },
    external: { control: "boolean" },
    iconRight: iconArgType,
    children: { control: "text" },
  },
  args: {
    href: "https://github.com/Joeydekruis/dockyrd",
    children: "View repository",
    external: true,
    iconRight: undefined,
  },
} satisfies Meta<typeof InlineLink>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
