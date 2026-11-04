"use client";

import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import { IconButton } from "@/components/dockyard/ui/icon-button";

type EnvVarRowProps = {
  name: string;
  value: string;
  secret?: boolean;
};

export function EnvVarRow({ name, value, secret = false }: EnvVarRowProps) {
  const [revealed, setRevealed] = useState(false);
  const display = secret && !revealed ? "••••••••" : value;
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
      {secret ? (
        <IconButton
          icon={revealed ? EyeOff : Eye}
          label={revealed ? "Hide value" : "Show value"}
          onClick={() => setRevealed((v) => !v)}
        />
      ) : null}
    </div>
  );
}
