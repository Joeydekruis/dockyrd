import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Backdrop } from "@/components/dockyrd/ui/backdrop";

const meta = {
  title: "Design/Components/Backdrop",
  component: Backdrop,
  tags: ["autodocs"],
  argTypes: {
    visible: { control: "boolean" },
    label: { control: "text" },
  },
  args: { visible: true, label: "Dismiss overlay" },
} satisfies Meta<typeof Backdrop>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const Demo: Story = {
  render: (args) => (
    <div
      className="relative h-48 rounded-xl border p-4"
      style={{ borderColor: "var(--border)" }}
    >
      <p className="text-sm">Content beneath overlay</p>
      <Backdrop {...args} onClick={() => {}} />
    </div>
  ),
};
