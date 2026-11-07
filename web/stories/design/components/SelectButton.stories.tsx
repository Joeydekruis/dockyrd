import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { SelectButton } from "@/components/dockyrd/ui/select-button";

const options = [
  { value: "prod", label: "Production" },
  { value: "staging", label: "Staging" },
  { value: "dev", label: "Development" },
];

const meta = {
  title: "Design/Components/Select button",
  component: SelectButton,
  tags: ["autodocs"],
  args: {
    options,
    value: "prod",
    onChange: () => {},
    ariaLabel: "Environment",
  },
} satisfies Meta<typeof SelectButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: {
    options,
    value: "prod",
    onChange: () => {},
    ariaLabel: "Environment",
  },
  render: () => {
    const [value, setValue] = useState("prod");
    return (
      <div className="p-6">
        <SelectButton
          options={options}
          value={value}
          onChange={(v) => setValue(v as string)}
          ariaLabel="Environment"
        />
      </div>
    );
  },
};

export const MultiSelect: Story = {
  args: {
    options,
    value: ["prod"],
    onChange: () => {},
    multi: true,
    ariaLabel: "Environments",
  },
  render: () => {
    const [value, setValue] = useState<string[]>(["prod"]);
    return (
      <div className="p-6">
        <SelectButton
          options={options}
          value={value}
          onChange={(v) => setValue(v as string[])}
          multi
          ariaLabel="Environments"
        />
      </div>
    );
  },
};
