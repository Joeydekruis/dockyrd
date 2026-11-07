"use client";

import { AnnouncementBanner } from "@/components/dockyrd/ui/announcement-banner";

export const BANNER_HEIGHT_PX = 44;

type TopBarProps = {
  className?: string;
  visible?: boolean;
  /** When false, the bar is collapsed (e.g. user scrolled main content). */
  expanded?: boolean;
  onDismiss?: () => void;
};

/** Fixed announcement strip spanning the full app width (above sidebar and header). */
export function TopBar({
  className = "",
  visible = true,
  expanded = true,
  onDismiss,
}: TopBarProps) {
  if (!visible) return null;

  return (
    <div
      className={className}
      style={{
        height: expanded ? BANNER_HEIGHT_PX : 0,
        minHeight: expanded ? BANNER_HEIGHT_PX : 0,
        overflow: "hidden",
        transition: "height 0.2s ease, min-height 0.2s ease",
      }}
    >
      <AnnouncementBanner onDismiss={onDismiss} />
    </div>
  );
}
