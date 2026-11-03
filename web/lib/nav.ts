export type NavItem = {
  label: string;
  href: string;
  group?: "main" | "project" | "server";
};

export const mainNav: NavItem[] = [
  { label: "Dashboard", href: "/", group: "main" },
  { label: "Projects", href: "/projects", group: "main" },
];

export const projectNav: NavItem[] = [
  { label: "Applications", href: "/applications", group: "project" },
  { label: "Databases", href: "/databases", group: "project" },
  { label: "Static sites", href: "/static-sites", group: "project" },
  { label: "Object storage", href: "/object-storage", group: "project" },
];

export const serverNav: NavItem[] = [
  { label: "Analytics", href: "/analytics", group: "server" },
];

export const settingsNav: NavItem = {
  label: "Settings",
  href: "/settings/profile",
  group: "main",
};

export const palettePages = [
  ...mainNav,
  ...projectNav,
  ...serverNav,
  settingsNav,
];
