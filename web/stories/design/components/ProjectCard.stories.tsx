import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ProjectCard } from "@/components/dockyard/ui/project-card";

/** **Source:** `web/components/dockyrd/ui/project-card.tsx` */
const meta = {
  title: "Design/Components/Project card",
  component: ProjectCard,
  tags: ["autodocs"],
  argTypes: {
    name: { control: "text" },
    repo: { control: "text" },
    servicesCount: { control: "number" },
    statusLabel: { control: "text" },
    statusVariant: {
      control: "select",
      options: ["ok", "warn", "err", "idle", "info"],
    },
  },
  args: {
    name: "commerce-api",
    repo: "github.com/acme/commerce",
    servicesCount: 4,
    statusLabel: "Healthy",
    statusVariant: "ok",
  },
} satisfies Meta<typeof ProjectCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const OverviewGrid: Story = {
  args: {
    name: "Client X",
    statusLabel: "Healthy",
    statusVariant: "ok",
    environments: ["Production", "Staging"],
    summaryLine: "3 apps · 1 database · 1 static site",
    activityLine: "Deployed 12 min ago",
    repo: undefined,
    servicesCount: undefined,
  },
};
