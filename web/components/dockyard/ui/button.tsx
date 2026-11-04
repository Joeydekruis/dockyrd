import type { LucideIcon } from "lucide-react";
import type { ComponentPropsWithoutRef, ReactNode } from "react";

export type ButtonVariant =
  | "primary"
  | "outline"
  | "secondary"
  | "link"
  | "danger"
  | "ghost"
  | "text";

const variantClass: Record<ButtonVariant, string> = {
  primary: "btn-primary",
  outline: "btn-outline",
  secondary: "btn-inverse",
  link: "btn-link",
  danger: "btn-danger",
  ghost: "btn-ghost",
  text: "btn-text",
};

type ButtonIconProps = {
  iconLeft?: LucideIcon;
  iconRight?: LucideIcon;
  children: ReactNode;
};

function ButtonIcons({
  iconLeft: IconLeft,
  iconRight: IconRight,
  children,
}: ButtonIconProps) {
  return (
    <>
      {IconLeft ? (
        <IconLeft strokeWidth={2} className="size-4 shrink-0" aria-hidden />
      ) : null}
      <span>{children}</span>
      {IconRight ? (
        <IconRight strokeWidth={2} className="size-4 shrink-0" aria-hidden />
      ) : null}
    </>
  );
}

type ButtonProps = {
  variant?: ButtonVariant;
  iconLeft?: LucideIcon;
  iconRight?: LucideIcon;
  iconOnly?: boolean;
} & (
  | (ComponentPropsWithoutRef<"button"> & { href?: undefined })
  | (ComponentPropsWithoutRef<"a"> & { href: string })
);

export function Button({
  variant = "primary",
  iconLeft,
  iconRight,
  iconOnly = false,
  className = "",
  children,
  ...props
}: ButtonProps) {
  const iconOnlyClass = iconOnly ? "btn-icon-only" : "";
  const base = `${variantClass[variant]} ${iconOnlyClass} ${className}`.trim();

  const IconOnly = iconLeft ?? iconRight;
  const content = iconOnly ? (
    IconOnly ? (
      <IconOnly strokeWidth={2} className="size-4 shrink-0" aria-hidden />
    ) : (
      children
    )
  ) : (
    <ButtonIcons iconLeft={iconLeft} iconRight={iconRight}>
      {children}
    </ButtonIcons>
  );

  if ("href" in props && props.href) {
    const { href, ...rest } = props;
    return (
      <a href={href} className={base} {...rest}>
        {content}
      </a>
    );
  }

  const {
    type = "button",
    disabled,
    ...rest
  } = props as ComponentPropsWithoutRef<"button">;
  const merged = disabled
    ? `${base} cursor-not-allowed disabled:opacity-50`
    : base;
  return (
    <button type={type} className={merged} disabled={disabled} {...rest}>
      {content}
    </button>
  );
}
