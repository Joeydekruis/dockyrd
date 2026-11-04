import type { ReactNode } from "react";

type CodeInlineProps = {
  children: ReactNode;
};

export function CodeInline({ children }: CodeInlineProps) {
  return (
    <code
      className="rounded-md px-1.5 py-0.5 font-mono text-[13px]"
      style={{ background: "var(--hover-row)", color: "var(--ink-navy)" }}
    >
      {children}
    </code>
  );
}
