import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Breadcrumbs } from "@/components/dockyard/ui/breadcrumbs";

const meta = {
  title: "Design/Components/Breadcrumbs",
  component: Breadcrumbs,
  tags: ["autodocs"],
  args: {
    items: [
      { label: "Projects", href: "#" },
      { label: "api-gateway", href: "#" },
      { label: "Settings" },
    ],
  },
} satisfies Meta<typeof Breadcrumbs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
