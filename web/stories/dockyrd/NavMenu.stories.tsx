import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Layers, LayoutDashboard } from "lucide-react";
import { useState } from "react";
import { NavMenuItem } from "@/components/dockyard/ui/nav-menu-item";

const meta = {
  title: "Dockyrd/Nav menu",
} satisfies Meta;

export default meta;

export const SidebarItems: StoryObj = {
  render: () => {
    const [active, setActive] = useState("dashboard");
    return (
      <nav className="flex w-56 flex-col gap-1" aria-label="Main">
        <NavMenuItem
          label="Dashboard"
          icon={LayoutDashboard}
          active={active === "dashboard"}
          onClick={() => setActive("dashboard")}
        />
        <NavMenuItem
          label="Projects"
          icon={Layers}
          active={active === "projects"}
          onClick={() => setActive("projects")}
        />
      </nav>
    );
  },
};
