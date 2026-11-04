type CommitChipProps = {
  sha: string;
  message?: string;
};

export function CommitChip({ sha, message }: CommitChipProps) {
  return (
    <span className="inline-flex max-w-full items-center gap-2 text-right text-[13px]">
      <code
        className="rounded-md px-2 py-0.5 font-mono text-[11px] uppercase"
        style={{ background: "var(--hover-row)", color: "var(--ink-navy)" }}
      >
        {sha.slice(0, 7)}
      </code>
      {message ? (
        <span
          className="truncate"
          style={{ color: "var(--muted-text)" }}
          title={message}
        >
          {message}
        </span>
      ) : null}
    </span>
  );
}
