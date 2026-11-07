import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ArrowUpRight, Copy } from "lucide-react";
import { Button } from "@/components/dockyard/ui/button";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import {
  Tooltip,
  TooltipInlineIcon,
  type TooltipPlacement,
} from "@/components/dockyard/ui/tooltip";

const meta = {
  title: "Design/Components/Tooltip",
  component: Tooltip,
  tags: ["autodocs"],
  argTypes: {
    placement: {
      control: "select",
      options: ["top", "bottom", "left", "right"] satisfies TooltipPlacement[],
    },
  },
  args: {
    placement: "top",
    content: "Copy URL",
    trigger: <Button variant="outline">Action</Button>,
  },
} satisfies Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  render: (args) => (
    <div className="p-24">
      <Tooltip {...args} />
    </div>
  ),
};

/** Design board: Tooltip · placements */
export const Placements: StoryObj = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex flex-wrap items-end gap-16 p-16">
      <Tooltip
        placement="top"
        defaultOpen
        trigger={
          <span
            className="inline-block size-10 rounded-lg border"
            style={{ borderColor: "var(--ink-navy)" }}
            aria-hidden
          />
        }
        content="Copy URL"
      />
      <Tooltip
        placement="bottom"
        defaultOpen
        trigger={
          <span
            className="inline-block size-10 rounded-lg border"
            style={{ borderColor: "var(--ink-navy)" }}
            aria-hidden
          />
        }
        content={
          <>
            Open site
            <TooltipInlineIcon>
              <ArrowUpRight strokeWidth={2} className="size-3.5" />
            </TooltipInlineIcon>
          </>
        }
      />
      <Tooltip
        placement="right"
        defaultOpen
        trigger={<IconButton icon={Copy} label="Copy URL" />}
        content="Copied"
      />
    </div>
  ),
};

export const OnIconButton: Story = {
  args: {
    placement: "top",
    content: "Copy URL",
    trigger: <IconButton icon={Copy} label="Copy" />,
  },
  render: (args) => (
    <div className="p-16">
      <Tooltip {...args} />
    </div>
  ),
};
