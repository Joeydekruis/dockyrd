import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { Switch } from "@/components/dockyard/ui/switch";

const meta = {
  title: "Design/Components/Switch",
  component: Switch,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    disabled: { control: "boolean" },
  },
  args: {
    checked: false,
    onChange: () => {},
    label: "Auto-deploy",
    disabled: false,
  },
} satisfies Meta<typeof Switch>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: { checked: false, onChange: () => {} },
  render: (args) => {
    const [checked, setChecked] = useState(false);
    return (
      <div className="p-6">
        <Switch {...args} checked={checked} onChange={setChecked} />
      </div>
    );
  },
};
