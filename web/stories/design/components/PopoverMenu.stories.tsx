import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { PopoverMenu } from "@/components/dockyrd/ui/popover-menu";

const items = [
  { id: "edit", label: "Edit" },
  { id: "logs", label: "View logs" },
  { id: "delete", label: "Delete", destructive: true },
];

const meta = {
  title: "Design/Components/Popover menu",
  component: PopoverMenu,
  tags: ["autodocs"],
  args: { items, label: "Service actions" },
} satisfies Meta<typeof PopoverMenu>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  render: (args) => (
    <div className="flex justify-end p-12">
      <PopoverMenu {...args} />
    </div>
  ),
};
