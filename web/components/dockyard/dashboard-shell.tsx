"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { AppHeader } from "@/components/dockyard/app-header";
import {
  CommandPalette,
  useCommandPalette,
} from "@/components/dockyard/command-palette";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { Sidebar } from "@/components/dockyard/sidebar";
import { TopBar } from "@/components/dockyard/top-bar";
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
          <TopBar className="fixed top-0 right-0 left-0 z-50" />
          <Sidebar />
          <div
            className="flex min-h-dvh flex-col pl-[232px]"
            style={{ paddingTop: "var(--dockyard-announcement-height, 44px)" }}
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
            <div className="flex-1 overflow-y-auto px-8 py-6">
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
