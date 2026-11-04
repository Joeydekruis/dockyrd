type CommitChipProps = {
  sha: string;
  message?: string;
  timeAgo?: string;
};

export function CommitChip({ sha, message, timeAgo }: CommitChipProps) {
  return (
    <span className="inline-flex max-w-[min(100%,280px)] flex-col items-end gap-0.5 text-right text-[13px] sm:max-w-[320px]">
      <span className="inline-flex items-center gap-2">
        <code
          className="rounded-md px-2 py-0.5 font-mono text-[11px] uppercase"
          style={{ background: "var(--hover-row)", color: "var(--ink-navy)" }}
        >
          {sha.slice(0, 7)}
        </code>
        {message ? (
          <span
            className="hidden truncate sm:inline"
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
