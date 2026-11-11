type ResourceListSectionProps = {
  title: string;
  countLabel?: string;
  filterValue?: string;
  onFilterChange?: (value: string) => void;
  filterPlaceholder?: string;
  filterAriaLabel?: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
};

export function ResourceListSection({
  title,
  countLabel,
  filterValue,
  onFilterChange,
  filterPlaceholder = "Filter by name or project",
  filterAriaLabel = "Filter list",
  children,
  footer,
}: ResourceListSectionProps) {
  const showFilter = onFilterChange !== undefined;

  return (
    <section aria-labelledby="resource-list-heading" className="mt-10">
      <div
        className="overflow-hidden rounded-[20px] border"
        style={{ background: "var(--surface)", borderColor: "var(--border)" }}
      >
        <div
          className="flex flex-wrap items-center justify-between gap-4 border-b px-6 py-5"
          style={{ borderColor: "var(--border)" }}
        >
          <h2
            id="resource-list-heading"
            className="text-xl font-semibold tracking-[-0.03em]"
          >
            {title}
          </h2>
          <div className="flex w-full min-w-0 flex-wrap items-center justify-end gap-3 sm:w-auto sm:flex-nowrap sm:gap-4">
            {showFilter ? (
              <label className="block w-full min-w-0 sm:w-56 md:w-64">
                <span className="sr-only">{filterAriaLabel}</span>
                <input
                  type="search"
                  value={filterValue ?? ""}
                  onChange={(e) => onFilterChange(e.target.value)}
                  placeholder={filterPlaceholder}
                  className="field-input w-full"
                  aria-label={filterAriaLabel}
                />
              </label>
            ) : null}
            {countLabel ? (
              <p
                className="shrink-0 font-mono text-[11px] uppercase tracking-widest tabular-nums"
                style={{ color: "var(--muted-text)" }}
              >
                {countLabel}
              </p>
            ) : null}
          </div>
        </div>
        <div className="min-w-0 px-6 py-2">{children}</div>
        {footer ? footer : null}
      </div>
    </section>
  );
}
