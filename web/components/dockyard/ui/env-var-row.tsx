"use client";

import { Copy, Eye, EyeOff, Link2 } from "lucide-react";
import { useState } from "react";
import { Chip } from "@/components/dockyard/ui/chip";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import { useCopyToClipboard } from "@/lib/snackbar-context";

type EnvVarRowProps = {
  name: string;
  value: string;
  secret?: boolean;
  linkId?: string;
};

const refPattern = /\$\{\{([a-z0-9][a-z0-9-]*)\.([A-Z][A-Z0-9_]*)\}\}/;

function refHint(value: string): { slug: string; key: string } | null {
  const m = refPattern.exec(value);
  if (!m) return null;
  return { slug: m[1], key: m[2] };
}

export function EnvVarRow({
  name,
  value,
  secret = false,
  linkId,
}: EnvVarRowProps) {
  const [revealed, setRevealed] = useState(false);
  const copyToClipboard = useCopyToClipboard();
  const display = secret && !revealed ? "••••••••" : value;
  const ref = refHint(value);
  const unresolved = value.includes("${{") && !ref;
  const canCopy = !secret || revealed;
  return (
    <div
      className="flex flex-wrap items-center gap-3 border-t py-3 font-mono text-[13px] first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <span
        className="min-w-[120px] font-semibold"
        style={{ color: "var(--ink-navy)" }}
      >
        {name}
      </span>
      <span
        className="min-w-0 flex-1 truncate"
        style={{ color: "var(--muted-text)" }}
      >
        {display}
      </span>
      {linkId ? <Chip>via link</Chip> : null}
      {ref ? (
        <Chip>
          <Link2 className="mr-1 inline size-3" aria-hidden />
          {ref.slug} · {ref.key}
        </Chip>
      ) : null}
      {unresolved ? <Chip>Not linked</Chip> : null}
      {secret ? (
        <IconButton
          icon={revealed ? EyeOff : Eye}
          label={revealed ? "Hide value" : "Show value"}
          onClick={() => setRevealed((v) => !v)}
        />
      ) : null}
      <IconButton
        icon={Copy}
        label={`Copy ${name}`}
        disabled={!canCopy}
        onClick={() => {
          if (!canCopy) return;
          void copyToClipboard(value, `Copied ${name}`);
        }}
      />
    </div>
  );
}
