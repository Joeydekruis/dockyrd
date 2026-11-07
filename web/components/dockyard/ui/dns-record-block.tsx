"use client";

import { Copy } from "lucide-react";
import { CodeInline } from "@/components/dockyard/ui/code-inline";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import { useCopyToClipboard } from "@/lib/snackbar-context";

type DnsRecordBlockProps = {
  type: string;
  name: string;
  nameAlt?: string;
  value: string;
  ttl?: string;
  copyable?: boolean;
};

function CopyableValue({ value }: { value: string }) {
  const copyToClipboard = useCopyToClipboard();
  return (
    <span className="inline-flex items-start gap-1">
      <span className="break-all">{value}</span>
      <IconButton
        icon={Copy}
        label="Copy value"
        className="shrink-0"
        onClick={() => {
          void copyToClipboard(value, "Copied DNS value");
        }}
      />
    </span>
  );
}

function CopyableName({ name, nameAlt }: { name: string; nameAlt?: string }) {
  const copyToClipboard = useCopyToClipboard();
  const names = nameAlt && nameAlt !== name ? [name, nameAlt] : [name];
  return (
    <span className="inline-flex flex-wrap items-center gap-2">
      {names.map((n) => (
        <span key={n} className="inline-flex items-center gap-0.5">
          <CodeInline>{n}</CodeInline>
          <IconButton
            icon={Copy}
            label={`Copy ${n}`}
            onClick={() => {
              void copyToClipboard(n, "Copied DNS name");
            }}
          />
        </span>
      ))}
    </span>
  );
}

export function DnsRecordBlock({
  type,
  name,
  nameAlt,
  value,
  ttl,
  copyable,
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
        <span className="min-w-0">
          <span style={{ color: "var(--muted-text)" }}>Name </span>
          {copyable ? (
            <CopyableName name={name} nameAlt={nameAlt} />
          ) : (
            <CodeInline>{name}</CodeInline>
          )}
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
        {copyable ? <CopyableValue value={value} /> : value}
      </p>
    </div>
  );
}
