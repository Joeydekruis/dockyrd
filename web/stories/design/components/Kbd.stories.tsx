import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Kbd } from "@/components/dockyard/ui/kbd";

/** **Source:** `web/components/dockyrd/ui/kbd.tsx` */
const meta = {
  title: "Design/Components/Kbd",
  component: Kbd,
  tags: ["autodocs"],
  argTypes: { children: { control: "text" } },
  args: { children: "⌘ K" },
} satisfies Meta<typeof Kbd>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
