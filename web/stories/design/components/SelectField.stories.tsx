import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { SelectField } from "@/components/dockyard/ui/select-field";

const branchOptions = [
  { value: "main", label: "main" },
  { value: "develop", label: "develop" },
  { value: "staging", label: "staging" },
  { value: "release/1.0", label: "release/1.0" },
];

const meta = {
  title: "Design/Components/Select field",
  component: SelectField,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    hint: { control: "text" },
    placeholder: { control: "text" },
    disabled: { control: "boolean" },
  },
  args: {
    label: "Branch",
    name: "branch",
    placeholder: "Choose a branch",
    disabled: false,
    options: branchOptions,
    defaultValue: "main",
  },
} satisfies Meta<typeof SelectField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const WithHint: Story = {
  args: {
    label: "Region",
    defaultValue: "eu-central-1",
    hint: "Cannot be changed after first deploy.",
    options: [
      { value: "eu-central-1", label: "EU Central" },
      { value: "us-east-1", label: "US East" },
    ],
  },
};

export const Placeholder: Story = {
  args: {
    label: "Region",
    defaultValue: "",
    placeholder: "Choose a region",
    options: [
      { value: "eu-central-1", label: "EU Central" },
      { value: "us-east-1", label: "US East" },
    ],
  },
};

/** Design board: Select · open */
export const Open: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const [value, setValue] = useState("main");
    return (
      <div className="max-w-sm p-8">
        <SelectField
          label="Branch"
          options={branchOptions}
          value={value}
          onChange={setValue}
          defaultOpen
        />
      </div>
    );
  },
};

export const WithError: Story = {
  args: {
    label: "Branch",
    defaultValue: "main",
    error: "Branch is protected.",
    options: branchOptions,
  },
};
