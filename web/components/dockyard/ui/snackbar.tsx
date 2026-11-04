import { X } from "lucide-react";

type SnackbarProps = {
  message: React.ReactNode;
  action?: React.ReactNode;
  onDismiss?: () => void;
  tone?: "info" | "warn" | "success";
  /** Floating toasts use shadow; inline banners (e.g. update notice) do not. */
  elevated?: boolean;
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
  elevated = true,
}: SnackbarProps) {
  const palette = toneStyles[tone];
  return (
    <div
      role="status"
      className={`flex flex-wrap items-center justify-between gap-3 rounded-lg border px-4 py-2 text-sm${elevated ? " shadow-lg" : ""}`}
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

type SnackbarViewportProps = {
  open: boolean;
  message: React.ReactNode;
  tone?: SnackbarProps["tone"];
  onDismiss: () => void;
};

/** Fixed bottom snackbar for save confirmations (see Design/Components/Snackbar). */
export function SnackbarViewport({
  open,
  message,
  tone = "info",
  onDismiss,
}: SnackbarViewportProps) {
  if (!open) return null;
  return (
    <div className="pointer-events-none fixed inset-x-0 bottom-6 z-50 flex justify-center px-4">
      <div className="pointer-events-auto w-full max-w-md">
        <Snackbar message={message} tone={tone} onDismiss={onDismiss} />
      </div>
    </div>
  );
}
