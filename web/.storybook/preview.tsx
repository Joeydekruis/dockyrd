import type { Preview } from "@storybook/nextjs-vite";
import "@/styles/globals.css";
import "./storybook-fonts.css";

const preview: Preview = {
  parameters: {
    layout: "padded",
    backgrounds: {
      default: "page",
      values: [{ name: "page", value: "#faf9f8" }],
    },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    a11y: {
      test: "todo",
    },
    nextjs: {
      appDirectory: true,
    },
  },
  decorators: [
    (Story) => (
      <div
        className="font-sans antialiased"
        style={{ color: "var(--ink-navy)" }}
      >
        <Story />
      </div>
    ),
  ],
};

export default preview;
