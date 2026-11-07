import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import {
  type TimeRange,
  TimeRangeToggle,
} from "@/components/dockyrd/ui/time-range-toggle";

const meta = {
  title: "Design/Components/Time range toggle",
  component: TimeRangeToggle,
  tags: ["autodocs"],
  args: {
    value: "24h" as const,
    onChange: () => {},
  },
} satisfies Meta<typeof TimeRangeToggle>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: { value: "24h", onChange: () => {} },
  render: () => {
    const [value, setValue] = useState<TimeRange>("24h");
    return (
      <div className="p-6">
        <TimeRangeToggle value={value} onChange={setValue} />
      </div>
    );
  },
};
