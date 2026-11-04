import type { LucideIcon } from "lucide-react";
import {
  ArrowLeft,
  BarChart3,
  Bell,
  Box,
  Clock,
  Database,
  Globe,
  HardDrive,
  KeyRound,
  Layers,
  LayoutDashboard,
  Mail,
  Plug,
  RefreshCw,
  Settings,
  Shield,
  SlidersHorizontal,
  User,
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

export const serverSettingsNav: NavItem = {
  label: "Server settings",
  href: "/settings/server/general",
  icon: Settings,
  group: "server",
};

export type SettingsNavItem = {
  label: string;
  href: string;
  icon: LucideIcon;
};

export const accountSettingsNav: SettingsNavItem[] = [
  { label: "Profile", href: "/settings/account/profile", icon: User },
  {
    label: "Password & 2FA",
    href: "/settings/account/password",
    icon: KeyRound,
  },
  {
    label: "Notifications",
    href: "/settings/account/notifications",
    icon: Bell,
  },
  {
    label: "Date & time",
    href: "/settings/account/regional",
    icon: Clock,
  },
  {
    label: "Security",
    href: "/settings/account/security",
    icon: Shield,
  },
];

export const serverSettingsSidebarNav: SettingsNavItem[] = [
  {
    label: "General",
    href: "/settings/server/general",
    icon: SlidersHorizontal,
  },
  { label: "Team", href: "/settings/server/team", icon: Users },
  { label: "Email", href: "/settings/server/email", icon: Mail },
  {
    label: "Git providers",
    href: "/settings/server/integrations",
    icon: Plug,
  },
  { label: "Updates", href: "/settings/server/updates", icon: RefreshCw },
];

export const accountSettingsBackNav = {
  label: "Back to dashboard",
  href: "/",
  icon: ArrowLeft,
};

export const serverSettingsBackNav = {
  label: "Back to dashboard",
  href: "/",
  icon: ArrowLeft,
};

/** @deprecated use accountSettingsNav */
export const settingsPersonalNav = accountSettingsNav;

/** @deprecated use serverSettingsSidebarNav */
export const settingsInstanceNav = serverSettingsSidebarNav;

/** @deprecated */
export const settingsBackNav = accountSettingsBackNav;

/** @deprecated */
export const settingsNav = serverSettingsNav;

export function isAccountSettingsPath(pathname: string): boolean {
  return pathname.startsWith("/settings/account");
}

export function isServerSettingsPath(pathname: string): boolean {
  return pathname.startsWith("/settings/server");
}

export function isSettingsPath(pathname: string): boolean {
  return isAccountSettingsPath(pathname) || isServerSettingsPath(pathname);
}

export const palettePages = [
  ...mainNav,
  ...projectNav,
  ...serverNav,
  serverSettingsNav,
  ...accountSettingsNav,
  ...serverSettingsSidebarNav,
];
