import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Button } from "@/components/dockyard/ui/button";
import { Snackbar, SnackbarViewport } from "@/components/dockyard/ui/snackbar";

const tones = ["info", "warn", "success"] as const;

const meta = {
  title: "Design/Components/Snackbar",
  component: Snackbar,
  tags: ["autodocs"],
  argTypes: {
    tone: { control: "select", options: tones },
    message: { control: "text" },
  },
  args: {
    tone: "info",
    message: "Settings saved.",
    onDismiss: () => {},
  },
} satisfies Meta<typeof Snackbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex max-w-lg flex-col gap-3 p-6">
      {tones.map((tone) => (
        <Snackbar key={tone} tone={tone} message={`${tone} snackbar`} />
      ))}
    </div>
  ),
};

export const WithAction: Story = {
  args: {
    message: "Deploy failed — view logs",
    action: <Button variant="link">View logs</Button>,
  },
};

export const Viewport: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="relative min-h-[200px]">
      <SnackbarViewport
        open
        tone="success"
        message="Settings saved."
        onDismiss={() => undefined}
      />
    </div>
  ),
};
