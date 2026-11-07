import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Spinner } from "@/components/dockyard/ui/spinner";

const sizes = ["sm", "md"] as const;

const meta = {
  title: "Design/Components/Spinner",
  component: Spinner,
  tags: ["autodocs"],
  argTypes: {
    size: { control: "select", options: sizes },
    label: { control: "text" },
  },
  args: { size: "md", label: "Loading" },
} satisfies Meta<typeof Spinner>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex items-center gap-8 p-6">
      {sizes.map((size) => (
        <Spinner key={size} size={size} label={`Size ${size}`} />
      ))}
    </div>
  ),
};
