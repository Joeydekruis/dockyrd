import type { LucideIcon } from "lucide-react";
import type { ButtonHTMLAttributes } from "react";

type IconButtonProps = {
  icon: LucideIcon;
  label: string;
  iconClassName?: string;
} & ButtonHTMLAttributes<HTMLButtonElement>;

export function IconButton({
  icon: Icon,
  label,
  iconClassName = "size-5",
  className = "",
  type = "button",
  ...props
}: IconButtonProps) {
  return (
    <button
      type={type}
      aria-label={label}
      title={label}
      className={`dockyrd-icon-btn ${className}`.trim()}
      {...props}
    >
      <Icon strokeWidth={1.8} className={iconClassName} aria-hidden />
    </button>
  );
}
