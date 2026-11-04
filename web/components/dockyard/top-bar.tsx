"use client";

import { useEffect, useState } from "react";
import { AnnouncementBanner } from "@/components/dockyard/ui/announcement-banner";

const DISMISS_KEY = "dockyrd.announcement.dismissed";

type TopBarProps = {
  visible: boolean;
};

export function TopBar({ visible }: TopBarProps) {
  const [dismissed, setDismissed] = useState<boolean | null>(null);

  useEffect(() => {
    setDismissed(window.localStorage.getItem(DISMISS_KEY) === "1");
  }, []);

  function dismiss() {
    window.localStorage.setItem(DISMISS_KEY, "1");
    setDismissed(true);
  }

  if (dismissed === null || dismissed || !visible) return null;

  return <AnnouncementBanner onDismiss={dismiss} />;
}
