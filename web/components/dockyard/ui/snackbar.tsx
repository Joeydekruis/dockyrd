"use client";

import type { LucideIcon } from "lucide-react";
import {
  AlertCircle,
  CheckCircle2,
  Info,
  TriangleAlert,
  X,
} from "lucide-react";
import { useEffect, useState, type ReactNode } from "react";

export type SnackbarTone = "info" | "success" | "warning" | "error" | "warn";

type SnackbarProps = {
  /** Primary line (bold on toast). */
  title?: ReactNode;
  /** Secondary line under the title. */
  description?: ReactNode;
  /** Shorthand for `title` (backward compatible). */
  message?: ReactNode;
  action?: ReactNode;
  /** Uppercase action link (toast design); ignored when `action` is set. */
  actionLabel?: string;
  onAction?: () => void;
  onDismiss?: () => void;
  tone?: SnackbarTone;
  /** 0–100 timer/progress bar; omit to hide. */
  progress?: number;
  /** Floating toasts use shadow; set false only for legacy inline use. */
  elevated?: boolean;
};

type ToneConfig = {
  progress: string;
  Icon: LucideIcon;
  iconBg: string;
};

const toneConfig: Record<"info" | "success" | "warning" | "error", ToneConfig> =
  {
    info: {
      progress: "var(--accent-blue)",
      Icon: Info,
      iconBg: "var(--accent-blue)",
    },
    success: {
      progress: "var(--status-ok-dot)",
      Icon: CheckCircle2,
      iconBg: "var(--status-ok-dot)",
    },
    warning: {
      progress: "#c47f00",
      Icon: TriangleAlert,
      iconBg: "#c47f00",
    },
    error: {
      progress: "var(--danger)",
      Icon: AlertCircle,
      iconBg: "var(--danger)",
    },
  };

function normalizeTone(tone: SnackbarTone): keyof typeof toneConfig {
  return tone === "warn" ? "warning" : tone;
}

export function Snackbar({
  title,
  description,
  message,
  action,
  actionLabel,
  onAction,
  onDismiss,
  tone = "info",
  progress,
  elevated = true,
}: SnackbarProps) {
  const heading = title ?? message;
  const key = normalizeTone(tone);
  const t = toneConfig[key];
  const Icon = t.Icon;
  const showProgress = progress !== undefined;

  const actionNode =
    action ??
    (actionLabel && onAction ? (
      <button
        type="button"
        className="snackbar-toast-action"
        onClick={onAction}
      >
        {actionLabel}
      </button>
    ) : null);

  return (
    <div
      role="status"
      className={`snackbar-toast${elevated ? " snackbar-toast-elevated" : ""}${showProgress ? " snackbar-toast-with-progress" : ""}`}
    >
      <div className="snackbar-toast-body">
        <span
          className="snackbar-toast-icon"
          style={{ background: t.iconBg }}
          aria-hidden
        >
          <Icon strokeWidth={2.2} className="size-3.5 text-white" />
        </span>
        <div className="snackbar-toast-content min-w-0 flex-1">
          {heading ? (
            <p className="snackbar-toast-title">{heading}</p>
          ) : null}
          {description ? (
            <p className="snackbar-toast-description">{description}</p>
          ) : null}
          {actionNode ? (
            <div className="snackbar-toast-actions">{actionNode}</div>
          ) : null}
        </div>
        {onDismiss ? (
          <button
            type="button"
            className="snackbar-toast-dismiss"
            aria-label="Dismiss"
            onClick={onDismiss}
          >
            <X strokeWidth={2} className="size-4" aria-hidden />
          </button>
        ) : null}
      </div>
      {showProgress ? (
        <div
          className="snackbar-toast-progress-track"
          role="progressbar"
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={Math.round(progress)}
        >
          <span
            className="snackbar-toast-progress-fill"
            style={{
              width: `${Math.max(0, Math.min(100, progress))}%`,
              background: t.progress,
            }}
          />
        </div>
      ) : null}
    </div>
  );
}

type SnackbarViewportProps = {
  open: boolean;
  title?: ReactNode;
  description?: ReactNode;
  message?: ReactNode;
  tone?: SnackbarTone;
  onDismiss: () => void;
  durationMs?: number;
};

/** Fixed bottom toast (Design/Components/Snackbar). */
export function SnackbarViewport({
  open,
  title,
  description,
  message,
  tone = "info",
  onDismiss,
  durationMs = 6000,
}: SnackbarViewportProps) {
  const [progress, setProgress] = useState(100);

  useEffect(() => {
    if (!open) return;
    setProgress(100);
    const start = performance.now();
    let frame = 0;
    const tick = (now: number) => {
      const elapsed = now - start;
      const remaining = Math.max(0, 100 - (elapsed / durationMs) * 100);
      setProgress(remaining);
      if (remaining > 0) frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [open, durationMs]);

  if (!open) return null;
  return (
    <div className="pointer-events-none fixed inset-x-0 bottom-6 z-50 flex justify-center px-4">
      <div className="pointer-events-auto w-full max-w-md">
        <Snackbar
          title={title}
          description={description}
          message={message}
          tone={tone}
          progress={progress}
          onDismiss={onDismiss}
        />
      </div>
    </div>
  );
}
