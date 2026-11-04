import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Avatar } from "@/components/dockyard/ui/avatar";

/** **Source:** `web/components/dockyard/ui/avatar.tsx` */
const meta = {
  title: "Design/Components/Avatar",
  component: Avatar,
  tags: ["autodocs"],
  argTypes: {
    initials: { control: "text" },
    size: { control: "select", options: ["sm", "md", "lg"] },
  },
  args: { initials: "JD", size: "md" },
} satisfies Meta<typeof Avatar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex items-center gap-4 p-6">
      {(["sm", "md", "lg"] as const).map((size) => (
        <Avatar key={size} initials="AB" size={size} />
      ))}
    </div>
  ),
};
