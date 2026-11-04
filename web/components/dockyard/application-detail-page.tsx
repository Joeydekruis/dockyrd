"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { listApplications } from "@/lib/api";
import { useRouteParam } from "@/lib/use-route-param";

export function ApplicationDetailPage() {
  const slug = useRouteParam("slug");
  const [name, setName] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!slug) return;
    const { applications } = await listApplications();
    const app = applications.find((a) => a.slug === slug);
    setName(app?.name ?? null);
  }, [slug]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <DashboardShell title={name ?? "Application"} breadcrumb="Applications">
      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
        Application detail for <span className="font-mono">{slug || "—"}</span>.
        Full deploy and logs views are coming next — open from{" "}
        <Link href="/applications" className="underline">
          Applications
        </Link>{" "}
        for now.
      </p>
    </DashboardShell>
  );
}
