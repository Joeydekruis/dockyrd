import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { SessionRow } from "@/components/dockyrd/ui/session-row";

/** **Source:** `web/components/dockyrd/ui/session-row.tsx` */
const meta = {
  title: "Design/Components/Session row",
  component: SessionRow,
  tags: ["autodocs"],
  argTypes: {
    device: { control: "text" },
    location: { control: "text" },
    lastActive: { control: "text" },
    current: { control: "boolean" },
  },
  args: {
    device: "MacBook Pro",
    location: "Amsterdam",
    lastActive: "Active now",
    current: true,
  },
} satisfies Meta<typeof SessionRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
export const Revokable: Story = {
  args: {
    current: false,
    lastActive: "2 days ago",
    onRevoke: () => {},
  },
};
