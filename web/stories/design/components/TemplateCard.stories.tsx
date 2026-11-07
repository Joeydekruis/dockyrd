import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { TemplateCard } from "@/components/dockyrd/ui/template-card";

/** **Source:** `web/components/dockyrd/ui/template-card.tsx` */
const meta = {
  title: "Design/Components/Template card",
  component: TemplateCard,
  tags: ["autodocs"],
  argTypes: {
    name: { control: "text" },
    description: { control: "text" },
    stack: { control: "text" },
    recommended: { control: "boolean" },
    href: { control: "text" },
  },
  args: {
    name: "Next.js",
    description: "App Router with static export ready for dockyrd.",
    stack: "Node 20",
    recommended: true,
    href: "#",
  },
} satisfies Meta<typeof TemplateCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
