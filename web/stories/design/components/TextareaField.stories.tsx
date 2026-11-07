import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { TextareaField } from "@/components/dockyard/ui/textarea-field";

const meta = {
  title: "Design/Components/Textarea field",
  component: TextareaField,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    hint: { control: "text" },
    error: { control: "text" },
    maxLength: { control: "number" },
    disabled: { control: "boolean" },
  },
  args: {
    label: "Description",
    name: "description",
    placeholder: "What does this project do?",
    hint: "Optional override.",
    maxLength: 200,
    disabled: false,
  },
} satisfies Meta<typeof TextareaField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  render: (args) => {
    const [value, setValue] = useState("");
    return (
      <div className="max-w-md p-6">
        <TextareaField
          {...args}
          value={value}
          onChange={(e) => setValue(e.target.value)}
        />
      </div>
    );
  },
};
