import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { FilterChips } from "@/components/dockyard/ui/filter-chips";

const options = [
  { value: "all", label: "All" },
  { value: "running", label: "Running" },
  { value: "failed", label: "Failed" },
];

const meta = {
  title: "Design/Components/Filter chips",
  component: FilterChips,
  tags: ["autodocs"],
  args: {
    options,
    active: "all",
    onChange: () => {},
  },
} satisfies Meta<typeof FilterChips>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: { options, active: "all", onChange: () => {} },
  render: () => {
    const [active, setActive] = useState("all");
    return (
      <div className="p-6">
        <FilterChips options={options} active={active} onChange={setActive} />
      </div>
    );
  },
};
