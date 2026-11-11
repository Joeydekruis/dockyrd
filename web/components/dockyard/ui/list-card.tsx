type ListCardProps = {
  title: string;
  description?: string;
  stat?: string;
  href?: string;
};

export function ListCard({
  title,
  description,
  stat,
  href = "#",
}: ListCardProps) {
  return (
    <a
      href={href}
      className="dockyrd-surface-interactive block rounded-[20px] border p-5"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="font-semibold">{title}</p>
          {description ? (
            <p
              className="mt-1 text-[13px]"
              style={{ color: "var(--muted-text)" }}
            >
              {description}
            </p>
          ) : null}
        </div>
        {stat ? (
          <span className="shrink-0 font-mono text-[11px] uppercase tracking-widest tabular-nums">
            {stat}
          </span>
        ) : null}
      </div>
    </a>
  );
}
