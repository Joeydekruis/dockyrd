import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { TokenSwatch } from "@/components/dockyard/ui/token-swatch";

/** **Source:** `web/components/dockyard/ui/token-swatch.tsx` */
const meta = {
  title: "Design/Components/Token swatch",
  component: TokenSwatch,
  tags: ["autodocs"],
  argTypes: {
    name: { control: "text" },
    hex: { control: "color" },
  },
  args: { name: "--ink-navy", hex: "#0b1f3a" },
} satisfies Meta<typeof TokenSwatch>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
