import { teamRoleMeta } from "@/lib/team-roles";

export function TeamRoleBadge({ role }: { role: string }) {
  const meta = teamRoleMeta(role);
  return (
    <span
      className="inline-flex rounded-md px-2 py-0.5 font-mono text-[11px] font-medium uppercase tracking-wide"
      style={{ background: meta.bg, color: meta.fg }}
    >
      {meta.label}
    </span>
  );
}
