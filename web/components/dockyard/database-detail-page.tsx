"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { listDatabases } from "@/lib/api";
import { useRouteParam } from "@/lib/use-route-param";

export function DatabaseDetailPage() {
  const slug = useRouteParam("slug");
  const [name, setName] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!slug) return;
    const { databases } = await listDatabases();
    const row = databases.find((d) => d.slug === slug);
    setName(row?.name ?? null);
  }, [slug]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <DashboardShell title={name ?? "Database"} breadcrumb="Databases">
      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
        Database detail for <span className="font-mono">{slug || "—"}</span>.
        Connection strings and backups will live here — see{" "}
        <Link href="/databases" className="underline">
          Databases
        </Link>{" "}
        for the list.
      </p>
    </DashboardShell>
  );
}
