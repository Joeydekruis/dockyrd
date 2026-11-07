"use client";

import { useEffect, useState } from "react";
import { DockyrdCubeIcon } from "@/components/dockyrd/dockyrd-cube-icon";

const LOADER_CUBE_SLOTS = ["a", "b", "c", "d", "e"] as const;

/** Avoid flashing the loader on sub‑200ms fetches; show only when wait is noticeable. */
const DEFAULT_SHOW_DELAY_MS = 250;

type DockyrdPageLoadingProps = {
  label?: string;
  /** Full viewport centering vs in-page placeholder. */
  layout?: "page" | "inline";
  className?: string;
  cubeSize?: number;
  /** Milliseconds before the cube animation appears; 0 = immediate. */
  showDelayMs?: number;
};

export function DockyrdPageLoading({
  label = "Loading…",
  layout = "inline",
  className = "",
  cubeSize = 36,
  showDelayMs = DEFAULT_SHOW_DELAY_MS,
}: DockyrdPageLoadingProps) {
  const [showAnimation, setShowAnimation] = useState(showDelayMs <= 0);

  useEffect(() => {
    if (showDelayMs <= 0) {
      setShowAnimation(true);
      return;
    }
    setShowAnimation(false);
    const id = window.setTimeout(() => setShowAnimation(true), showDelayMs);
    return () => window.clearTimeout(id);
  }, [showDelayMs]);

  if (!showAnimation) {
    if (layout === "page") {
      return (
        <div
          className="min-h-dvh"
          style={{ background: "var(--page-bg)" }}
          aria-hidden
        />
      );
    }
    return null;
  }

  const loader = (
    <>
      <div className="dockyrd-loader" aria-hidden>
        {LOADER_CUBE_SLOTS.map((slot) => (
          <DockyrdCubeIcon
            key={slot}
            size={cubeSize}
            className="dockyrd-loader-cube"
          />
        ))}
      </div>
      <span className="sr-only">{label}</span>
      {layout === "page" ? (
        <p className="mt-6 text-sm" style={{ color: "var(--muted-text)" }}>
          {label}
        </p>
      ) : null}
    </>
  );

  if (layout === "page") {
    return (
      <div
        className="flex min-h-dvh flex-col items-center justify-center"
        style={{ background: "var(--page-bg)" }}
        role="status"
        aria-live="polite"
      >
        {loader}
      </div>
    );
  }

  return (
    <div
      className={`flex flex-col items-center py-8 ${className}`.trim()}
      role="status"
      aria-live="polite"
    >
      {loader}
    </div>
  );
}
