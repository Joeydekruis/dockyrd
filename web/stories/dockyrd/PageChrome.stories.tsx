import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Breadcrumbs } from "@/components/dockyard/ui/breadcrumbs";
import { Button } from "@/components/dockyard/ui/button";
import { PageHeader } from "@/components/dockyard/ui/page-header";

const meta = {
  title: "Dockyrd/Page chrome",
} satisfies Meta;

export default meta;

export const BreadcrumbsExample: StoryObj = {
  render: () => (
    <Breadcrumbs
      items={[
        { label: "Home", href: "/" },
        { label: "Projects", href: "/projects" },
        { label: "Client X" },
      ]}
    />
  ),
};

export const PageHeaderExample: StoryObj = {
  render: () => (
    <PageHeader
      eyebrow="Project"
      title="Client X"
      subtitle="Production workloads for the marketing site and API."
      action={<Button>Create service</Button>}
    />
  ),
};
