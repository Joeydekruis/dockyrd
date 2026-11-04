import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Chip } from "@/components/dockyard/ui/chip";

/** **Source:** `web/components/dockyard/ui/chip.tsx` */
const meta = {
  title: "Design/Components/Chip",
  component: Chip,
  tags: ["autodocs"],
  argTypes: { children: { control: "text" } },
  args: { children: "Production" },
} satisfies Meta<typeof Chip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
