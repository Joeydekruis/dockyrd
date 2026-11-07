import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { RadioGroup } from "@/components/dockyrd/ui/radio-group";

const options = [
  { value: "docker", label: "Dockerfile", description: "Build from repo root" },
  {
    value: "nixpacks",
    label: "Nixpacks",
    description: "Zero-config buildpack",
  },
];

const meta = {
  title: "Design/Components/Radio group",
  component: RadioGroup,
  tags: ["autodocs"],
  args: {
    name: "build",
    legend: "Build method",
    options,
    value: "docker",
    onChange: () => {},
  },
} satisfies Meta<typeof RadioGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: {
    name: "build",
    legend: "Build method",
    options,
    value: "docker",
    onChange: () => {},
  },
  render: () => {
    const [value, setValue] = useState("docker");
    return (
      <div className="max-w-md p-6">
        <RadioGroup
          name="build"
          legend="Build method"
          options={options}
          value={value}
          onChange={setValue}
        />
      </div>
    );
  },
};
