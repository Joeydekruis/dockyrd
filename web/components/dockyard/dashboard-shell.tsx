"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { AppHeader } from "@/components/dockyard/app-header";
import {
  CommandPalette,
  useCommandPalette,
} from "@/components/dockyard/command-palette";
import { Sidebar } from "@/components/dockyard/sidebar";
import { TopBar } from "@/components/dockyard/top-bar";
import { UpdateAvailableNotifier } from "@/components/dockyard/update-available-modal";
import { getMe } from "@/lib/api";

type DashboardShellProps = {
  title: string;
  breadcrumb?: string;
  children: React.ReactNode;
};

export function DashboardShell({
  title,
  breadcrumb,
  children,
}: DashboardShellProps) {
  const router = useRouter();
  const [ready, setReady] = useState(false);
  const palette = useCommandPalette();
  const scrollRef = useRef<HTMLDivElement>(null);
  const [topBarVisible, setTopBarVisible] = useState(true);
  const [announcementOccupied, setAnnouncementOccupied] = useState(false);

  const onScroll = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    setTopBarVisible(el.scrollTop <= 8);
  }, []);

  useEffect(() => {
    let cancelled = false;
    getMe()
      .then((me) => {
        if (cancelled) return;
        if (
          me.role === "owner" &&
          !me.totp_enabled &&
          !window.location.pathname.startsWith("/settings/profile")
        ) {
          router.replace("/settings/profile");
          return;
        }
        setReady(true);
      })
      .catch(() => {
        if (!cancelled) router.replace("/sign-in");
      });
    return () => {
      cancelled = true;
    };
  }, [router]);

  if (!ready) {
    return (
      <div
        className="flex min-h-dvh items-center justify-center text-sm"
        style={{ color: "var(--muted-text)" }}
      >
        Loading…
      </div>
    );
  }

  return (
    <div className="min-h-dvh" style={{ background: "var(--page-bg)" }}>
      <TopBar
        visible={topBarVisible}
        onOccupiedChange={setAnnouncementOccupied}
        className="fixed top-0 right-0 left-0 z-50"
      />
      <Sidebar />
      <div
        className="flex min-h-dvh flex-col pl-[232px]"
        style={{
          paddingTop: announcementOccupied
            ? "var(--dockyard-announcement-height, 44px)"
            : undefined,
        }}
      >
        <div
          className="sticky z-30 shrink-0"
          style={{
            top: announcementOccupied
              ? "var(--dockyard-announcement-height, 44px)"
              : 0,
            background: "var(--page-bg)",
          }}
        >
          <AppHeader
            title={title}
            breadcrumb={breadcrumb}
            onOpenSearch={() => palette.setOpen(true)}
          />
        </div>
        <div
          ref={scrollRef}
          onScroll={onScroll}
          className="flex-1 overflow-y-auto px-8 py-6"
        >
          <UpdateAvailableNotifier />
          {children}
        </div>
      </div>
      <CommandPalette open={palette.open} onClose={palette.close} />
    </div>
  );
}
