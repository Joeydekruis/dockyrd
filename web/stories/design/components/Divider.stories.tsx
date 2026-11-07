import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Divider } from "@/components/dockyard/ui/divider";

/** **Source:** `web/components/dockyrd/ui/divider.tsx` */
const meta = {
  title: "Design/Components/Divider",
  component: Divider,
  tags: ["autodocs"],
  argTypes: { label: { control: "text" } },
  args: { label: "Or continue with" },
} satisfies Meta<typeof Divider>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
export const WithoutLabel: Story = {
  args: { label: undefined },
};
