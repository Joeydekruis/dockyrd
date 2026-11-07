import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { KeyValueRow } from "@/components/dockyrd/ui/key-value-row";

/** **Source:** `web/components/dockyrd/ui/key-value-row.tsx` */
const meta = {
  title: "Design/Components/Key value row",
  component: KeyValueRow,
  tags: ["autodocs"],
  argTypes: {
    label: { control: "text" },
    value: { control: "text" },
  },
  args: { label: "Region", value: "eu-west" },
} satisfies Meta<typeof KeyValueRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
