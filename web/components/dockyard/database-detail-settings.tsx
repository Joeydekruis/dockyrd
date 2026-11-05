"use client";

import { TriangleAlert } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { DangerZone } from "@/components/dockyard/ui/danger-zone";
import { Field } from "@/components/dockyard/ui/field";
import { Modal } from "@/components/dockyard/ui/modal";
import { SelectField } from "@/components/dockyard/ui/select-field";
import {
  SettingsDetailBlock,
  SettingsEnabledStatus,
  SettingsGroupHeading,
} from "@/components/dockyard/ui/settings-detail-block";
import {
  type DatabaseBackupPolicy,
  type DatabaseBackupRun,
  type DatabaseRow,
  deleteDatabaseService,
  getDatabaseBackupPolicy,
  getServiceConnection,
  listDatabaseBackupRuns,
  patchDatabaseBackupPolicy,
  restoreDatabaseBackup,
  triggerDatabaseBackup,
} from "@/lib/api";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";
import { databaseEngineLabel } from "@/lib/service-display";

function formatUtcRunTime(runAtMinute: number) {
  const h = Math.floor(runAtMinute / 60);
  const m = runAtMinute % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")} UTC`;
}

function dbNameFromConnection(vars: Record<string, string>): string | null {
  const url = vars.DATABASE_URL;
  if (!url) {
    return null;
  }
  try {
    const u = new URL(url);
    const name = u.pathname.replace(/^\//, "");
    return name || null;
  } catch {
    return null;
  }
}

type DatabaseDetailSettingsProps = {
  row: DatabaseRow;
};

export function DatabaseDetailSettings({ row }: DatabaseDetailSettingsProps) {
  const router = useRouter();
  const [policy, setPolicy] = useState<DatabaseBackupPolicy | null>(null);
  const [backupModal, setBackupModal] = useState(false);
  const [backupForm, setBackupForm] = useState({
    enabled: false,
    interval: "daily" as DatabaseBackupPolicy["interval"],
    retention_days: 30,
    local_enabled: true,
    remote_enabled: false,
    run_at_minute: 180,
    run_weekday: 0,
  });
  const [backupBusy, setBackupBusy] = useState(false);
  const [backupError, setBackupError] = useState<string | null>(null);
  const [runs, setRuns] = useState<DatabaseBackupRun[]>([]);
  const [runNowBusy, setRunNowBusy] = useState(false);
  const [restoreOpen, setRestoreOpen] = useState(false);
  const [restoreRunId, setRestoreRunId] = useState<string | null>(null);
  const [restoreConfirm, setRestoreConfirm] = useState("");
  const [restoreDbNameHint, setRestoreDbNameHint] = useState<string | null>(null);
  const [restoreBusy, setRestoreBusy] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteBusy, setDeleteBusy] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [confirmSlug, setConfirmSlug] = useState("");

  const loadPolicy = useCallback(async () => {
    try {
      const [p, runList] = await Promise.all([
        getDatabaseBackupPolicy(row.id),
        listDatabaseBackupRuns(row.id),
      ]);
      setPolicy(p);
      setRuns(runList.runs);
    } catch {
      setPolicy(null);
      setRuns([]);
    }
  }, [row.id]);

  useEffect(() => {
    void loadPolicy();
  }, [loadPolicy]);

  function openBackupModal() {
    if (policy) {
      setBackupForm({
        enabled: policy.enabled,
        interval: policy.interval,
        retention_days: policy.retention_days,
        local_enabled: policy.local_enabled,
        remote_enabled: policy.remote_enabled,
        run_at_minute: policy.run_at_minute ?? 180,
        run_weekday: policy.run_weekday ?? 0,
      });
    }
    setBackupError(null);
    setBackupModal(true);
  }

  async function openRestoreModal(runId: string) {
    setRestoreRunId(runId);
    setRestoreConfirm("");
    setRestoreDbNameHint(null);
    setBackupError(null);
    if (row.kind === "postgres") {
      try {
        const conn = await getServiceConnection(row.id, true);
        setRestoreDbNameHint(dbNameFromConnection(conn.variables));
      } catch {
        setRestoreDbNameHint(null);
      }
    }
    setRestoreOpen(true);
  }

  async function confirmRestore() {
    if (!restoreRunId) {
      return;
    }
    setRestoreBusy(true);
    setBackupError(null);
    try {
      await restoreDatabaseBackup(row.id, restoreRunId, {
        confirm: restoreConfirm.trim(),
      });
      setRestoreOpen(false);
      await loadPolicy();
    } catch (e) {
      setBackupError(e instanceof Error ? e.message : "Restore failed");
    } finally {
      setRestoreBusy(false);
    }
  }

  async function runBackupNow() {
    setRunNowBusy(true);
    setBackupError(null);
    try {
      await triggerDatabaseBackup(row.id);
      await loadPolicy();
    } catch (e) {
      setBackupError(e instanceof Error ? e.message : "Backup failed to queue");
    } finally {
      setRunNowBusy(false);
    }
  }

  async function saveBackupPolicy() {
    setBackupBusy(true);
    setBackupError(null);
    try {
      const updated = await patchDatabaseBackupPolicy(row.id, backupForm);
      setPolicy(updated);
      setBackupModal(false);
    } catch (e) {
      setBackupError(e instanceof Error ? e.message : "Save failed");
    } finally {
      setBackupBusy(false);
    }
  }

  async function confirmDelete() {
    if (confirmSlug.trim() !== row.slug) {
      setDeleteError(`Type the slug "${row.slug}" to confirm.`);
      return;
    }
    setDeleteBusy(true);
    setDeleteError(null);
    try {
      await deleteDatabaseService(row.id, {
        confirm_slug: row.slug,
        delete_volume: true,
      });
      router.push("/databases");
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : "Delete failed");
    } finally {
      setDeleteBusy(false);
    }
  }

  const intervalLabel =
    policy?.interval === "hourly"
      ? "Every hour"
      : policy?.interval === "weekly"
        ? `Weekly · ${formatUtcRunTime(policy.run_at_minute ?? 180)}`
        : policy?.enabled
          ? `Daily · ${formatUtcRunTime(policy.run_at_minute ?? 180)}`
          : "Every day";

  return (
    <>
      <section
        className={RESOURCE_DETAIL_CARD}
        style={resourceDetailCardStyle()}
        aria-label="Settings"
      >
        <h2 className="text-lg font-semibold tracking-tight">Settings</h2>
        <SettingsDetailBlock
          title="Engine"
          description="Major version upgrades require dump and restore — not available in the panel yet."
          rows={[
            { label: "Type", value: databaseEngineLabel(row.kind) },
            {
              label: "Slug",
              value: <span className="font-mono">{row.slug}</span>,
            },
          ]}
          actionDisabled
        />
        <SettingsDetailBlock
          title="Backups"
          description="Scheduled dumps on this server and optional upload to the server S3 target."
          rows={[
            {
              label: "Schedule",
              value: policy?.enabled ? intervalLabel : "Disabled",
            },
            {
              label: "Retention",
              value: policy
                ? `${policy.retention_days} days`
                : "30 days (default)",
            },
            {
              label: "Local copy",
              value: (
                <SettingsEnabledStatus
                  enabled={policy?.local_enabled ?? true}
                />
              ),
            },
            {
              label: "Remote (S3)",
              value: (
                <SettingsEnabledStatus
                  enabled={policy?.remote_enabled ?? false}
                />
              ),
            },
          ]}
          actionLabel="Configure"
          onAction={openBackupModal}
        />
        <div className="mt-4 flex flex-wrap gap-3">
          <Button
            type="button"
            variant="outline"
            disabled={runNowBusy || !policy?.enabled}
            onClick={() => void runBackupNow()}
          >
            {runNowBusy ? "Queuing…" : "Backup now"}
          </Button>
        </div>
        {runs.length > 0 ? (
          <ul
            className="mt-4 space-y-2 text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            {runs.slice(0, 5).map((run) => (
              <li
                key={run.id}
                className="flex flex-wrap items-center justify-between gap-2 font-mono text-[12px]"
              >
                <span>
                  {run.status} · {run.destination}
                  {run.verified ? " · verified" : ""}
                  {run.started_at
                    ? ` · ${new Date(run.started_at * 1000).toLocaleString()}`
                    : ""}
                </span>
                {run.status === "succeeded" && row.kind === "postgres" ? (
                  <Button
                    type="button"
                    variant="outline"
                    className="text-[11px]"
                    disabled={restoreBusy}
                    onClick={() => void openRestoreModal(run.id)}
                  >
                    Restore
                  </Button>
                ) : null}
                {run.status === "succeeded" && row.kind === "redis" ? (
                  <span className="text-[11px]">Restore not available</span>
                ) : null}
              </li>
            ))}
          </ul>
        ) : null}
        {backupError ? (
          <p className="mt-2 text-sm text-red-600" role="alert">
            {backupError}
          </p>
        ) : null}
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Remote backups need a server target in{" "}
          <Link
            href="/settings/server/backups"
            className="underline"
            style={{ color: "var(--link-blue)" }}
          >
            Server → Backups
          </Link>
          .
        </p>
      </section>

      <div className="mt-8">
        <SettingsGroupHeading
          icon={TriangleAlert}
          label="Danger zone"
          tone="danger"
        />
        <DangerZone
          title="Delete database"
          description="Stops the container and removes all data on this server. Apps must stop referencing this connection first."
          action={
            <Button
              type="button"
              variant="danger"
              className="font-mono text-[11px] uppercase tracking-widest"
              onClick={() => {
                setConfirmSlug("");
                setDeleteError(null);
                setDeleteOpen(true);
              }}
            >
              Delete database
            </Button>
          }
        />
      </div>

      <Modal
        open={backupModal}
        title="Backup schedule"
        description="Dockyard runs pg_dump (Postgres) or Valkey RDB snapshots on this interval."
        onClose={() => setBackupModal(false)}
        footer={
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => setBackupModal(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              disabled={backupBusy}
              onClick={() => void saveBackupPolicy()}
            >
              Save
            </Button>
          </>
        }
      >
        {backupError ? (
          <p className="mb-4 text-sm text-red-600" role="alert">
            {backupError}
          </p>
        ) : null}
        <div className="space-y-4">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={backupForm.enabled}
              onChange={(e) =>
                setBackupForm((f) => ({ ...f, enabled: e.target.checked }))
              }
            />
            Enable scheduled backups
          </label>
          <SelectField
            label="Interval"
            value={backupForm.interval}
            onChange={(value) =>
              setBackupForm((f) => ({
                ...f,
                interval: value as DatabaseBackupPolicy["interval"],
              }))
            }
            options={[
              { value: "hourly", label: "Every hour" },
              { value: "daily", label: "Every day" },
              { value: "weekly", label: "Every week" },
            ]}
          />
          <Field
            label="Run at (UTC, minutes from midnight)"
            type="number"
            min={0}
            max={1439}
            value={String(backupForm.run_at_minute)}
            onChange={(e) =>
              setBackupForm((f) => ({
                ...f,
                run_at_minute: Number.parseInt(e.target.value, 10) || 0,
              }))
            }
          />
          {backupForm.interval === "weekly" ? (
            <SelectField
              label="Weekday (UTC)"
              value={String(backupForm.run_weekday)}
              onChange={(value) =>
                setBackupForm((f) => ({
                  ...f,
                  run_weekday: Number.parseInt(value, 10) || 0,
                }))
              }
              options={[
                { value: "0", label: "Sunday" },
                { value: "1", label: "Monday" },
                { value: "2", label: "Tuesday" },
                { value: "3", label: "Wednesday" },
                { value: "4", label: "Thursday" },
                { value: "5", label: "Friday" },
                { value: "6", label: "Saturday" },
              ]}
            />
          ) : null}
          <Field
            label="Retention (days)"
            type="number"
            min={1}
            value={String(backupForm.retention_days)}
            onChange={(e) =>
              setBackupForm((f) => ({
                ...f,
                retention_days: Number.parseInt(e.target.value, 10) || 30,
              }))
            }
          />
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={backupForm.local_enabled}
              onChange={(e) =>
                setBackupForm((f) => ({
                  ...f,
                  local_enabled: e.target.checked,
                }))
              }
            />
            Keep copies on this server
          </label>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={backupForm.remote_enabled}
              onChange={(e) =>
                setBackupForm((f) => ({
                  ...f,
                  remote_enabled: e.target.checked,
                }))
              }
            />
            Upload to server S3 target
          </label>
        </div>
      </Modal>

      <Modal
        open={restoreOpen}
        title="Restore backup"
        description="This replaces the live database. A local snapshot is taken immediately before restore."
        onClose={() => setRestoreOpen(false)}
        footer={
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => setRestoreOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="danger"
              disabled={restoreBusy || !restoreConfirm.trim()}
              onClick={() => void confirmRestore()}
            >
              {restoreBusy ? "Queuing…" : "Restore"}
            </Button>
          </>
        }
      >
        <p className="mb-4 text-sm" style={{ color: "var(--muted-text)" }}>
          Type the Postgres database name to confirm
          {restoreDbNameHint ? (
            <>
              {" "}
              (<span className="font-mono">{restoreDbNameHint}</span>)
            </>
          ) : null}
          .
        </p>
        <Field
          label="Database name"
          value={restoreConfirm}
          onChange={(e) => setRestoreConfirm(e.target.value)}
          autoFocus
        />
      </Modal>

      <Modal
        open={deleteOpen}
        title="Delete database"
        description={`This permanently deletes "${row.name}" and its volume.`}
        onClose={() => setDeleteOpen(false)}
        footer={
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => setDeleteOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="danger"
              disabled={deleteBusy}
              onClick={() => void confirmDelete()}
            >
              Delete
            </Button>
          </>
        }
      >
        {deleteError ? (
          <p className="mb-4 text-sm text-red-600" role="alert">
            {deleteError}
          </p>
        ) : null}
        <Field
          label={`Confirm slug (${row.slug})`}
          value={confirmSlug}
          onChange={(e) => setConfirmSlug(e.target.value)}
          autoFocus
        />
      </Modal>
    </>
  );
}
