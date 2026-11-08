"use client";

import { NotificationsBell } from "@/components/dockyard/notifications-bell";
import {
  type BreadcrumbItem,
  Breadcrumbs,
} from "@/components/dockyard/ui/breadcrumbs";
import { Button } from "@/components/dockyard/ui/button";
import { SearchTrigger } from "@/components/dockyard/ui/search-trigger";
import { UserMenu } from "@/components/dockyard/user-menu";
import type { MeResponse } from "@/lib/api";
import { PUBLIC_DOCS_INTRO_URL } from "@/lib/public-links";

type AppHeaderProps = {
  title: string;
  breadcrumb?: string;
  breadcrumbItems?: BreadcrumbItem[];
  user: MeResponse;
  onOpenSearch: () => void;
};

export function AppHeader({
  title,
  breadcrumb = "Dashboard",
  breadcrumbItems,
  user,
  onOpenSearch,
}: AppHeaderProps) {
  const items =
    breadcrumbItems ??
    ([{ label: breadcrumb }, { label: title }] satisfies BreadcrumbItem[]);

  return (
    <header
      className="flex min-h-16 items-center justify-between gap-4 border-b px-8 py-3"
      style={{ borderColor: "var(--border)", background: "var(--page-bg)" }}
    >
      <Breadcrumbs items={items} />
      <div className="flex items-center gap-3">
        <SearchTrigger onClick={onOpenSearch} />
        <Button href={PUBLIC_DOCS_INTRO_URL} variant="outline">
          Docs
        </Button>
        <NotificationsBell />
        <UserMenu user={user} />
      </div>
    </header>
  );
}
