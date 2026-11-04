import type { LucideIcon } from "lucide-react";
import { AlertCircle, CheckCircle2, Info, TriangleAlert } from "lucide-react";
import type { ReactNode } from "react";

export type AlertTone = "info" | "success" | "warning" | "error";

const toneStyles: Record<
  AlertTone,
  { bg: string; fg: string; border: string; Icon: LucideIcon }
> = {
  info: { bg: "#e3f0ff", fg: "#0b3d91", border: "#b8d4f5", Icon: Info },
  success: {
    bg: "#e6f4ec",
    fg: "#176b3d",
    border: "#b8dfc8",
    Icon: CheckCircle2,
  },
  warning: {
    bg: "#fbf0dc",
    fg: "#8a5200",
    border: "#ecd9a8",
    Icon: TriangleAlert,
  },
  error: { bg: "#fbe8e6", fg: "#b3261e", border: "#f0c4c0", Icon: AlertCircle },
};

type AlertProps = {
  tone?: AlertTone;
  message: ReactNode;
  action?: ReactNode;
};

export function Alert({ tone = "info", message, action }: AlertProps) {
  const t = toneStyles[tone];
  const Icon = t.Icon;
  return (
    <div
      role="status"
      className="flex flex-wrap items-start gap-3 rounded-xl border px-4 py-3 text-sm"
      style={{ background: t.bg, borderColor: t.border, color: t.fg }}
    >
      <Icon strokeWidth={1.8} className="mt-0.5 size-4 shrink-0" aria-hidden />
      <p className="min-w-0 flex-1">{message}</p>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  );
}
