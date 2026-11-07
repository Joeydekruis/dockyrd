import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyard/ui/status-badge";

export type ContainerTableRow = {
  id: string;
  name: string;
  image: string;
  statusLabel: string;
  statusVariant?: StatusVariant;
  cpu?: string;
  memory?: string;
};

type ContainerTableProps = {
  rows: ContainerTableRow[];
};

export function ContainerTable({ rows }: ContainerTableProps) {
  return (
    <div
      className="overflow-x-auto rounded-[20px] border"
      style={{ borderColor: "var(--border)" }}
    >
      <table className="w-full min-w-[560px] border-collapse text-left text-sm">
        <thead>
          <tr style={{ background: "var(--hover-row)" }}>
            {["Container", "Image", "Status", "CPU", "Memory"].map((h) => (
              <th
                key={h}
                className="px-4 py-3 font-mono text-[11px] font-medium uppercase tracking-widest"
                style={{ color: "var(--muted-text)" }}
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr
              key={row.id}
              className="border-t"
              style={{ borderColor: "var(--border)" }}
            >
              <td className="px-4 py-3 font-medium">{row.name}</td>
              <td
                className="px-4 py-3 font-mono text-[13px]"
                style={{ color: "var(--muted-text)" }}
              >
                {row.image}
              </td>
              <td className="px-4 py-3">
                <StatusBadge
                  label={row.statusLabel}
                  variant={row.statusVariant ?? "ok"}
                />
              </td>
              <td className="px-4 py-3 font-mono tabular-nums">
                {row.cpu ?? "—"}
              </td>
              <td className="px-4 py-3 font-mono tabular-nums">
                {row.memory ?? "—"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
