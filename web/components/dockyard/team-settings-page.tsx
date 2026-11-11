"use client";

import { Mail, MoreHorizontal, Plus, Shield, UserPlus, X } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { InviteTeamMemberModal } from "@/components/dockyard/invite-team-member-modal";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { TeamChangeRoleModal } from "@/components/dockyard/team-change-role-modal";
import { TeamRemoveMemberModal } from "@/components/dockyard/team-remove-member-modal";
import { TeamRoleBadge } from "@/components/dockyard/team-role-badge";
import { Button } from "@/components/dockyard/ui/button";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import { PopoverMenu } from "@/components/dockyard/ui/popover-menu";
import {
  type AuditEntry,
  createInvite,
  getMe,
  listAuditLog,
  listInvites,
  listTeamUsers,
  revokeInvite,
  type TeamInvite,
  type TeamUser,
} from "@/lib/api";
import { describeTeamAuditEntry } from "@/lib/team-audit-label";
import {
  formatInviteExpiry,
  formatRelativeTime,
  initialsFromName,
} from "@/lib/team-format";
import { INVITE_ROLE_OPTIONS, TEAM_ROLE_META, memberScopeText } from "@/lib/team-roles";

export function TeamSettingsPage() {
  const [users, setUsers] = useState<TeamUser[]>([]);
  const [invites, setInvites] = useState<TeamInvite[]>([]);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [meId, setMeId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [roleTarget, setRoleTarget] = useState<TeamUser | null>(null);
  const [removeTarget, setRemoveTarget] = useState<TeamUser | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [me, { users: us }, { invites: inv }] = await Promise.all([
        getMe(),
        listTeamUsers(),
        listInvites(),
      ]);
      setMeId(me.id);
      setUsers(us);
      setInvites(inv);
      setError(null);
      try {
        const { entries } = await listAuditLog();
        const teamish = entries.filter(
          (e) =>
            e.action.startsWith("user.") || e.action.startsWith("invite."),
        );
        setAudit(teamish.slice(0, 6));
      } catch {
        setAudit([]);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load team");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  function flash(msg: string) {
    setToast(msg);
    window.setTimeout(() => setToast(null), 2500);
  }

  function inviteUrl(token: string | undefined) {
    if (!token) return null;
    return `${window.location.origin}/accept-invite?token=${token}`;
  }

  async function copyInviteLink(inv: TeamInvite) {
    try {
      if (inv.token) {
        const url = inviteUrl(inv.token);
        if (url) await navigator.clipboard.writeText(url);
        flash("Invite link copied");
        return;
      }
      const created = await createInvite({ email: inv.email, role: inv.role });
      if (created.token) {
        const url = inviteUrl(created.token);
        if (url) await navigator.clipboard.writeText(url);
      }
      flash("Invite link copied");
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not copy link");
    }
  }

  async function resendInvite(inv: TeamInvite) {
    try {
      await createInvite({ email: inv.email, role: inv.role });
      flash(`Invite sent again to ${inv.email || "recipient"}`);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Resend failed");
    }
  }

  async function revoke(inv: TeamInvite) {
    try {
      await revokeInvite(inv.id);
      flash(`Invite${inv.email ? ` for ${inv.email}` : ""} revoked`);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not revoke invite");
    }
  }

  return (
    <DashboardShell title="Team" breadcrumb="Server settings">
      <SettingsPageBody>
        <div className="team-settings-layout">
          {error ? (
            <p className="mb-4 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-700">
              {error}
            </p>
          ) : null}

          <div className="mb-7 flex flex-wrap items-start justify-between gap-4">
            <div className="max-w-[620px] space-y-1.5">
              <h1 className="text-[30px] font-semibold leading-tight tracking-tight">
                Team
              </h1>
              <p className="text-[15px]" style={{ color: "var(--muted-text)" }}>
                People who can sign in to this Dockyrd instance. Give each person a
                role, and limit freelancers or clients to the projects they need.
              </p>
            </div>
            <Button
              type="button"
              iconLeft={Plus}
              onClick={() => setInviteOpen(true)}
            >
              Invite member
            </Button>
          </div>

          <div className="flex flex-wrap items-start gap-6">
            <div className="min-w-0 flex-1 space-y-7">
              <section aria-labelledby="team-members-heading" className="space-y-3">
                <h2
                  id="team-members-heading"
                  className="text-lg font-semibold tracking-tight"
                >
                  Members{" "}
                  <span className="font-normal" style={{ color: "var(--muted-text)" }}>
                    · {users.length}
                  </span>
                </h2>
                {loading ? (
                  <DockyrdPageLoading className="py-2" cubeSize={28} />
                ) : (
                  <div
                    className="overflow-x-auto rounded-[20px] border"
                    style={{ borderColor: "var(--border)", background: "var(--surface)" }}
                  >
                    <div className="min-w-[640px]">
                      <div
                        className="grid items-center gap-3 border-b px-5 py-3 font-mono text-[11px] uppercase tracking-wider"
                        style={{
                          gridTemplateColumns:
                            "minmax(0,2.2fr) minmax(0,1.5fr) minmax(0,1.1fr) 44px",
                          color: "var(--muted-text)",
                          borderColor: "var(--border)",
                        }}
                      >
                        <span>Member</span>
                        <span>Role &amp; access</span>
                        <span>Activity</span>
                        <span />
                      </div>
                      {users.map((u, index) => {
                        const canManage = u.role !== "owner" && u.id !== meId;
                        return (
                          <div
                            key={u.id}
                            className="grid items-center gap-3 px-5 py-3.5"
                            style={{
                              gridTemplateColumns:
                                "minmax(0,2.2fr) minmax(0,1.5fr) minmax(0,1.1fr) 44px",
                              borderTop:
                                index === 0 ? undefined : "1px solid var(--border)",
                            }}
                          >
                            <div className="flex min-w-0 items-center gap-3">
                              <span
                                className="inline-flex size-10 shrink-0 items-center justify-center rounded-full text-[13px] font-semibold"
                                style={{ background: "#E3F0FF", color: "#0B57C9" }}
                              >
                                {initialsFromName(u.name)}
                              </span>
                              <div className="min-w-0">
                                <p className="flex items-center gap-2 truncate font-semibold">
                                  {u.name}
                                  {u.id === meId ? (
                                    <span className="rounded-md bg-[#F1F0ED] px-1.5 py-0.5 font-mono text-[10px] font-medium uppercase tracking-wide text-[#3A4A63]">
                                      You
                                    </span>
                                  ) : null}
                                </p>
                                <p
                                  className="truncate text-[13px]"
                                  style={{ color: "var(--muted-text)" }}
                                >
                                  {u.email}
                                </p>
                              </div>
                            </div>
                            <div className="min-w-0 space-y-1">
                              <TeamRoleBadge role={u.role} />
                              <p
                                className="truncate text-[13px]"
                                style={{ color: "var(--muted-text)" }}
                              >
                                {memberScopeText(u.role)}
                              </p>
                            </div>
                            <div className="min-w-0 text-[13px]">
                              <p>
                                {formatRelativeTime(
                                  u.updated_at ?? u.created_at,
                                )}
                              </p>
                              <p
                                className="inline-flex items-center gap-1.5 text-xs font-medium"
                                style={{
                                  color: u.totp_enabled ? "#176B3D" : "#8A5200",
                                }}
                              >
                                <Shield className="size-3.5" strokeWidth={2} aria-hidden />
                                {u.totp_enabled ? "2FA on" : "2FA off"}
                              </p>
                            </div>
                            <div className="flex justify-end">
                              {canManage ? (
                                <PopoverMenu
                                  label={`Actions for ${u.name}`}
                                  items={[
                                    {
                                      id: "role",
                                      label: "Change role & access",
                                      onSelect: () => setRoleTarget(u),
                                    },
                                    {
                                      id: "remove",
                                      label: "Remove from team",
                                      destructive: true,
                                      dividerBefore: true,
                                      onSelect: () => setRemoveTarget(u),
                                    },
                                  ]}
                                  trigger={
                                    <IconButton
                                      label={`Actions for ${u.name}`}
                                      icon={MoreHorizontal}
                                    />
                                  }
                                />
                              ) : null}
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                )}
              </section>

              <section aria-labelledby="team-invites-heading" className="space-y-3">
                <h2
                  id="team-invites-heading"
                  className="text-lg font-semibold tracking-tight"
                >
                  Pending invites{" "}
                  <span className="font-normal" style={{ color: "var(--muted-text)" }}>
                    · {invites.length}
                  </span>
                </h2>
                <div
                  className="overflow-hidden rounded-[20px] border"
                  style={{ borderColor: "var(--border)", background: "var(--surface)" }}
                >
                  {!loading && invites.length === 0 ? (
                    <div className="flex flex-col items-start gap-2.5 px-5 py-6">
                      <p className="font-semibold">No pending invites</p>
                      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
                        Invite a colleague, freelancer or client to give them access.
                      </p>
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() => setInviteOpen(true)}
                      >
                        Invite member
                      </Button>
                    </div>
                  ) : (
                    invites.map((inv, index) => {
                      const expiry = formatInviteExpiry(inv.expires_at);
                      const emailLabel = inv.email || "Invite link";
                      return (
                        <div
                          key={inv.id}
                          className="flex flex-wrap items-center gap-3 px-5 py-3.5"
                          style={{
                            borderTop:
                              index === 0 ? undefined : "1px solid var(--border)",
                          }}
                        >
                          <span
                            className="inline-flex size-10 shrink-0 items-center justify-center rounded-full"
                            style={{ background: "#F1F0ED", color: "var(--muted-text)" }}
                          >
                            <Mail className="size-[18px]" strokeWidth={1.8} aria-hidden />
                          </span>
                          <div className="min-w-0 flex-1">
                            <p className="flex flex-wrap items-center gap-2 font-semibold">
                              {emailLabel}
                              <TeamRoleBadge role={inv.role} />
                            </p>
                            <p className="text-[13px]" style={{ color: "var(--muted-text)" }}>
                              {memberScopeText(inv.role)} · {expiry.detail}
                            </p>
                          </div>
                          <span
                            className="inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 font-mono text-[11px] font-medium uppercase tracking-wide"
                            style={{
                              background: expiry.expired ? "#FBE8E6" : "#FBF0DC",
                              color: expiry.expired ? "#B3261E" : "#8A5200",
                            }}
                          >
                            <span
                              className="size-1.5 rounded-full"
                              style={{
                                background: expiry.expired ? "#B3261E" : "#D98A00",
                              }}
                            />
                            {expiry.status}
                          </span>
                          <div className="flex items-center gap-1.5">
                            <Button
                              type="button"
                              variant="outline"
                              className="min-h-11 px-3.5"
                              onClick={() => void copyInviteLink(inv)}
                            >
                              Copy link
                            </Button>
                            {inv.email ? (
                              <Button
                                type="button"
                                variant="outline"
                                className="min-h-11 px-3.5"
                                onClick={() => void resendInvite(inv)}
                              >
                                Resend
                              </Button>
                            ) : null}
                            <IconButton
                              label={`Revoke invite for ${emailLabel}`}
                              icon={X}
                              className="text-red-700"
                              onClick={() => void revoke(inv)}
                            />
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>
              </section>

              {audit.length > 0 ? (
                <section aria-labelledby="team-activity-heading" className="space-y-3">
                  <div className="flex items-center justify-between gap-3">
                    <h2
                      id="team-activity-heading"
                      className="text-lg font-semibold tracking-tight"
                    >
                      Team activity
                    </h2>
                    <span className="text-sm" style={{ color: "var(--muted-text)" }}>
                      Recent team events
                    </span>
                  </div>
                  <div
                    className="overflow-hidden rounded-[20px] border"
                    style={{ borderColor: "var(--border)", background: "var(--surface)" }}
                  >
                    {audit.map((row, index) => {
                      const names = new Map(users.map((u) => [u.id, u.name]));
                      const line = describeTeamAuditEntry(row, names);
                      return (
                        <div
                          key={row.id}
                          className="flex items-center gap-3.5 px-5 py-3.5"
                          style={{
                            borderTop:
                              index === 0 ? undefined : "1px solid var(--border)",
                          }}
                        >
                          <span
                            className="inline-flex size-9 shrink-0 items-center justify-center rounded-[10px]"
                            style={{ background: "#F1F0ED", color: "#3A4A63" }}
                          >
                            <UserPlus
                              className="size-[18px]"
                              strokeWidth={1.8}
                              aria-hidden
                            />
                          </span>
                          <div className="min-w-0 flex-1">
                            <p className="text-sm">
                              <span className="font-semibold">{line.who}</span>{" "}
                              {line.text}
                            </p>
                            {line.ref ? (
                              <p
                                className="font-mono text-[13px]"
                                style={{ color: "var(--muted-text)" }}
                              >
                                {line.ref}
                              </p>
                            ) : null}
                          </div>
                          <span
                            className="shrink-0 text-[13px]"
                            style={{ color: "var(--muted-text)" }}
                          >
                            {line.time}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                </section>
              ) : null}
            </div>

            <aside
              aria-label="About roles"
              className="w-full min-w-[280px] max-w-[360px] shrink-0 space-y-4"
            >
              <section
                className="space-y-4 rounded-[20px] border p-5"
                style={{ borderColor: "var(--border)", background: "var(--surface)" }}
              >
                <h2 className="text-lg font-semibold tracking-tight">Roles</h2>
                {(["owner", ...INVITE_ROLE_OPTIONS] as const).map((key) => {
                  const meta = TEAM_ROLE_META[key];
                  return (
                    <div key={key} className="space-y-1.5">
                      <TeamRoleBadge role={key} />
                      <p className="text-sm" style={{ color: "#3A4A63" }}>
                        {meta.description}
                      </p>
                    </div>
                  );
                })}
              </section>
              <section
                className="space-y-2 rounded-[20px] p-5"
                style={{ background: "#E3F0FF" }}
              >
                <p className="flex items-center gap-2 font-semibold">
                  <Shield className="size-[18px]" strokeWidth={1.8} aria-hidden />
                  Working with clients?
                </p>
                <p className="text-sm" style={{ color: "#3A4A63" }}>
                  Give clients the Viewer role on just their project. They can see
                  logs and metrics but cannot deploy or change anything.
                </p>
              </section>
            </aside>
          </div>
        </div>
      </SettingsPageBody>

      {toast ? (
        <p
          role="status"
          className="fixed bottom-8 left-1/2 z-50 -translate-x-1/2 rounded-xl px-4 py-3 text-sm font-medium text-white shadow-lg"
          style={{ background: "#0B1F3A" }}
        >
          {toast}
        </p>
      ) : null}

      <InviteTeamMemberModal
        open={inviteOpen}
        onClose={() => setInviteOpen(false)}
        onInvited={refresh}
      />
      <TeamChangeRoleModal
        open={!!roleTarget}
        user={roleTarget}
        onClose={() => setRoleTarget(null)}
        onSaved={refresh}
      />
      <TeamRemoveMemberModal
        open={!!removeTarget}
        user={removeTarget}
        onClose={() => setRemoveTarget(null)}
        onRemoved={refresh}
      />
    </DashboardShell>
  );
}
