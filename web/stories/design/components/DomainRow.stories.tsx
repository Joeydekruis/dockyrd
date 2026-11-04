import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { DomainRow } from "@/components/dockyard/ui/domain-row";

/** **Source:** `web/components/dockyard/ui/domain-row.tsx` */
const meta = {
  title: "Design/Components/Domain row",
  component: DomainRow,
  tags: ["autodocs"],
  argTypes: {
    hostname: { control: "text" },
    statusLabel: { control: "text" },
    statusVariant: {
      control: "select",
      options: ["ok", "warn", "err", "idle", "info"],
    },
    primary: { control: "boolean" },
  },
  args: {
    hostname: "api.example.com",
    statusLabel: "Verified",
    statusVariant: "ok",
    primary: true,
  },
} satisfies Meta<typeof DomainRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
