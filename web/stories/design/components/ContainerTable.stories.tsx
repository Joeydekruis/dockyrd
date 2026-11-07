import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import {
  ContainerTable,
  type ContainerTableRow,
} from "@/components/dockyrd/ui/container-table";

const sampleRows: ContainerTableRow[] = [
  {
    id: "1",
    name: "web",
    image: "registry/app:1.2.0",
    statusLabel: "Running",
    statusVariant: "ok",
    cpu: "12%",
    memory: "256 MiB",
  },
  {
    id: "2",
    name: "worker",
    image: "registry/worker:1.2.0",
    statusLabel: "Restarting",
    statusVariant: "warn",
    cpu: "—",
    memory: "128 MiB",
  },
];

const meta = {
  title: "Design/Components/Container table",
  component: ContainerTable,
  tags: ["autodocs"],
  args: { rows: sampleRows },
} satisfies Meta<typeof ContainerTable>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  render: (args) => (
    <div className="max-w-3xl p-6">
      <ContainerTable {...args} />
    </div>
  ),
};
