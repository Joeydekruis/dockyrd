import type { LucideIcon } from "lucide-react";
import {
  ArrowLeft,
  BarChart3,
  Box,
  Database,
  Globe,
  HardDrive,
  Layers,
  LayoutDashboard,
  Plug,
  RefreshCw,
  Settings,
  Shield,
  SlidersHorizontal,
  Users,
} from "lucide-react";

export type NavItem = {
  label: string;
  href: string;
  icon: LucideIcon;
  group?: "main" | "project" | "server";
};

export const mainNav: NavItem[] = [
  { label: "Dashboard", href: "/", icon: LayoutDashboard, group: "main" },
  { label: "Projects", href: "/projects", icon: Layers, group: "main" },
];

export const projectNav: NavItem[] = [
  { label: "Applications", href: "/applications", icon: Box, group: "project" },
  { label: "Databases", href: "/databases", icon: Database, group: "project" },
  {
    label: "Static sites",
    href: "/static-sites",
    icon: Globe,
    group: "project",
  },
  {
    label: "Object storage",
    href: "/object-storage",
    icon: HardDrive,
    group: "project",
  },
];

export const serverNav: NavItem[] = [
  { label: "Analytics", href: "/analytics", icon: BarChart3, group: "server" },
];

export const settingsNav: NavItem = {
  label: "Settings",
  href: "/settings/profile",
  icon: Settings,
  group: "main",
};

export type SettingsNavItem = {
  label: string;
  href: string;
  icon: LucideIcon;
};

export const settingsPersonalNav: SettingsNavItem[] = [
  {
    label: "Profile & security",
    href: "/settings/profile",
    icon: Shield,
  },
];

export const settingsInstanceNav: SettingsNavItem[] = [
  {
    label: "General",
    href: "/settings/instance",
    icon: SlidersHorizontal,
  },
  { label: "Team", href: "/settings/team", icon: Users },
  {
    label: "Git providers",
    href: "/settings/integrations",
    icon: Plug,
  },
  { label: "Updates", href: "/settings/updates", icon: RefreshCw },
];

export const settingsBackNav = {
  label: "Back to dashboard",
  href: "/",
  icon: ArrowLeft,
};

export const palettePages = [
  ...mainNav,
  ...projectNav,
  ...serverNav,
  settingsNav,
  ...settingsPersonalNav,
  ...settingsInstanceNav,
];
