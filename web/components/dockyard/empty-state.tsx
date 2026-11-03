type EmptyStateProps = {
  title: string;
  description: string;
  actionLabel?: string;
  actionHref?: string;
};

export function EmptyState({
  title,
  description,
  actionLabel,
  actionHref,
}: EmptyStateProps) {
  return (
    <section
      className="rounded-[20px] border px-8 py-12 text-center"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <h2 className="text-xl font-semibold tracking-tight">{title}</h2>
      <p className="mx-auto mt-2 max-w-md" style={{ color: "var(--muted-text)" }}>
        {description}
      </p>
      {actionLabel && actionHref ? (
        <a
          href={actionHref}
          className="mt-6 inline-flex min-h-11 items-center rounded-[10px] px-5 font-mono text-xs uppercase tracking-wide text-white"
          style={{ background: "var(--ink-navy)" }}
        >
          {actionLabel}
        </a>
      ) : null}
    </section>
  );
}
