export type AssignableTeamRole = "admin" | "member" | "viewer";

export type TeamRoleKey = AssignableTeamRole | "owner";

export const TEAM_ROLE_META: Record<
  TeamRoleKey,
  { label: string; description: string; bg: string; fg: string }
> = {
  owner: {
    label: "Owner",
    description: "Full control of this instance. There is one owner.",
    bg: "#0B1F3A",
    fg: "#FFFFFF",
  },
  admin: {
    label: "Admin",
    description: "Manages team, server settings, security and every project.",
    bg: "#E3F0FF",
    fg: "#0B3D91",
  },
  member: {
    label: "Developer",
    description:
      "Creates and deploys apps, databases and static sites. No server or team settings.",
    bg: "#F1F0ED",
    fg: "#3A4A63",
  },
  viewer: {
    label: "Viewer",
    description:
      "Read-only access to projects, logs and metrics. Cannot deploy or change anything.",
    bg: "#ECEEF1",
    fg: "#4A5361",
  },
};

export const INVITE_ROLE_OPTIONS: AssignableTeamRole[] = [
  "admin",
  "member",
  "viewer",
];

export function teamRoleLabel(role: string): string {
  const key = role as TeamRoleKey;
  return TEAM_ROLE_META[key]?.label ?? role;
}

export function teamRoleMeta(role: string) {
  const key = role as TeamRoleKey;
  return TEAM_ROLE_META[key] ?? TEAM_ROLE_META.member;
}

export function memberScopeText(_role: string): string {
  return "All projects";
}
