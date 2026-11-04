import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { AddCard } from "@/components/dockyard/ui/add-card";

/** **Source:** `web/components/dockyard/ui/add-card.tsx` */
const meta = {
  title: "Design/Components/Add card",
  component: AddCard,
  tags: ["autodocs"],
  argTypes: {
    title: { control: "text" },
    description: { control: "text" },
    href: { control: "text" },
  },
  args: {
    title: "Add service",
    description: "Deploy a new container from Git or a template.",
    href: "#",
  },
} satisfies Meta<typeof AddCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
