"use client";

import { usePathname, useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { AppHeader } from "@/components/dockyard/app-header";
import {
  CommandPalette,
  useCommandPalette,
} from "@/components/dockyard/command-palette";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { Sidebar } from "@/components/dockyard/sidebar";
import { BANNER_HEIGHT_PX, TopBar } from "@/components/dockyard/top-bar";
import { isAnnouncementDismissed } from "@/components/dockyard/ui/announcement-banner";
import { UpdateAvailableNotifier } from "@/components/dockyard/update-available-modal";
import { getMe, type MeResponse } from "@/lib/api";
import { isAccountSettingsPath, isSettingsPath } from "@/lib/nav";
import { SnackbarProvider } from "@/lib/snackbar-context";
import { UserLocaleProvider } from "@/lib/user-locale-context";

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
  const pathname = usePathname();
  const [ready, setReady] = useState(false);
  const [user, setUser] = useState<MeResponse | null>(null);
  const palette = useCommandPalette();
  const mainScrollRef = useRef<HTMLDivElement>(null);
  const [bannerExpanded, setBannerExpanded] = useState(true);
  const [bannerVisible, setBannerVisible] = useState(true);

  const syncBannerHeight = useCallback(
    (visible: boolean, expanded: boolean) => {
      if (typeof document === "undefined") return;
      const px =
        visible && expanded && !isAnnouncementDismissed()
          ? `${BANNER_HEIGHT_PX}px`
          : "0px";
      document.documentElement.style.setProperty(
        "--dockyard-announcement-height",
        px,
      );
    },
    [],
  );

  useEffect(() => {
    syncBannerHeight(bannerVisible, bannerExpanded);
  }, [bannerVisible, bannerExpanded, syncBannerHeight]);

  useEffect(() => {
    setBannerVisible(!isAnnouncementDismissed());
  }, []);

  useEffect(() => {
    let cancelled = false;
    getMe()
      .then((me) => {
        if (cancelled) return;
        if (
          me.role === "owner" &&
          !me.totp_enabled &&
          !isAccountSettingsPath(window.location.pathname)
        ) {
          router.replace("/settings/account/password");
          return;
        }
        setUser(me);
        setReady(true);
      })
      .catch(() => {
        if (!cancelled) router.replace("/sign-in");
      });
    return () => {
      cancelled = true;
    };
  }, [router]);

  if (!ready || !user) {
    return <DockyrdPageLoading layout="page" />;
  }

  return (
    <SnackbarProvider>
      <UserLocaleProvider>
        <div className="min-h-dvh" style={{ background: "var(--page-bg)" }}>
          <TopBar
            className="fixed top-0 right-0 left-0 z-50"
            visible={bannerVisible}
            expanded={bannerExpanded}
            onDismiss={() => {
              setBannerVisible(false);
              syncBannerHeight(false, false);
            }}
          />
          <Sidebar />
          <div
            className="flex min-h-dvh flex-col pl-[232px]"
            style={{
              paddingTop: bannerVisible
                ? "var(--dockyard-announcement-height, 44px)"
                : 0,
            }}
          >
            <div
              className="sticky z-20 shrink-0"
              style={{
                top: "var(--dockyard-announcement-height, 44px)",
                background: "var(--page-bg)",
              }}
            >
              <AppHeader
                title={title}
                breadcrumb={breadcrumb}
                user={user}
                onOpenSearch={() => palette.setOpen(true)}
              />
            </div>
            <div
              ref={mainScrollRef}
              className="flex-1 overflow-y-auto px-8 py-6"
              onScroll={() => {
                const el = mainScrollRef.current;
                if (!el || !bannerVisible) return;
                setBannerExpanded(el.scrollTop < 8);
              }}
            >
              <div className="max-w-6xl mx-auto">
                {isSettingsPath(pathname) ? (
                  <div className="settings-page-body">
                    <UpdateAvailableNotifier />
                  </div>
                ) : (
                  <UpdateAvailableNotifier />
                )}
                {children}
              </div>
            </div>
          </div>
          <CommandPalette open={palette.open} onClose={palette.close} />
        </div>
      </UserLocaleProvider>
    </SnackbarProvider>
  );
}
