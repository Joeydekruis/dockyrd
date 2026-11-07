type CodeBlockProps = {
  code: string;
  label?: string;
};

export function CodeBlock({ code, label }: CodeBlockProps) {
  return (
    <div className="flex flex-col gap-2">
      {label ? (
        <span
          className="font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          {label}
        </span>
      ) : null}
      <pre
        className="overflow-x-auto rounded-md border p-3 font-mono text-xs leading-relaxed"
        style={{
          borderColor: "var(--border)",
          background: "var(--surface-raised)",
          color: "var(--text)",
        }}
      >
        {code}
      </pre>
    </div>
  );
}
