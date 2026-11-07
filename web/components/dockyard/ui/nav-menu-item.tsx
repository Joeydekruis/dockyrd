import type { LucideIcon } from "lucide-react";

type NavMenuItemProps = {
  label: string;
  href?: string;
  icon: LucideIcon;
  active?: boolean;
  onClick?: () => void;
};

export function NavMenuItem({
  label,
  href = "#",
  icon: Icon,
  active = false,
  onClick,
}: NavMenuItemProps) {
  return (
    <a
      href={href}
      aria-current={active ? "page" : undefined}
      className="dockyrd-nav-link"
      onClick={onClick}
    >
      <Icon strokeWidth={1.8} className="size-[18px] shrink-0" aria-hidden />
      <span>{label}</span>
    </a>
  );
}
