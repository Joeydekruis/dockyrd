"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  type Deployment,
  listDeployments,
  listProjects,
  listServices,
  type Project,
  type Service,
} from "@/lib/api";
import { useRouteParam } from "@/lib/use-route-param";

type StaticSiteRow = {
  project: Project;
  service: Service;
  deployments: Deployment[];
};

export function StaticSiteDetailPage() {
  const slug = useRouteParam("slug");
  const [row, setRow] = useState<StaticSiteRow | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    if (!slug) {
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const { projects } = await listProjects();
      for (const project of projects) {
        const { services } = await listServices(project.id, "static_site");
        const service = services.find((s) => s.slug === slug);
        if (!service) {
          continue;
        }
        const { deployments } = await listDeployments(service.id);
        setRow({ project, service, deployments });
        setError(null);
        setLoading(false);
        return;
      }
      setRow(null);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load static site");
    } finally {
      setLoading(false);
    }
  }, [slug]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const title = useMemo(
    () => row?.service.name ?? "Static site",
    [row?.service.name],
  );

  return (
    <DashboardShell title={title} breadcrumb="Static sites">
      {loading ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : error ? (
        <p className="text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : !row ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Static site not found.
        </p>
      ) : (
        <section
          className="rounded-[20px] border px-6 py-5"
          style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        >
          <h2 className="text-lg font-semibold tracking-tight">
            {row.service.name}
          </h2>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            Project{" "}
            <Link
              href={`/projects/${row.project.slug}`}
              className="font-medium hover:underline"
              style={{ color: "var(--link-blue)" }}
            >
              {row.project.name}
            </Link>
          </p>
          <p className="mt-4 text-sm">
            Status:{" "}
            <span className="font-medium">{row.service.desired_state}</span>
          </p>
          {row.deployments[0]?.commit_sha ? (
            <p
              className="mt-2 text-sm font-mono"
              style={{ color: "var(--muted-text)" }}
            >
              Latest deploy {row.deployments[0].commit_sha.slice(0, 7)}
            </p>
          ) : null}
        </section>
      )}
    </DashboardShell>
  );
}
