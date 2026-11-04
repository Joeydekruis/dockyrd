import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { EventFeedItem } from "@/components/dockyard/ui/event-feed-item";

/** **Source:** `web/components/dockyard/ui/event-feed-item.tsx` */
const meta = {
  title: "Design/Components/Event feed item",
  component: EventFeedItem,
  tags: ["autodocs"],
  argTypes: {
    time: { control: "text" },
    title: { control: "text" },
    detail: { control: "text" },
    actor: { control: "text" },
  },
  args: {
    time: "10:42",
    title: "Deploy completed",
    detail: "web · v1.2.0",
    actor: "joey",
  },
} satisfies Meta<typeof EventFeedItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
