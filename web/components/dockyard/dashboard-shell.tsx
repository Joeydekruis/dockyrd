"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getMe } from "@/lib/api";
import { Sidebar } from "@/components/dockyard/sidebar";
import { AppHeader } from "@/components/dockyard/app-header";
import {
  CommandPalette,
  useCommandPalette,
} from "@/components/dockyard/command-palette";

type DashboardShellProps = {
  title: string;
  breadcrumb?: string;
  children: React.ReactNode;
};

export function DashboardShell({ title, breadcrumb, children }: DashboardShellProps) {
  const router = useRouter();
  const [ready, setReady] = useState(false);
  const palette = useCommandPalette();

  useEffect(() => {
    let cancelled = false;
    getMe()
      .then(() => {
        if (!cancelled) setReady(true);
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
      <div className="flex min-h-dvh items-center justify-center text-sm" style={{ color: "var(--muted-text)" }}>
        Loading…
      </div>
    );
  }

  return (
    <div className="flex min-h-dvh">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <AppHeader title={title} breadcrumb={breadcrumb} onOpenSearch={() => palette.setOpen(true)} />
        <div className="flex-1 px-8 py-6">{children}</div>
      </div>
      <CommandPalette open={palette.open} onClose={palette.close} />
    </div>
  );
}
