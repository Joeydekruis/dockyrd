import { Chip } from "@/components/dockyard/ui/chip";

type TemplateCardProps = {
  name: string;
  description: string;
  stack?: string;
  recommended?: boolean;
  href?: string;
  onSelect?: () => void;
};

export function TemplateCard({
  name,
  description,
  stack,
  recommended,
  href,
  onSelect,
}: TemplateCardProps) {
  const className =
    "relative block rounded-[20px] border p-5 text-left transition-colors hover:bg-[var(--hover-row)]";
  const style = { background: "var(--surface)", borderColor: "var(--border)" };
  const body = (
    <>
      {recommended ? (
        <span className="absolute right-4 top-4">
          <Chip>Recommended</Chip>
        </span>
      ) : null}
      <p className="pr-24 font-semibold">{name}</p>
      <p className="mt-2 text-[13px]" style={{ color: "var(--muted-text)" }}>
        {description}
      </p>
      {stack ? (
        <p className="mt-3 font-mono text-[11px] uppercase tracking-widest">
          {stack}
        </p>
      ) : null}
    </>
  );
  if (href) {
    return (
      <a href={href} className={className} style={style}>
        {body}
      </a>
    );
  }
  return (
    <button
      type="button"
      onClick={onSelect}
      className={className}
      style={style}
    >
      {body}
    </button>
  );
}
