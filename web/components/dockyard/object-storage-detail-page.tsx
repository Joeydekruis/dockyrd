"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { getStorageStatus } from "@/lib/api";
import { useRouteParam } from "@/lib/use-route-param";

export function ObjectStorageDetailPage() {
  const slug = useRouteParam("slug");
  const [name, setName] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!slug) return;
    const storage = await getStorageStatus();
    const bucket = storage.buckets?.find((b) => b.slug === slug);
    setName(bucket?.name ?? null);
  }, [slug]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <DashboardShell title={name ?? "Bucket"} breadcrumb="Object storage">
      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
        Bucket detail for <span className="font-mono">{slug || "—"}</span>. S3
        credentials and policies will live here — see{" "}
        <Link href="/object-storage" className="underline">
          Object storage
        </Link>
        .
      </p>
    </DashboardShell>
  );
}
