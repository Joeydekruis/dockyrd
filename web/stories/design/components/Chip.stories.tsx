import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Chip } from "@/components/dockyrd/ui/chip";

/** **Source:** `web/components/dockyrd/ui/chip.tsx` */
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
