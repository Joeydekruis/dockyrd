import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyard/ui/status-badge";

const variants: StatusVariant[] = ["ok", "warn", "err", "idle", "info"];

const meta = {
  title: "Design/Components/Status badge",
  component: StatusBadge,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    variant: { control: "select", options: variants },
  },
  args: { label: "Running", variant: "ok" },
} satisfies Meta<typeof StatusBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex flex-wrap gap-3 p-6">
      {variants.map((variant) => (
        <StatusBadge key={variant} variant={variant} label={variant} />
      ))}
    </div>
  ),
};
