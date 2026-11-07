import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ProgressBar } from "@/components/dockyard/ui/progress-bar";

const variants = ["determinate", "complete", "error", "indeterminate"] as const;

const meta = {
  title: "Design/Components/Progress bar",
  component: ProgressBar,
  tags: ["autodocs"],
  argTypes: {
    variant: { control: "select", options: variants },
    value: { control: { type: "range", min: 0, max: 100 } },
    label: { control: "text" },
  },
  args: { variant: "determinate", value: 45, label: "Uploading" },
} satisfies Meta<typeof ProgressBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex max-w-md flex-col gap-6 p-6">
      <ProgressBar label="Determinate" variant="determinate" value={60} />
      <ProgressBar label="Complete" variant="complete" />
      <ProgressBar label="Error" variant="error" value={30} />
      <ProgressBar label="Indeterminate" variant="indeterminate" />
    </div>
  ),
};
