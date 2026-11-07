import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Button } from "@/components/dockyard/ui/button";
import { PageHeader } from "@/components/dockyard/ui/page-header";

const meta = {
  title: "Design/Components/Page header",
  component: PageHeader,
  tags: ["autodocs"],
  argTypes: {
    eyebrow: { control: "text" },
    title: { control: "text" },
    subtitle: { control: "text" },
  },
  args: {
    eyebrow: "Project",
    title: "api-gateway",
    subtitle: "Production services and domains.",
    action: <Button iconLeft={undefined}>Deploy</Button>,
  },
} satisfies Meta<typeof PageHeader>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
