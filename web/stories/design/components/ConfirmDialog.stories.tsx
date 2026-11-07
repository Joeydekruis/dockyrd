import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { ConfirmDialog } from "@/components/dockyard/ui/confirm-dialog";

const meta = {
  title: "Design/Components/Confirm dialog",
  component: ConfirmDialog,
  tags: ["autodocs"],
  argTypes: {
    title: { control: "text" },
    description: { control: "text" },
    confirmLabel: { control: "text" },
    cancelLabel: { control: "text" },
    destructive: { control: "boolean" },
  },
  args: {
    open: true,
    title: "Delete service?",
    description: "This removes containers and cannot be undone.",
    confirmLabel: "Delete",
    cancelLabel: "Cancel",
    destructive: true,
    onConfirm: () => {},
    onCancel: () => {},
  },
} satisfies Meta<typeof ConfirmDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: {
    open: true,
    onConfirm: () => {},
    onCancel: () => {},
  },
  render: (args) => {
    const [open, setOpen] = useState(true);
    return (
      <>
        <Button variant="danger" onClick={() => setOpen(true)}>
          Open dialog
        </Button>
        <ConfirmDialog
          {...args}
          open={open}
          onConfirm={() => setOpen(false)}
          onCancel={() => setOpen(false)}
        />
      </>
    );
  },
};
