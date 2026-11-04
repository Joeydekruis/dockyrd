import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { AppHeader } from "@/components/dockyard/app-header";
import { DockyrdLogo } from "@/components/dockyard/dockyrd-logo";
import { Sidebar } from "@/components/dockyard/sidebar";
import { TopBar } from "@/components/dockyard/top-bar";
import { ServerHealthCard } from "@/components/dockyard/ui/server-health-card";

const meta = {
  title: "Dockyrd/Shell pieces",
} satisfies Meta;

export default meta;

export const Logo: StoryObj = {
  render: () => <DockyrdLogo />,
};

export const TopBarAnnouncement: StoryObj = {
  render: () => <TopBar visible />,
};

export const AppHeaderBar: StoryObj = {
  render: () => (
    <AppHeader
      breadcrumb="Home"
      title="Dashboard"
      onOpenSearch={() => undefined}
    />
  ),
};

export const SidebarMain: StoryObj = {
  render: () => (
    <div
      className="h-[640px] w-[232px] border-r"
      style={{ borderColor: "var(--border)" }}
    >
      <Sidebar />
    </div>
  ),
  parameters: {
    layout: "fullscreen",
    nextjs: {
      navigation: { pathname: "/" },
    },
  },
};

export const SidebarSettings: StoryObj = {
  render: () => (
    <div
      className="h-[640px] w-[232px] border-r"
      style={{ borderColor: "var(--border)" }}
    >
      <Sidebar />
    </div>
  ),
  parameters: {
    layout: "fullscreen",
    nextjs: {
      navigation: { pathname: "/settings/profile" },
    },
  },
};

export const ServerHealth: StoryObj = {
  render: () => (
    <div className="max-w-xs">
      <ServerHealthCard version="v0.0.27" statusLine="Update available" />
    </div>
  ),
};

export const FormField: StoryObj = {
  render: () => (
    <label className="block max-w-md text-sm">
      <span style={{ color: "var(--muted-text)" }}>Public URL</span>
      <input
        className="field-input mt-1"
        defaultValue="https://panel.example.com"
        aria-label="Public URL"
      />
    </label>
  ),
};
