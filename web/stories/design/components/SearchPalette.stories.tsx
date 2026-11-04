import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import {
  SearchPalette,
  type SearchResultRow,
} from "@/components/dockyard/ui/search-palette";

const results: SearchResultRow[] = [
  { kind: "project", label: "api-gateway", href: "#" },
  { kind: "service", label: "web", href: "#" },
  { kind: "page", label: "Settings", href: "#" },
];

const meta = {
  title: "Design/Components/Search palette",
  component: SearchPalette,
  tags: ["autodocs"],
  parameters: {
    nextjs: { appDirectory: true },
  },
  args: {
    query: "",
    onQueryChange: () => {},
    results,
  },
} satisfies Meta<typeof SearchPalette>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  args: {
    query: "api",
    onQueryChange: () => {},
    results,
  },
  render: () => {
    const [query, setQuery] = useState("api");
    const filtered = results.filter((r) =>
      r.label.toLowerCase().includes(query.toLowerCase()),
    );
    return (
      <div className="p-8">
        <SearchPalette
          query={query}
          onQueryChange={setQuery}
          results={filtered}
        />
      </div>
    );
  },
};
