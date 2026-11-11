import { Plus } from "lucide-react";
import type { ReactNode } from "react";

type AddCardProps = {
  title: string;
  description?: string;
  onClick?: () => void;
  href?: string;
  children?: ReactNode;
};

export function AddCard({
  title,
  description,
  onClick,
  href,
  children,
}: AddCardProps) {
  const inner = (
    <>
      <Plus
        strokeWidth={2}
        className="mb-2 size-6 text-[var(--link-blue)]"
        aria-hidden
      />
      <p className="text-sm font-semibold">{title}</p>
      {description ? (
        <p className="mt-1 text-[13px]" style={{ color: "var(--muted-text)" }}>
          {description}
        </p>
      ) : null}
      {children}
    </>
  );
  const className =
    "dockyrd-surface-interactive flex min-h-[120px] flex-col items-center justify-center rounded-[20px] border border-dashed p-6 text-center";
  const style = {
    borderColor: "var(--input-border)",
    color: "var(--ink-navy)",
  };
  if (href) {
    return (
      <a href={href} className={className} style={style}>
        {inner}
      </a>
    );
  }
  return (
    <button type="button" onClick={onClick} className={className} style={style}>
      {inner}
    </button>
  );
}
