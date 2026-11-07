import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { CodeInline } from "@/components/dockyrd/ui/code-inline";

/** **Source:** `web/components/dockyrd/ui/code-inline.tsx` */
const meta = {
  title: "Design/Components/Code inline",
  component: CodeInline,
  tags: ["autodocs"],
  argTypes: { children: { control: "text" } },
  args: { children: "NODE_ENV=production" },
} satisfies Meta<typeof CodeInline>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
