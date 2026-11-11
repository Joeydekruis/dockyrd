import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { RevealableCopyField } from "@/components/dockyard/ui/revealable-copy-field";

/** **Source:** `web/components/dockyrd/ui/revealable-copy-field.tsx` */
const meta = {
  title: "Design/Components/Revealable copy field",
  component: RevealableCopyField,
  tags: ["autodocs"],
} satisfies Meta<typeof RevealableCopyField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Plain: Story = {
  args: {
    displayValue: "postgres.internal:5432",
    copyValue: "postgres.internal:5432",
    copyLabel: "Copy host",
    copyMessage: "Copied host",
  },
};

export const Secret: Story = {
  args: {
    displayValue: "postgres://user:secret@db:5432/app",
    copyValue: "postgres://user:secret@db:5432/app",
    copyLabel: "Copy DATABASE_URL",
    revealable: true,
  },
  render: function SecretDemo() {
    const [revealed, setRevealed] = useState(false);
    const value = "postgres://user:secret@db:5432/app";
    return (
      <RevealableCopyField
        displayValue={value}
        copyValue={value}
        copyLabel="Copy DATABASE_URL"
        copyMessage="Copied DATABASE_URL"
        canCopy={revealed}
        revealable
        revealed={revealed}
        onToggleReveal={() => setRevealed((v) => !v)}
      />
    );
  },
};
