import type { ReactNode } from "react";

export type TypographyRole = "h1" | "h2" | "body" | "muted" | "mono";

const roleClass: Record<TypographyRole, string> = {
  h1: "text-[32px] font-semibold leading-tight tracking-tight",
  h2: "text-xl font-semibold",
  body: "text-sm",
  muted: "text-sm",
  mono: "font-mono text-[13px]",
};

type TypographySampleProps = {
  variant: TypographyRole;
  children: ReactNode;
};

export function TypographySample({ variant, children }: TypographySampleProps) {
  const color =
    variant === "muted"
      ? "var(--muted-text)"
      : variant === "mono"
        ? "var(--ink-navy)"
        : undefined;
  return (
    <p className={roleClass[variant]} style={color ? { color } : undefined}>
      {children}
    </p>
  );
}
