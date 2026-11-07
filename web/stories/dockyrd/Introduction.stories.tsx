import type { Meta, StoryObj } from "@storybook/nextjs-vite";

function IntroductionDocsPlaceholder() {
  return null;
}

const meta = {
  title: "Dockyrd/Introduction",
  component: IntroductionDocsPlaceholder,
  parameters: {
    docs: {
      description: {
        component:
          "Stories mirror design-index tokens and briefing primitives. Each story imports web/components/dockyrd/ui/* — the same modules the panel uses.",
      },
    },
  },
} satisfies Meta<typeof IntroductionDocsPlaceholder>;

export default meta;

type Story = StoryObj<typeof meta>;

/** Docs-only anchor for Introduction.mdx (hidden from story list). */
export const Catalog: Story = {
  parameters: { docsOnly: true },
  render: () => <span className="sr-only">Introduction</span>,
};
