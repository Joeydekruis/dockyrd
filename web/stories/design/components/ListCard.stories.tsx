import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ListCard } from "@/components/dockyard/ui/list-card";

/** **Source:** `web/components/dockyrd/ui/list-card.tsx` */
const meta = {
  title: "Design/Components/List card",
  component: ListCard,
  tags: ["autodocs"],
  argTypes: {
    title: { control: "text" },
    description: { control: "text" },
    stat: { control: "text" },
    href: { control: "text" },
  },
  args: {
    title: "Environment variables",
    description: "Secrets and config for this service",
    stat: "12 vars",
    href: "#",
  },
} satisfies Meta<typeof ListCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
