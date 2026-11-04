import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { AnnouncementBanner } from "@/components/dockyard/ui/announcement-banner";

const meta = {
  title: "Design/Components/Announcement banner",
  component: AnnouncementBanner,
  tags: ["autodocs"],
} satisfies Meta<typeof AnnouncementBanner>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: { onDismiss: () => {} },
};
