import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { DangerZone } from "@/components/dockyard/ui/danger-zone";
import { Modal } from "@/components/dockyard/ui/modal";
import { SearchPalette } from "@/components/dockyard/ui/search-palette";
import { Snackbar } from "@/components/dockyard/ui/snackbar";
import { Tabs } from "@/components/dockyard/ui/tabs";
import { TimeRangeToggle } from "@/components/dockyard/ui/time-range-toggle";

const meta = {
  title: "Dockyrd/Overlays",
} satisfies Meta;

export default meta;

export const SearchPopup: StoryObj = {
  render: () => {
    const [query, setQuery] = useState("app");
    const results = [
      { kind: "page", label: "Applications", href: "/applications" },
      { kind: "resource", label: "marketing", href: "/applications/marketing" },
    ].filter((row) => row.label.toLowerCase().includes(query.toLowerCase()));

    return (
      <div className="mx-auto max-w-xl pt-8">
        <SearchPalette
          query={query}
          onQueryChange={setQuery}
          results={results}
        />
      </div>
    );
  },
};

export const ModalDialog: StoryObj = {
  render: () => (
    <Modal
      open
      title="Update available"
      description="Running v0.0.27. Latest release: v0.0.28."
      footer={
        <>
          <Button>Go to Updates</Button>
          <Button variant="outline">Close</Button>
        </>
      }
    />
  ),
  parameters: { layout: "fullscreen" },
};

export const SnackbarInfo: StoryObj = {
  render: () => (
    <Snackbar
      message={
        <>
          Update <span className="font-mono">v0.0.28</span> is available.
        </>
      }
      action={<Button variant="outline">View</Button>}
      onDismiss={() => undefined}
    />
  ),
};

export const SnackbarWarn: StoryObj = {
  render: () => (
    <Snackbar
      tone="warn"
      message="Owners must enable two-factor authentication."
      onDismiss={() => undefined}
    />
  ),
};

export const TabsExample: StoryObj = {
  render: () => {
    const [tab, setTab] = useState("overview");
    return (
      <Tabs
        activeId={tab}
        onChange={setTab}
        tabs={[
          { id: "overview", label: "Overview" },
          { id: "deployments", label: "Deployments" },
          { id: "logs", label: "Logs" },
        ]}
      />
    );
  },
};

export const TimeRange: StoryObj = {
  render: () => {
    const [range, setRange] = useState<"1h" | "24h" | "7d" | "30d">("24h");
    return <TimeRangeToggle value={range} onChange={setRange} />;
  },
};

export const DangerZoneExample: StoryObj = {
  render: () => (
    <DangerZone
      title="Delete project"
      description="This permanently removes all services, volumes, and deployment history."
      action={
        <Button variant="outline" style={{ color: "var(--danger)" }}>
          Delete project
        </Button>
      }
    />
  ),
};
