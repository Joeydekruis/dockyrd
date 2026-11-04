type DividerProps = {
  label?: string;
};

export function Divider({ label }: DividerProps) {
  if (!label) {
    return (
      <hr
        className="border-0 border-t"
        style={{ borderColor: "var(--border)" }}
      />
    );
  }
  return (
    <div className="flex items-center gap-3">
      <hr
        className="min-w-0 flex-1 border-0 border-t"
        style={{ borderColor: "var(--border)" }}
      />
      <span
        className="shrink-0 font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {label}
      </span>
      <hr
        className="min-w-0 flex-1 border-0 border-t"
        style={{ borderColor: "var(--border)" }}
      />
    </div>
  );
}
