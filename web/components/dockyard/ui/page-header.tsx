type PageHeaderProps = {
  eyebrow?: string;
  title: string;
  subtitle?: string;
  action?: React.ReactNode;
};

export function PageHeader({
  eyebrow,
  title,
  subtitle,
  action,
}: PageHeaderProps) {
  return (
    <header className="mb-8 flex flex-wrap items-start justify-between gap-4">
      <div>
        {eyebrow ? (
          <p
            className="mb-2 font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            {eyebrow}
          </p>
        ) : null}
        <h1 className="text-[44px] font-semibold leading-[1.04] tracking-[-0.04em]">
          {title}
        </h1>
        {subtitle ? (
          <p className="mt-2 max-w-2xl" style={{ color: "var(--muted-text)" }}>
            {subtitle}
          </p>
        ) : null}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </header>
  );
}
