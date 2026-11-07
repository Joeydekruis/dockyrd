import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Skeleton } from "@/components/dockyrd/ui/skeleton";

/** **Source:** `web/components/dockyrd/ui/skeleton.tsx` */
const meta = {
  title: "Design/Components/Skeleton",
  component: Skeleton,
  tags: ["autodocs"],
  argTypes: { lines: { control: "number" } },
  args: { lines: 3 },
} satisfies Meta<typeof Skeleton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
