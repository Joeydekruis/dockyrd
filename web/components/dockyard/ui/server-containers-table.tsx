import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyrd/ui/status-badge";

export type ServerContainersTableRow = {
  id: string;
  name: string;
  projectLabel: string;
  cpu?: string;
  memory?: string;
  uptime?: string;
  statusLabel: string;
  statusVariant?: StatusVariant;
};

type ServerContainersTableProps = {
  rows: ServerContainersTableRow[];
  runningCount?: number;
};

export function ServerContainersTable({
  rows,
  runningCount,
}: ServerContainersTableProps) {
  return (
    <div
      className="overflow-x-auto rounded-[20px] border"
      style={{ borderColor: "var(--border)" }}
    >
      <div
        className="flex items-center justify-between border-b px-5 py-4"
        style={{ borderColor: "var(--border)" }}
      >
        <h2 className="text-lg font-semibold">Running containers</h2>
        {runningCount !== undefined ? (
          <span
            className="rounded-md px-2 py-0.5 font-mono text-[11px] uppercase tracking-widest"
            style={{
              background: "var(--hover-row)",
              color: "var(--muted-text)",
            }}
          >
            {runningCount} running
          </span>
        ) : null}
      </div>
      <table className="w-full min-w-[720px] border-collapse text-left text-sm">
        <thead>
          <tr style={{ background: "var(--hover-row)" }}>
            {["Container", "Project", "CPU", "Memory", "Uptime", "Status"].map(
              (h) => (
                <th
                  key={h}
                  className="px-4 py-3 font-mono text-[11px] font-medium uppercase tracking-widest"
                  style={{ color: "var(--muted-text)" }}
                >
                  {h}
                </th>
              ),
            )}
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td
                colSpan={6}
                className="px-4 py-8 text-center text-[var(--muted-text)]"
              >
                No running dockyrd containers.
              </td>
            </tr>
          ) : (
            rows.map((row) => (
              <tr
                key={row.id}
                className="border-t"
                style={{ borderColor: "var(--border)" }}
              >
                <td className="px-4 py-3 font-mono text-[13px]">{row.name}</td>
                <td className="px-4 py-3">{row.projectLabel}</td>
                <td className="px-4 py-3 font-mono tabular-nums">
                  {row.cpu ?? "—"}
                </td>
                <td className="px-4 py-3 font-mono tabular-nums">
                  {row.memory ?? "—"}
                </td>
                <td className="px-4 py-3 font-mono tabular-nums">
                  {row.uptime ?? "—"}
                </td>
                <td className="px-4 py-3">
                  <StatusBadge
                    label={row.statusLabel}
                    variant={row.statusVariant ?? "ok"}
                  />
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
