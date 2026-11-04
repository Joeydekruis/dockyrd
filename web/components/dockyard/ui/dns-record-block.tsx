import { CodeInline } from "@/components/dockyard/ui/code-inline";

type DnsRecordBlockProps = {
  type: string;
  name: string;
  value: string;
  ttl?: string;
};

export function DnsRecordBlock({
  type,
  name,
  value,
  ttl,
}: DnsRecordBlockProps) {
  return (
    <div
      className="rounded-xl border p-4 font-mono text-[13px]"
      style={{ borderColor: "var(--border)", background: "var(--hover-row)" }}
    >
      <div className="flex flex-wrap gap-x-4 gap-y-1">
        <span>
          <span style={{ color: "var(--muted-text)" }}>Type </span>
          <CodeInline>{type}</CodeInline>
        </span>
        <span>
          <span style={{ color: "var(--muted-text)" }}>Name </span>
          <CodeInline>{name}</CodeInline>
        </span>
        {ttl ? (
          <span>
            <span style={{ color: "var(--muted-text)" }}>TTL </span>
            <CodeInline>{ttl}</CodeInline>
          </span>
        ) : null}
      </div>
      <p className="mt-2 break-all">
        <span style={{ color: "var(--muted-text)" }}>Value </span>
        {value}
      </p>
    </div>
  );
}
