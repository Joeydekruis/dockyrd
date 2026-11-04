import type { ComponentPropsWithoutRef } from "react";

type ButtonVariant = "primary" | "outline";

type ButtonProps = {
  variant?: ButtonVariant;
} & (
  | (ComponentPropsWithoutRef<"button"> & { href?: undefined })
  | (ComponentPropsWithoutRef<"a"> & { href: string })
);

export function Button({
  variant = "primary",
  className = "",
  ...props
}: ButtonProps) {
  const base =
    `${variant === "primary" ? "btn-primary" : "btn-outline"} ${className}`.trim();

  if ("href" in props && props.href) {
    const { href, ...rest } = props;
    return <a href={href} className={base} {...rest} />;
  }

  const {
    type = "button",
    disabled,
    ...rest
  } = props as ComponentPropsWithoutRef<"button">;
  const merged = disabled ? `${base} disabled:opacity-50` : base;
  return (
    <button type={type} className={merged} disabled={disabled} {...rest} />
  );
}
