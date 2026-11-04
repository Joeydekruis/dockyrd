import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { Select } from "@/components/dockyard/ui/select";

const meta = {
  title: "Design/Components/Select",
  component: Select,
  tags: ["autodocs"],
  args: {
    placeholder: "Choose…",
    options: [
      { value: "main", label: "main" },
      { value: "develop", label: "develop" },
    ],
    defaultValue: "main",
  },
} satisfies Meta<typeof Select>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const Open: Story = {
  render: (args) => {
    const [value, setValue] = useState("main");
    return (
      <div className="max-w-xs p-8">
        <Select
          {...args}
          value={value}
          onChange={setValue}
          defaultOpen
          aria-label="Branch"
        />
      </div>
    );
  },
};
