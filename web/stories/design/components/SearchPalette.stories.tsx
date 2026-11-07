import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import {
  SearchPalette,
  type SearchResultRow,
} from "@/components/dockyard/ui/search-palette";

const demoResults: SearchResultRow[] = [
  { kind: "page", label: "Dashboard", href: "/" },
  { kind: "page", label: "Applications", href: "/applications" },
  { kind: "page", label: "Databases", href: "/databases" },
  {
    kind: "project",
    label: "marketing",
    href: "/projects/marketing",
    meta: "Static site · Client X",
  },
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
    results: demoResults,
    activeIndex: 1,
    onClose: () => {},
  },
} satisfies Meta<typeof SearchPalette>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithResults: Story = {
  args: {
    query: "",
    activeIndex: 1,
  },
};

export const EmptyQuery: Story = {
  args: {
    query: "zzz",
    results: [],
    activeIndex: 0,
  },
};

export const Playground: Story = {
  render: () => {
    const [query, setQuery] = useState("");
    const [activeIndex, setActiveIndex] = useState(0);
    const filtered = demoResults.filter((r) =>
      r.label.toLowerCase().includes(query.toLowerCase()),
    );
    return (
      <div className="p-8">
        <SearchPalette
          query={query}
          onQueryChange={(v) => {
            setQuery(v);
            setActiveIndex(0);
          }}
          results={filtered}
          activeIndex={Math.min(activeIndex, Math.max(filtered.length - 1, 0))}
          onClose={() => undefined}
        />
      </div>
    );
  },
};
