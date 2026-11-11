"use client";

import { Link2 } from "lucide-react";
import { useState } from "react";
import { Chip } from "@/components/dockyard/ui/chip";
import { RevealableCopyField } from "@/components/dockyard/ui/revealable-copy-field";
import { parseEnvVarRef } from "@/lib/env-var-ref";

type EnvVarRowProps = {
  name: string;
  value: string;
  secret?: boolean;
  linkId?: string;
  /** When true (e.g. page-level “Reveal secrets”), secret values show unmasked. */
  secretsRevealed?: boolean;
};

export function EnvVarRow({
  name,
  value,
  secret = false,
  linkId,
  secretsRevealed = false,
}: EnvVarRowProps) {
  const [rowRevealed, setRowRevealed] = useState(false);
  const revealed = !secret || secretsRevealed || rowRevealed;
  const ref = parseEnvVarRef(value);
  const unresolved = value.includes("${{") && !ref;
  const canCopy = !secret || revealed;

  return (
    <div
      className="flex flex-wrap items-center gap-3 border-t py-3 text-[13px] first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <span
        className="min-w-[120px] font-mono font-semibold"
        style={{ color: "var(--ink-navy)" }}
      >
        {name}
      </span>
      <RevealableCopyField
        className="min-w-[200px] flex-1"
        displayValue={value}
        copyValue={value}
        copyLabel={`Copy ${name}`}
        copyMessage={`Copied ${name}`}
        canCopy={canCopy}
        revealable={secret}
        revealed={revealed}
        onToggleReveal={() => setRowRevealed((v) => !v)}
        revealLabel={`Show ${name}`}
        hideLabel={`Hide ${name}`}
      />
      {linkId ? <Chip>via link</Chip> : null}
      {ref ? (
        <Chip>
          <Link2 className="mr-1 inline size-3" aria-hidden />
          {ref.slug} · {ref.key}
        </Chip>
      ) : null}
      {unresolved ? <Chip>Not linked</Chip> : null}
    </div>
  );
}
