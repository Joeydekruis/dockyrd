import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ChartPanel } from "@/components/dockyrd/ui/chart-panel";

const meta = {
  title: "Design/Components/Chart panel",
  component: ChartPanel,
  tags: ["autodocs"],
  argTypes: {
    title: { control: "text" },
    subtitle: { control: "text" },
    height: { control: "number" },
  },
  args: {
    title: "Request rate",
    subtitle: "Last 24 hours",
    height: 160,
  },
} satisfies Meta<typeof ChartPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const Demo: Story = {
  render: () => (
    <div className="max-w-md p-6">
      <ChartPanel title="CPU usage" subtitle="7 day trend" height={180} />
    </div>
  ),
};
