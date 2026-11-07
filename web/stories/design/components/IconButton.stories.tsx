import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Settings, X } from "lucide-react";
import { IconButton } from "@/components/dockyrd/ui/icon-button";

/** **Source:** `web/components/dockyrd/ui/icon-button.tsx` (`.dockyrd-icon-btn`) */
const meta = {
  title: "Design/Components/Icon button",
  component: IconButton,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    disabled: { control: "boolean" },
  },
  args: {
    icon: Settings,
    label: "Settings",
    disabled: false,
  },
} satisfies Meta<typeof IconButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const Dismiss: Story = {
  args: { icon: X, label: "Dismiss announcement" },
};
