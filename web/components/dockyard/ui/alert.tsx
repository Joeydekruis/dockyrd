import type { LucideIcon } from "lucide-react";
import {
  AlertCircle,
  CheckCircle2,
  Info,
  TriangleAlert,
  X,
} from "lucide-react";
import type { ReactNode } from "react";

export type AlertTone = "info" | "success" | "warning" | "error";

const toneStyles: Record<
  AlertTone,
  { bg: string; fg: string; border: string; Icon: LucideIcon; iconBg: string }
> = {
  info: {
    bg: "#e3f0ff",
    fg: "#0b3d91",
    border: "#b8d4f5",
    Icon: Info,
    iconBg: "#0b57c9",
  },
  success: {
    bg: "#e6f4ec",
    fg: "#176b3d",
    border: "#b8dfc8",
    Icon: CheckCircle2,
    iconBg: "#1a7f4b",
  },
  warning: {
    bg: "#fbf0dc",
    fg: "#8a5200",
    border: "#ecd9a8",
    Icon: TriangleAlert,
    iconBg: "#c47f00",
  },
  error: {
    bg: "#fbe8e6",
    fg: "#b3261e",
    border: "#f0c4c0",
    Icon: AlertCircle,
    iconBg: "#b3261e",
  },
};

type AlertProps = {
  tone?: AlertTone;
  title?: ReactNode;
  description?: ReactNode;
  /** Shorthand for `title`. */
  message?: ReactNode;
  action?: ReactNode;
  onDismiss?: () => void;
};

export function Alert({
  tone = "info",
  title,
  description,
  message,
  action,
  onDismiss,
}: AlertProps) {
  const t = toneStyles[tone];
  const Icon = t.Icon;
  const heading = title ?? message;

  return (
    <div
      role="status"
      className="message-banner"
      style={{ background: t.bg, borderColor: t.border, color: t.fg }}
    >
      <span
        className="message-banner-icon"
        style={{ background: t.iconBg }}
        aria-hidden
      >
        <Icon strokeWidth={2.2} className="size-3.5 text-white" />
      </span>
      <div className="message-banner-content min-w-0 flex-1">
        {heading ? <p className="message-banner-title">{heading}</p> : null}
        {description ? (
          <p className="message-banner-description">{description}</p>
        ) : null}
        {action ? <div className="mt-2">{action}</div> : null}
      </div>
      {onDismiss ? (
        <button
          type="button"
          className="message-banner-dismiss"
          aria-label="Dismiss"
          onClick={onDismiss}
          style={{ color: t.fg }}
        >
          <X strokeWidth={2} className="size-4" aria-hidden />
        </button>
      ) : null}
    </div>
  );
}
