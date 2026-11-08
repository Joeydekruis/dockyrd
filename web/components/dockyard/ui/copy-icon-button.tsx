"use client";

import { Check, Copy } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import { useCopyToClipboard } from "@/lib/snackbar-context";

type CopyIconButtonProps = {
  value: string;
  label: string;
  message?: string;
  disabled?: boolean;
};

const COPIED_MS = 2000;

export function CopyIconButton({
  value,
  label,
  message,
  disabled = false,
}: CopyIconButtonProps) {
  const copyToClipboard = useCopyToClipboard();
  const [copied, setCopied] = useState(false);
  const resetId = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (resetId.current) clearTimeout(resetId.current);
    };
  }, []);

  const onCopy = useCallback(() => {
    if (disabled) return;
    void copyToClipboard(value, message).then(() => {
      setCopied(true);
      if (resetId.current) clearTimeout(resetId.current);
      resetId.current = setTimeout(() => setCopied(false), COPIED_MS);
    });
  }, [copyToClipboard, disabled, message, value]);

  return (
    <span className="inline-flex items-center">
      <span
        className="overflow-hidden transition-[width,opacity] duration-200 ease-out"
        style={{
          width: copied ? 20 : 0,
          opacity: copied ? 1 : 0,
        }}
        aria-hidden
      >
        <Check
          className="size-4 shrink-0"
          strokeWidth={2}
          style={{ color: "#2d8a4e" }}
        />
      </span>
      <IconButton
        icon={Copy}
        label={label}
        disabled={disabled}
        iconClassName="size-4"
        onClick={onCopy}
      />
    </span>
  );
}
