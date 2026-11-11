import { Plus } from "lucide-react";

type ResourceListAddLinkProps = {
  label: string;
  onClick?: () => void;
  href?: string;
};

export function ResourceListAddLink({
  label,
  onClick,
  href,
}: ResourceListAddLinkProps) {
  const className =
    "dockyrd-surface-interactive dockyrd-text-link flex w-full items-center justify-center gap-2 border-t py-4 text-sm no-underline";
  const style = { borderColor: "var(--border)" };
  const content = (
    <>
      <Plus strokeWidth={2} className="size-4 shrink-0" aria-hidden />
      {label}
    </>
  );
  if (href) {
    return (
      <a href={href} className={className} style={style}>
        {content}
      </a>
    );
  }
  return (
    <button type="button" onClick={onClick} className={className} style={style}>
      {content}
    </button>
  );
}
