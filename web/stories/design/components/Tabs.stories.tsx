import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { Tabs } from "@/components/dockyrd/ui/tabs";

const tabs = [
  { id: "overview", label: "Overview" },
  { id: "deploys", label: "Deploys" },
  { id: "settings", label: "Settings" },
];

const meta = {
  title: "Design/Components/Tabs",
  component: Tabs,
  tags: ["autodocs"],
  args: {
    tabs,
    activeId: "overview",
    onChange: () => {},
  },
} satisfies Meta<typeof Tabs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: { tabs, activeId: "overview", onChange: () => {} },
  render: () => {
    const [activeId, setActiveId] = useState("overview");
    return (
      <div className="max-w-xl p-6">
        <Tabs tabs={tabs} activeId={activeId} onChange={setActiveId} />
        <p className="mt-4 text-sm" style={{ color: "var(--muted-text)" }}>
          Active: {activeId}
        </p>
      </div>
    );
  },
};
