import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Checkbox } from "@/components/dockyrd/ui/checkbox";

const meta = {
  title: "Design/Components/Checkbox",
  component: Checkbox,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    disabled: { control: "boolean" },
  },
  args: {
    label: "Enable auto-deploy on push",
    name: "auto-deploy",
    disabled: false,
  },
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
