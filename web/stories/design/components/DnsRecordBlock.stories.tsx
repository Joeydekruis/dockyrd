import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { DnsRecordBlock } from "@/components/dockyard/ui/dns-record-block";

/** **Source:** `web/components/dockyrd/ui/dns-record-block.tsx` */
const meta = {
  title: "Design/Components/DNS record block",
  component: DnsRecordBlock,
  tags: ["autodocs"],
  argTypes: {
    type: { control: "text" },
    name: { control: "text" },
    value: { control: "text" },
    ttl: { control: "text" },
  },
  args: {
    type: "CNAME",
    name: "app",
    value: "proxy.dockyrd.dev",
    ttl: "3600",
  },
} satisfies Meta<typeof DnsRecordBlock>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
