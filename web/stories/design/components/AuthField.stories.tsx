import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { AuthField } from "@/components/dockyard/ui/auth-field";

const meta = {
  title: "Design/Components/Auth field",
  component: AuthField,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    type: { control: "select", options: ["text", "email", "password"] },
    placeholder: { control: "text" },
    disabled: { control: "boolean" },
  },
  args: {
    label: "Email",
    name: "email",
    type: "email",
    placeholder: "you@example.com",
    disabled: false,
  },
} satisfies Meta<typeof AuthField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
