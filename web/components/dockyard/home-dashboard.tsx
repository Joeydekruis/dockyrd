"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { listApplications, listDatabases, listProjects } from "@/lib/api";

export function HomeDashboard() {
  const [projects, setProjects] = useState(0);
  const [apps, setApps] = useState(0);
  const [dbs, setDbs] = useState(0);

  useEffect(() => {
    Promise.all([listProjects(), listApplications(), listDatabases()])
      .then(([p, a, d]) => {
        setProjects(p.projects.length);
        setApps(a.applications.length);
        setDbs(d.databases.length);
      })
      .catch(() => {});
  }, []);

  return (
    <DashboardShell title="Dashboard" breadcrumb="Home">
      <div className="mb-8">
        <h1 className="text-4xl font-semibold tracking-tight">Dashboard</h1>
        <p className="mt-2 text-[var(--muted-foreground)]">
          {projects} projects · {apps} applications · {dbs} databases
        </p>
      </div>
      <div className="grid gap-4 md:grid-cols-3">
        <QuickCard title="Projects" href="/projects" count={projects} />
        <QuickCard title="Applications" href="/applications" count={apps} />
        <QuickCard title="Analytics" href="/analytics" />
      </div>
    </DashboardShell>
  );
}

function QuickCard({
  title,
  href,
  count,
}: {
  title: string;
  href: string;
  count?: number;
}) {
  return (
    <Link
      href={href}
      className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6 transition hover:border-[var(--primary)]"
    >
      <p className="text-lg font-medium">{title}</p>
      {count !== undefined ? (
        <p className="mt-2 text-3xl font-semibold tabular-nums">{count}</p>
      ) : (
        <p className="mt-2 text-sm text-[var(--muted-foreground)]">
          Server metrics & events
        </p>
      )}
    </Link>
  );
}
