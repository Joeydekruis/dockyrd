"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { SelectField } from "@/components/dockyard/ui/select-field";
import {
  type AuditEntry,
  createInvite,
  listAuditLog,
  listInvites,
  listTeamUsers,
  type TeamInvite,
  type TeamUser,
} from "@/lib/api";

export function TeamSettingsPage() {
  const [users, setUsers] = useState<TeamUser[]>([]);
  const [invites, setInvites] = useState<TeamInvite[]>([]);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("member");
  const [lastToken, setLastToken] = useState<string | null>(null);
  const [inviteEmailed, setInviteEmailed] = useState(false);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [{ users: us }, { invites: inv }] = await Promise.all([
        listTeamUsers(),
        listInvites(),
      ]);
      setUsers(us);
      setInvites(inv);
      setError(null);
      try {
        const { entries } = await listAuditLog();
        setAudit(entries.slice(0, 20));
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

  async function onInvite(e: React.FormEvent) {
    e.preventDefault();
    if (!email.trim()) return;
    try {
      const inv = await createInvite({ email: email.trim(), role });
      setLastToken(inv.token ?? null);
      setInviteEmailed(Boolean(inv.email_sent));
      setEmail("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invite failed");
    }
  }

  return (
    <DashboardShell title="Team" breadcrumb="Settings">
      <SettingsPageBody>
        {error ? (
          <p className="mb-4 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-700">
            {error}
          </p>
        ) : null}

        <section className="mb-8 max-w-xl space-y-4 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
          <h2 className="text-lg font-medium">Invite member</h2>
          <form
            onSubmit={onInvite}
            className="flex flex-col gap-3 sm:flex-row sm:items-end"
          >
            <label className="flex flex-1 flex-col gap-1 text-sm">
              <span className="text-[var(--muted-foreground)]">Email</span>
              <input
                type="email"
                required
                placeholder="colleague@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="field-input"
              />
            </label>
            <SelectField
              label="Role"
              value={role}
              onChange={setRole}
              options={[
                { value: "admin", label: "Admin" },
                { value: "member", label: "Member" },
                { value: "viewer", label: "Viewer" },
              ]}
            />
            <Button type="submit">Send invite</Button>
          </form>
          {lastToken ? (
            <>
              {inviteEmailed ? (
                <p className="mb-2 text-sm text-green-700" role="status">
                  Invite email sent via SMTP.
                </p>
              ) : (
                <p className="mb-2 text-sm text-[var(--muted-foreground)]">
                  SMTP not configured — copy the link below (or set SMTP under
                  Instance settings).
                </p>
              )}
              <p className="text-sm text-[var(--muted-foreground)]">
                Share this link with the invitee (token in URL):{" "}
                <code className="break-all rounded bg-[var(--muted)]/30 px-1">
                  {typeof window !== "undefined"
                    ? `${window.location.origin}/accept-invite?token=${lastToken}`
                    : `/accept-invite?token=${lastToken}`}
                </code>
              </p>
            </>
          ) : null}
        </section>

        <section className="mb-8 space-y-3">
          <h2 className="text-lg font-medium">Users</h2>
          {loading ? (
            <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
          ) : (
            <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
              {users.map((u) => (
                <li
                  key={u.id}
                  className="flex justify-between px-4 py-3 text-sm"
                >
                  <span>
                    {u.name} · {u.email}
                  </span>
                  <span className="text-[var(--muted-foreground)]">
                    {u.role}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="space-y-3">
          <h2 className="text-lg font-medium">Pending invites</h2>
          <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
            {invites.length === 0 ? (
              <li className="px-4 py-4 text-sm text-[var(--muted-foreground)]">
                No pending invites.
              </li>
            ) : (
              invites.map((inv) => (
                <li
                  key={inv.id}
                  className="flex justify-between px-4 py-3 text-sm"
                >
                  <span>{inv.email}</span>
                  <span className="text-[var(--muted-foreground)]">
                    {inv.role}
                  </span>
                </li>
              ))
            )}
          </ul>
        </section>

        {audit.length > 0 ? (
          <section className="mt-8 space-y-3">
            <h2 className="text-lg font-medium">Recent audit log</h2>
            <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
              {audit.map((row) => (
                <li key={row.id} className="px-4 py-3 text-sm">
                  <span className="font-medium">{row.action}</span>
                  {row.target_kind ? (
                    <span className="text-[var(--muted-foreground)]">
                      {" "}
                      · {row.target_kind}
                      {row.target_id ? ` ${row.target_id}` : ""}
                    </span>
                  ) : null}
                </li>
              ))}
            </ul>
          </section>
        ) : null}
      </SettingsPageBody>
    </DashboardShell>
  );
}
