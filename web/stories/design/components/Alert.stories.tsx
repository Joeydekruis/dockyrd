import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Alert, type AlertTone } from "@/components/dockyard/ui/alert";
import { Button } from "@/components/dockyard/ui/button";

const tones: AlertTone[] = ["info", "success", "warning", "error"];

/** **Source:** `web/components/dockyard/ui/alert.tsx` */
const meta = {
  title: "Design/Components/Alert",
  component: Alert,
  tags: ["autodocs"],
  argTypes: {
    tone: { control: "select", options: tones },
    message: { control: "text" },
  },
  args: {
    tone: "info",
    message: "Deploy finished successfully.",
  },
} satisfies Meta<typeof Alert>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex max-w-xl flex-col gap-3 p-6">
      {tones.map((tone) => (
        <Alert key={tone} tone={tone} message={`${tone} alert message`} />
      ))}
    </div>
  ),
};

export const WithAction: Story = {
  args: {
    tone: "warning",
    message: "Certificate expires in 7 days.",
    action: <Button variant="outline">Renew</Button>,
  },
};
