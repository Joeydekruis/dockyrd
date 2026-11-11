"use client";

import { Eye, EyeOff } from "lucide-react";
import { CopyIconButton } from "@/components/dockyard/ui/copy-icon-button";
import { IconButton } from "@/components/dockyard/ui/icon-button";

const DEFAULT_MASK = "••••••••";

type RevealableCopyFieldProps = {
  displayValue: string;
  copyValue: string;
  copyLabel: string;
  copyMessage?: string;
  canCopy?: boolean;
  /** When true, shows the reveal segment after a vertical divider. */
  revealable?: boolean;
  revealed?: boolean;
  onToggleReveal?: () => void;
  revealLabel?: string;
  hideLabel?: string;
  className?: string;
};

export function RevealableCopyField({
  displayValue,
  copyValue,
  copyLabel,
  copyMessage,
  canCopy = true,
  revealable = false,
  revealed = false,
  onToggleReveal,
  revealLabel = "Show value",
  hideLabel = "Hide value",
  className = "",
}: RevealableCopyFieldProps) {
  const showMask = revealable && !revealed;
  const text = showMask ? DEFAULT_MASK : displayValue;

  return (
    <span
      className={`dockyrd-revealable-copy-field inline-flex min-w-0 max-w-full items-stretch overflow-hidden rounded-full border font-mono text-[13px] ${className}`.trim()}
      style={{
        borderColor: "var(--border)",
        background: "var(--hover-row)",
        color: "var(--ink-navy)",
      }}
    >
      <span
        className={`flex min-w-0 flex-1 items-center py-1 pl-3 pr-1 ${showMask ? "tracking-widest" : ""}`}
        title={showMask ? undefined : displayValue}
      >
        <span className="min-w-0 truncate">{text}</span>
      </span>
      <span className="flex shrink-0 items-center pr-0.5">
        <CopyIconButton
          value={copyValue}
          label={copyLabel}
          message={copyMessage}
          disabled={!canCopy}
        />
      </span>
      {revealable ? (
        <>
          <span
            className="w-px shrink-0 self-stretch"
            style={{ background: "var(--border)" }}
            aria-hidden
          />
          <span className="flex shrink-0 items-center pl-0.5 pr-0.5">
            <IconButton
              icon={revealed ? EyeOff : Eye}
              iconClassName="size-4"
              label={revealed ? hideLabel : revealLabel}
              onClick={onToggleReveal}
            />
          </span>
        </>
      ) : null}
    </span>
  );
}
