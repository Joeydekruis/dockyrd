import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Modal } from "@/components/dockyard/ui/modal";

const meta = {
  title: "Design/Components/Modal",
  component: Modal,
  tags: ["autodocs"],
  argTypes: {
    title: { control: "text" },
    description: { control: "text" },
  },
  args: {
    open: true,
    title: "Edit domain",
    description: "Update the hostname served by this service.",
  },
} satisfies Meta<typeof Modal>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: { open: true },
  render: (args) => {
    const [open, setOpen] = useState(true);
    return (
      <>
        <Button onClick={() => setOpen(true)}>Open modal</Button>
        <Modal
          {...args}
          open={open}
          onClose={() => setOpen(false)}
          footer={
            <>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button onClick={() => setOpen(false)}>Save</Button>
            </>
          }
        >
          <p className="text-sm">Modal body content goes here.</p>
        </Modal>
      </>
    );
  },
};
