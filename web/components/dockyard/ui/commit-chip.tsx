type CommitChipProps = {
  sha: string;
  message?: string;
  timeAgo?: string;
};

export function CommitChip({ sha, message, timeAgo }: CommitChipProps) {
  return (
    <span className="inline-flex w-full min-w-0 max-w-full flex-col items-end gap-0.5 text-right text-[13px]">
      <span className="inline-flex w-full min-w-0 max-w-full items-center justify-end gap-2">
        <code
          className="shrink-0 rounded-md px-2 py-0.5 font-mono text-[11px] uppercase"
          style={{ background: "var(--hover-row)", color: "var(--ink-navy)" }}
        >
          {sha.slice(0, 7)}
        </code>
        {message ? (
          <span
            className="min-w-0 truncate"
            style={{ color: "var(--muted-text)" }}
            title={message}
          >
            {message}
          </span>
        ) : null}
      </span>
      {timeAgo ? (
        <span className="text-[13px]" style={{ color: "var(--muted-text)" }}>
          {timeAgo}
        </span>
      ) : null}
    </span>
  );
}
