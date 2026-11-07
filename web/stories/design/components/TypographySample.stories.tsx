import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import {
  type TypographyRole,
  TypographySample,
} from "@/components/dockyrd/ui/typography-sample";

const roles: TypographyRole[] = ["h1", "h2", "body", "muted", "mono"];

const meta = {
  title: "Design/Components/Typography sample",
  component: TypographySample,
  tags: ["autodocs"],
  argTypes: {
    variant: { control: "select", options: roles },
    children: { control: "text" },
  },
  args: {
    variant: "body",
    children: "The quick brown fox jumps over the lazy dog.",
  },
} satisfies Meta<typeof TypographySample>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex flex-col gap-4 p-6">
      {roles.map((variant) => (
        <TypographySample key={variant} variant={variant}>
          {variant} — Sample text
        </TypographySample>
      ))}
    </div>
  ),
};
