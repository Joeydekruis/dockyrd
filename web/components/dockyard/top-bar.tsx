"use client";

import { useEffect, useState } from "react";
import { AnnouncementBanner } from "@/components/dockyard/ui/announcement-banner";

const DISMISS_KEY = "dockyrd.announcement.dismissed";
const BANNER_HEIGHT_PX = 44;

type TopBarProps = {
  visible: boolean;
  /** Called when the banner is shown or hidden (dismiss, scroll, or hydration). */
  onOccupiedChange?: (occupied: boolean) => void;
  className?: string;
};

export function TopBar({
  visible,
  onOccupiedChange,
  className = "",
}: TopBarProps) {
  const [dismissed, setDismissed] = useState<boolean | null>(null);

  useEffect(() => {
    setDismissed(window.localStorage.getItem(DISMISS_KEY) === "1");
  }, []);

  const shown = dismissed !== null && !dismissed && visible;

  useEffect(() => {
    onOccupiedChange?.(shown);
  }, [shown, onOccupiedChange]);

  if (!shown) return null;

  return (
    <div className={className} style={{ minHeight: BANNER_HEIGHT_PX }}>
      <AnnouncementBanner
        onDismiss={() => {
          window.localStorage.setItem(DISMISS_KEY, "1");
          setDismissed(true);
        }}
      />
    </div>
  );
}

export { BANNER_HEIGHT_PX };
