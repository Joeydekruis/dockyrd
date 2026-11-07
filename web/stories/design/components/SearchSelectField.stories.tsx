import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { SearchSelectField } from "@/components/dockyrd/ui/search-select-field";

const repoOptions = [
  { value: "acme/marketing", label: "acme/marketing" },
  { value: "acme/docs", label: "acme/docs" },
  { value: "acme/web-app", label: "acme/web-app" },
  { value: "acme/design-system", label: "acme/design-system" },
  { value: "acme/internal-tools", label: "acme/internal-tools" },
];

const meta = {
  title: "Design/Components/Search select",
  component: SearchSelectField,
  tags: ["autodocs"],
  args: {
    label: "Repository",
    placeholder: "Choose a repository",
    searchPlaceholder: "Search repositories…",
    options: repoOptions,
    defaultValue: "acme/marketing",
    hint: "Type to search your repositories.",
  },
} satisfies Meta<typeof SearchSelectField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const Open: Story = {
  parameters: { controls: { disable: true } },
  render: () => {
    const [value, setValue] = useState("acme/marketing");
    return (
      <div className="max-w-md p-8">
        <SearchSelectField
          label="Repository"
          options={repoOptions}
          value={value}
          onChange={setValue}
          searchPlaceholder="Search repositories…"
          hint="Type to search your repositories."
          defaultOpen
        />
      </div>
    );
  },
};
