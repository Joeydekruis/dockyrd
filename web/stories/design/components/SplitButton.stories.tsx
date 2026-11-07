import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ArrowDownToLine, GitBranch, RotateCcw } from "lucide-react";
import {
  SplitButton,
  type SplitButtonMenuItem,
  type SplitButtonVariant,
} from "@/components/dockyrd/ui/split-button";

const deployMenuItems: SplitButtonMenuItem[] = [
  { id: "latest", label: "Deploy latest commit", icon: ArrowDownToLine },
  { id: "redeploy", label: "Redeploy last build", icon: RotateCcw },
  { id: "specific", label: "Deploy specific commit…", icon: GitBranch },
];

const variants: SplitButtonVariant[] = ["primary", "outline", "secondary"];

/** **Source:** `web/components/dockyrd/ui/split-button.tsx` */
const meta = {
  title: "Design/Components/Split button",
  component: SplitButton,
  tags: ["autodocs"],
  argTypes: {
    variant: { control: "select", options: variants },
    children: { control: "text" },
    menuLabel: { control: "text" },
  },
  args: {
    children: "Deploy now",
    variant: "primary",
    menuLabel: "More deploy options",
    menuItems: deployMenuItems,
  },
} satisfies Meta<typeof SplitButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  render: (args) => (
    <div className="p-8">
      <SplitButton {...args} />
    </div>
  ),
};

/** Design board: Split button · open */
export const Open: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="p-8">
      <SplitButton
        menuItems={deployMenuItems}
        menuLabel="More deploy options"
        defaultMenuOpen
      >
        Deploy now
      </SplitButton>
    </div>
  ),
};

export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex flex-col gap-6 p-8">
      {variants.map((variant) => (
        <SplitButton
          key={variant}
          variant={variant}
          menuItems={
            variant === "primary"
              ? deployMenuItems
              : [
                  { id: "a", label: "Option one" },
                  { id: "b", label: "Option two" },
                ]
          }
        >
          {variant === "primary" ? "Deploy now" : "Visit site"}
        </SplitButton>
      ))}
      <div
        className="inline-flex rounded-xl p-6"
        style={{ background: "var(--hero-gradient)" }}
      >
        <SplitButton
          variant="secondary"
          menuItems={[
            { id: "add", label: "Add service" },
            { id: "import", label: "Import from Git" },
          ]}
        >
          Add service
        </SplitButton>
      </div>
    </div>
  ),
};
