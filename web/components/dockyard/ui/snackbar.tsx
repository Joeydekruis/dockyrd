import { X } from "lucide-react";

type SnackbarProps = {
  message: React.ReactNode;
  action?: React.ReactNode;
  onDismiss?: () => void;
  tone?: "info" | "warn" | "success";
};

const toneStyles = {
  info: {
    border: "var(--border)",
    bg: "var(--surface)",
    text: "var(--ink-navy)",
  },
  warn: {
    border: "#fcd34d",
    bg: "#fffbeb",
    text: "#78350f",
  },
  success: {
    border: "#86efac",
    bg: "#f0fdf4",
    text: "#166534",
  },
};

export function Snackbar({
  message,
  action,
  onDismiss,
  tone = "info",
}: SnackbarProps) {
  const palette = toneStyles[tone];
  return (
    <div
      role="status"
      className="flex flex-wrap items-center justify-between gap-3 rounded-lg border px-4 py-2 text-sm shadow-lg"
      style={{
        borderColor: palette.border,
        background: palette.bg,
        color: palette.text,
      }}
    >
      <div className="min-w-0 flex-1">{message}</div>
      <div className="flex shrink-0 items-center gap-2">
        {action}
        {onDismiss ? (
          <button
            type="button"
            className="dockyard-icon-btn"
            aria-label="Dismiss"
            onClick={onDismiss}
          >
            <X strokeWidth={1.8} className="size-4" aria-hidden />
          </button>
        ) : null}
      </div>
    </div>
  );
}
