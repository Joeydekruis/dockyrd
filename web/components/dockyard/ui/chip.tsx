type ChipProps = {
  children: React.ReactNode;
};

export function Chip({ children }: ChipProps) {
  return (
    <span
      className="inline-flex items-center rounded-md px-2 py-0.5 font-mono text-[11px] uppercase tracking-widest"
      style={{ background: "#f1f0ed", color: "var(--ink-navy)" }}
    >
      {children}
    </span>
  );
}
