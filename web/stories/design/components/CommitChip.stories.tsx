import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { CommitChip } from "@/components/dockyrd/ui/commit-chip";

/** **Source:** `web/components/dockyrd/ui/commit-chip.tsx` */
const meta = {
  title: "Design/Components/Commit chip",
  component: CommitChip,
  tags: ["autodocs"],
  argTypes: {
    sha: { control: "text" },
    message: { control: "text" },
  },
  args: { sha: "a1b2c3d", message: "fix: proxy headers" },
} satisfies Meta<typeof CommitChip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
