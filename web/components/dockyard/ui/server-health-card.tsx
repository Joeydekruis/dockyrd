type ServerHealthCardProps = {
  healthy?: boolean;
  version: string;
  statusLine?: string;
};

export function ServerHealthCard({
  healthy = true,
  version,
  statusLine = "Up to date",
}: ServerHealthCardProps) {
  return (
    <div
      className="rounded-[20px] border p-4 text-xs"
      style={{ borderColor: "var(--border)", background: "var(--surface)" }}
    >
      <div className="flex items-center gap-2">
        <span
          className="inline-block size-2 rounded-full"
          style={{
            background: healthy ? "var(--status-ok-dot)" : "var(--danger)",
          }}
          aria-hidden
        />
        <span className="font-medium">
          {healthy ? "Server healthy" : "Needs attention"}
        </span>
      </div>
      <p
        className="mt-1 font-mono text-[11px]"
        style={{ color: "var(--muted-text)" }}
      >
        {version}
      </p>
      <p className="mt-1" style={{ color: "var(--muted-text)" }}>
        {statusLine}
      </p>
    </div>
  );
}
