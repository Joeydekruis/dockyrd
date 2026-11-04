import { Chip } from "@/components/dockyard/ui/chip";
import { Button } from "@/components/dockyard/ui/button";

type TemplateCardProps = {
  name: string;
  description: string;
  stack?: string;
  recommended?: boolean;
  popular?: boolean;
  actionLabel?: string;
  actionDisabled?: boolean;
  href?: string;
  onSelect?: () => void;
};

export function TemplateCard({
  name,
  description,
  stack,
  recommended,
  popular,
  actionLabel,
  actionDisabled,
  href,
  onSelect,
}: TemplateCardProps) {
  const badge = recommended ? "Recommended" : popular ? "Popular" : null;
  const className =
    "relative flex h-full flex-col rounded-[20px] border p-5 text-left";
  const style = { background: "var(--surface)", borderColor: "var(--border)" };

  const body = (
    <>
      {badge ? (
        <span className="absolute right-4 top-4">
          <Chip>{badge}</Chip>
        </span>
      ) : null}
      <p className="pr-24 font-semibold">{name}</p>
      <p className="mt-2 flex-1 text-[13px]" style={{ color: "var(--muted-text)" }}>
        {description}
      </p>
      {stack ? (
        <p
          className="mt-3 font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          {stack}
        </p>
      ) : null}
      {actionLabel ? (
        <Button
          type="button"
          variant="outline"
          className="mt-5 w-full"
          disabled={actionDisabled}
          onClick={onSelect}
        >
          {actionLabel}
        </Button>
      ) : null}
    </>
  );

  if (href && !actionLabel) {
    return (
      <a href={href} className={`${className} transition-colors hover:bg-[var(--hover-row)]`} style={style}>
        {body}
      </a>
    );
  }

  return (
    <article className={className} style={style}>
      {body}
    </article>
  );
}
