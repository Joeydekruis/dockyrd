import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Button } from "@/components/dockyrd/ui/button";
import { DangerZone } from "@/components/dockyrd/ui/danger-zone";

const meta = {
  title: "Design/Components/Danger zone",
  component: DangerZone,
  tags: ["autodocs"],
  argTypes: {
    title: { control: "text" },
    description: { control: "text" },
  },
  args: {
    title: "Delete project",
    description: "All services and volumes will be permanently removed.",
    action: <Button variant="danger">Delete project</Button>,
  },
} satisfies Meta<typeof DangerZone>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
