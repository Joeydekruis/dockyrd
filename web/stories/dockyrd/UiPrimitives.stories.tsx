import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { AnnouncementBanner } from "@/components/dockyard/ui/announcement-banner";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { SearchTrigger } from "@/components/dockyard/ui/search-trigger";

const meta = {
  title: "Dockyrd/Shell pieces",
} satisfies Meta;

export default meta;

export const Announcement: StoryObj = {
  render: () => <AnnouncementBanner />,
};

export const SearchTriggerControl: StoryObj = {
  render: () => <SearchTrigger onClick={() => undefined} />,
};

export const Empty: StoryObj = {
  render: () => (
    <EmptyState
      title="No buckets yet"
      description="Create a bucket to store assets for your static sites."
      actionLabel="Create bucket"
      actionHref="#"
    />
  ),
};
