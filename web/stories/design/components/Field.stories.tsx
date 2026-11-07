import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Field } from "@/components/dockyrd/ui/field";

const meta = {
  title: "Design/Components/Field",
  component: Field,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    hint: { control: "text" },
    error: { control: "text" },
    placeholder: { control: "text" },
    disabled: { control: "boolean" },
  },
  args: {
    label: "Service name",
    name: "service-name",
    placeholder: "my-api",
    hint: "Lowercase, hyphens allowed.",
    disabled: false,
  },
} satisfies Meta<typeof Field>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const WithError: Story = {
  args: { error: "Name is already in use", hint: undefined },
};
