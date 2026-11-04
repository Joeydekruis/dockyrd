"use client";

import { Breadcrumbs } from "@/components/dockyard/ui/breadcrumbs";
import { Button } from "@/components/dockyard/ui/button";
import { SearchTrigger } from "@/components/dockyard/ui/search-trigger";

type AppHeaderProps = {
  title: string;
  breadcrumb?: string;
  onOpenSearch: () => void;
};

export function AppHeader({
  title,
  breadcrumb = "Dashboard",
  onOpenSearch,
}: AppHeaderProps) {
  return (
    <header
      className="flex min-h-16 items-center justify-between gap-4 border-b px-8 py-3"
      style={{ borderColor: "var(--border)", background: "var(--page-bg)" }}
    >
      <Breadcrumbs items={[{ label: breadcrumb }, { label: title }]} />
      <div className="flex items-center gap-3">
        <SearchTrigger onClick={onOpenSearch} />
        <Button href="https://github.com/Joeydekruis/dockyrd" variant="outline">
          Docs
        </Button>
      </div>
    </header>
  );
}
