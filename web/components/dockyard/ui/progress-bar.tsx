type ProgressBarProps = {
  value?: number;
  variant?: "determinate" | "complete" | "error" | "indeterminate";
  label?: string;
};

export function ProgressBar({
  value = 0,
  variant = "determinate",
  label,
}: ProgressBarProps) {
  const pct = variant === "complete" ? 100 : Math.min(100, Math.max(0, value));
  const fill =
    variant === "error"
      ? "var(--danger)"
      : variant === "complete"
        ? "#1a7f4b"
        : "var(--accent-blue)";
  return (
    <div className="w-full">
      {label ? (
        <p
          className="mb-1 font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          {label}
        </p>
      ) : null}
      <div
        className="h-2 overflow-hidden rounded-full"
        style={{ background: "var(--hover-row)" }}
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={variant === "indeterminate" ? undefined : pct}
      >
        <div
          className={`h-full rounded-full transition-all ${variant === "indeterminate" ? "progress-indeterminate w-1/3" : ""}`}
          style={{
            width: variant === "indeterminate" ? undefined : `${pct}%`,
            background: fill,
          }}
        />
      </div>
    </div>
  );
}
