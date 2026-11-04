export type StatusVariant = "ok" | "warn" | "err" | "idle" | "info";

const styles: Record<StatusVariant, { fg: string; bg: string; dot: string }> = {
  ok: { fg: "#176b3d", bg: "#e6f4ec", dot: "#1a7f4b" },
  warn: { fg: "#8a5200", bg: "#fbf0dc", dot: "#d98a00" },
  err: { fg: "#b3261e", bg: "#fbe8e6", dot: "#b3261e" },
  idle: { fg: "#4a5361", bg: "#eceef1", dot: "#6b7686" },
  info: { fg: "#0b3d91", bg: "#e3f0ff", dot: "#3b93f5" },
};

type StatusBadgeProps = {
  label: string;
  variant?: StatusVariant;
};

export function StatusBadge({ label, variant = "ok" }: StatusBadgeProps) {
  const tone = styles[variant];
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-md px-2 py-1 font-mono text-[11px] uppercase tracking-widest"
      style={{ color: tone.fg, background: tone.bg }}
    >
      <span
        className="inline-block size-1.5 rounded-full"
        style={{ background: tone.dot }}
        aria-hidden
      />
      {label}
    </span>
  );
}
