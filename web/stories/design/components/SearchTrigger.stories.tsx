import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { SearchTrigger } from "@/components/dockyrd/ui/search-trigger";

const meta = {
  title: "Design/Components/Search trigger",
  component: SearchTrigger,
  tags: ["autodocs"],
  args: { onClick: () => {} },
} satisfies Meta<typeof SearchTrigger>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
