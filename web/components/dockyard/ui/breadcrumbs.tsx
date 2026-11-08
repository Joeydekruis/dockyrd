export type BreadcrumbItem = { label: string; href?: string };

type BreadcrumbsProps = {
  items: BreadcrumbItem[];
};

export function Breadcrumbs({ items }: BreadcrumbsProps) {
  return (
    <nav
      aria-label="Breadcrumb"
      className="text-sm"
      style={{ color: "var(--muted-text)" }}
    >
      {items.map((item, index) => {
        const isLast = index === items.length - 1;
        const key = item.href ?? item.label;
        return (
          <span key={key}>
            {index > 0 ? (
              <span className="mx-2" aria-hidden>
                /
              </span>
            ) : null}
            {isLast ? (
              <span style={{ color: "var(--ink-navy)" }}>{item.label}</span>
            ) : item.href ? (
              <a href={item.href} className="hover:underline">
                {item.label}
              </a>
            ) : (
              <span>{item.label}</span>
            )}
          </span>
        );
      })}
    </nav>
  );
}
