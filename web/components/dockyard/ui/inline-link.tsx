import type { LucideIcon } from "lucide-react";
import { ExternalLink } from "lucide-react";
import type { AnchorHTMLAttributes, ReactNode } from "react";

type InlineLinkProps = {
  href: string;
  external?: boolean;
  iconRight?: LucideIcon;
  children: ReactNode;
} & Omit<AnchorHTMLAttributes<HTMLAnchorElement>, "href" | "children">;

export function InlineLink({
  href,
  external,
  iconRight: IconRight,
  children,
  className = "",
  ...props
}: InlineLinkProps) {
  const isExternal = external ?? href.startsWith("http");
  const Trailing = IconRight ?? (isExternal ? ExternalLink : undefined);
  return (
    <a
      href={href}
      className={`inline-flex items-center gap-1 font-medium hover:underline ${className}`.trim()}
      style={{ color: "var(--link-blue)" }}
      {...(isExternal ? { target: "_blank", rel: "noopener noreferrer" } : {})}
      {...props}
    >
      {children}
      {Trailing ? (
        <Trailing strokeWidth={2} className="size-3.5 shrink-0" aria-hidden />
      ) : null}
    </a>
  );
}
