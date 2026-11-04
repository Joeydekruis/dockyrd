type EventFeedItemProps = {
  time: string;
  title: string;
  detail?: string;
  actor?: string;
};

export function EventFeedItem({
  time,
  title,
  detail,
  actor,
}: EventFeedItemProps) {
  return (
    <div
      className="relative border-l pb-6 pl-4 last:pb-0"
      style={{ borderColor: "var(--border)" }}
    >
      <span
        className="absolute -left-[5px] top-1.5 size-2 rounded-full"
        style={{ background: "var(--accent-blue)" }}
        aria-hidden
      />
      <p
        className="font-mono text-[11px] tabular-nums"
        style={{ color: "var(--muted-text)" }}
      >
        {time}
      </p>
      <p className="mt-1 text-sm font-medium">{title}</p>
      {detail ? (
        <p
          className="mt-0.5 text-[13px]"
          style={{ color: "var(--muted-text)" }}
        >
          {detail}
        </p>
      ) : null}
      {actor ? (
        <p
          className="mt-1 text-[13px]"
          style={{ color: "var(--secondary-text)" }}
        >
          {actor}
        </p>
      ) : null}
    </div>
  );
}
