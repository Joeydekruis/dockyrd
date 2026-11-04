type StatCardProps = {
  label: string;
  value: string | number;
  hint?: string;
};

export function StatCard({ label, value, hint }: StatCardProps) {
  return (
    <article
      className="rounded-[20px] border p-5"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <p
        className="font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {label}
      </p>
      <p className="mt-2 text-[40px] font-semibold leading-none tracking-tight tabular-nums">
        {value}
      </p>
      {hint ? (
        <p className="mt-2 text-[13px]" style={{ color: "var(--muted-text)" }}>
          {hint}
        </p>
      ) : null}
    </article>
  );
}
