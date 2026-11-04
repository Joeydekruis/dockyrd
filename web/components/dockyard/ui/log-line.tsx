import { Chip } from "@/components/dockyard/ui/chip";

type LogLineProps = {
  timestamp: string;
  status: number;
  method: string;
  path: string;
};

function statusTone(status: number): "ok" | "warn" | "err" | "info" {
  if (status >= 500) return "err";
  if (status >= 400) return "warn";
  if (status >= 300) return "info";
  return "ok";
}

const toneColor: Record<string, string> = {
  ok: "#176b3d",
  warn: "#8a5200",
  err: "#b3261e",
  info: "#0b3d91",
};

export function LogLine({ timestamp, status, method, path }: LogLineProps) {
  const tone = statusTone(status);
  return (
    <div
      className="flex flex-wrap items-center gap-2 border-t py-2 font-mono text-[12px] first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <span className="tabular-nums" style={{ color: "var(--muted-text)" }}>
        {timestamp}
      </span>
      <span style={{ color: toneColor[tone], fontWeight: 600 }}>{status}</span>
      <Chip>{method}</Chip>
      <span className="min-w-0 truncate" style={{ color: "var(--ink-navy)" }}>
        {path}
      </span>
    </div>
  );
}
