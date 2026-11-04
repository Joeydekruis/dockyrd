import type { ReactNode } from "react";

type KbdProps = {
  children: ReactNode;
};

export function Kbd({ children }: KbdProps) {
  return (
    <kbd
      className="inline-flex min-h-6 items-center rounded-md border px-1.5 font-mono text-[11px] uppercase tracking-wide"
      style={{
        borderColor: "var(--input-border)",
        background: "var(--page-bg)",
        color: "var(--secondary-text)",
      }}
    >
      {children}
    </kbd>
  );
}
