import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Chip } from "@/components/dockyard/ui/chip";
import { CommitChip } from "@/components/dockyard/ui/commit-chip";

const meta = {
  title: "Dockyrd/Chips",
} satisfies Meta;

export default meta;

export const Commit: StoryObj = {
  render: () => (
    <CommitChip sha="a1b2c3d4e5f6" message="Rotate database credentials" />
  ),
};

export const Label: StoryObj = {
  render: () => <Chip>Open source</Chip>,
};
