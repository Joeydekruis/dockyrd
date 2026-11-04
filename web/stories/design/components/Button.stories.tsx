import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { ArrowRight, Plus, X } from "lucide-react";
import { Button, type ButtonVariant } from "@/components/dockyard/ui/button";

const iconArgType = {
  control: "select" as const,
  options: ["none", "plus", "arrowRight", "close"],
  mapping: {
    none: undefined,
    plus: Plus,
    arrowRight: ArrowRight,
    close: X,
  },
};

const variants: ButtonVariant[] = [
  "primary",
  "outline",
  "secondary",
  "link",
  "danger",
  "ghost",
  "text",
];

/**
 * **Source:** `web/components/dockyard/ui/button.tsx` — same import path as the panel.
 *
 * Use **Playground** (Controls) to tune variant, icons, disabled, and label.
 * Use **All variants** as a visual checklist against the design board.
 */
const meta = {
  title: "Design/Components/Button",
  component: Button,
  tags: ["autodocs"],
  argTypes: {
    variant: { control: "select", options: variants },
    iconLeft: iconArgType,
    iconRight: iconArgType,
    iconOnly: { control: "boolean" },
    disabled: { control: "boolean" },
    children: { control: "text" },
  },
  args: {
    children: "Create application",
    variant: "primary",
    iconLeft: undefined,
    iconRight: undefined,
    iconOnly: false,
    disabled: false,
  },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Primary control surface — matches production props 1:1. */
export const Playground: Story = {};

/** Every `ButtonVariant` in one row (light background). */
export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="flex max-w-3xl flex-col gap-4 p-8">
      {variants
        .filter((v) => v !== "secondary" && v !== "ghost")
        .map((variant) => (
          <div key={variant} className="flex flex-wrap items-center gap-4">
            <span
              className="w-24 shrink-0 font-mono text-[11px] uppercase tracking-widest"
              style={{ color: "var(--muted-text)" }}
            >
              {variant}
            </span>
            <Button variant={variant}>
              {variant === "link" ? "View analytics" : "Label"}
            </Button>
            <Button variant={variant} iconLeft={Plus}>
              With icon
            </Button>
          </div>
        ))}
      <div
        className="mt-2 rounded-xl p-6"
        style={{ background: "var(--hero-gradient)" }}
      >
        <p className="mb-3 font-mono text-[11px] uppercase tracking-widest text-white/80">
          On gradient
        </p>
        <div className="flex flex-wrap gap-3">
          <Button variant="secondary" iconLeft={Plus}>
            Add service
          </Button>
          <Button variant="ghost">Read docs</Button>
        </div>
      </div>
    </div>
  ),
};

export const PrimaryWithIcon: Story = {
  args: {
    iconLeft: Plus,
    children: "Add service",
  },
};

export const Outline: Story = {
  args: { variant: "outline", children: "Docs" },
};

export const SecondaryOnGradient: Story = {
  args: {
    variant: "secondary",
    iconLeft: Plus,
    children: "Add service",
  },
  decorators: [
    (Story) => (
      <div
        className="inline-flex p-8"
        style={{ background: "var(--hero-gradient)" }}
      >
        <Story />
      </div>
    ),
  ],
};

export const LinkWithIcon: Story = {
  args: {
    variant: "link",
    iconRight: ArrowRight,
    children: "View analytics",
  },
};

export const DangerOutline: Story = {
  args: { variant: "danger", children: "Delete site" },
};

export const TextButton: Story = {
  args: { variant: "text", children: "Show" },
};

export const IconOnlyDismiss: Story = {
  args: {
    variant: "outline",
    iconOnly: true,
    iconLeft: X,
    children: "Dismiss",
    "aria-label": "Dismiss",
  },
};

export const Disabled: Story = {
  args: { disabled: true, children: "Saving…" },
};

export const AsAnchor: Story = {
  args: {
    href: "https://github.com/Joeydekruis/dockyrd",
    variant: "outline",
    children: "View on GitHub",
  },
};
