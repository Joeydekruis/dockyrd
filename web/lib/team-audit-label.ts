import type { AuditEntry } from "@/lib/api";
import { formatRelativeTime } from "@/lib/team-format";

const ACTION_LABELS: Record<string, string> = {
  "user.role_change": "changed a team member role",
  "user.remove": "removed a team member",
  "invite.create": "sent a team invite",
  "invite.revoke": "revoked a team invite",
  "invite.accept": "accepted a team invite",
};

export function describeTeamAuditEntry(
  row: AuditEntry,
  userNameById?: Map<string, string>,
) {
  const who =
    row.user_id && userNameById?.get(row.user_id)
      ? userNameById.get(row.user_id)
      : row.user_id
        ? "Team member"
        : "Dockyrd";
  const text = ACTION_LABELS[row.action] ?? row.action.replaceAll(".", " ");
  const ref = [row.target_kind, row.target_id].filter(Boolean).join(" · ");
  let time = "—";
  if (row.created_at > 0) {
    const rel = formatRelativeTime(row.created_at);
    time = rel === "Active now" ? "Just now" : rel.replace(/^Active /, "");
  }
  return { who, text, ref, time };
}
