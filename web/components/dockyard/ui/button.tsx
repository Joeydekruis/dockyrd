import type { LucideIcon } from "lucide-react";
import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { Spinner } from "@/components/dockyard/ui/spinner";

/** `outline` = bordered secondary action on light surfaces (design-index “secondary”). `secondary` = white inverse — gradient/dark backgrounds only. */
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
  /** Replaces label with a spinner; keeps button height and uses a stable min-width. */
  loading?: boolean;
} & (
  | (ComponentPropsWithoutRef<"button"> & { href?: undefined })
  | (ComponentPropsWithoutRef<"a"> & { href: string })
);

export function Button({
  variant = "primary",
  iconLeft,
  iconRight,
  iconOnly = false,
  loading = false,
  className = "",
  children,
  ...props
}: ButtonProps) {
  const iconOnlyClass = iconOnly ? "btn-icon-only" : "";
  const loadingClass = loading ? "pointer-events-none min-w-[5.25rem]" : "";
  const base =
    `${variantClass[variant]} ${iconOnlyClass} ${loadingClass} ${className}`.trim();

  const IconOnly = iconLeft ?? iconRight;
  const spinnerInverse =
    loading && (variant === "primary" || variant === "danger");
  const content = loading ? (
    <Spinner size="sm" label="Loading" inverse={spinnerInverse} />
  ) : iconOnly ? (
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
  const isDisabled = disabled || loading;
  const merged = isDisabled
    ? `${base} cursor-not-allowed disabled:opacity-50`
    : base;
  return (
    <button type={type} className={merged} disabled={isDisabled} {...rest}>
      {content}
    </button>
  );
}
