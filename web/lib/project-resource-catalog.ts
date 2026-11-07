import type { LucideIcon } from "lucide-react";
import { Box, Database, Globe, HardDrive } from "lucide-react";
import type { SearchSelectOption } from "@/components/dockyrd/ui/search-select";
import {
  type ApplicationRow,
  type DatabaseRow,
  getStorageStatus,
  listApplications,
  listDatabases,
  listProjects,
  listServices,
  type Project,
  type Service,
} from "@/lib/api";

export type ProjectResourceKind = "app" | "static_site" | "postgres" | "bucket";

export const PROJECT_RESOURCE_KINDS: {
  id: ProjectResourceKind;
  label: string;
  icon: LucideIcon;
  apiKind: string;
}[] = [
  { id: "app", label: "App", icon: Box, apiKind: "app" },
  {
    id: "static_site",
    label: "Static site",
    icon: Globe,
    apiKind: "static_site",
  },
  { id: "postgres", label: "Database", icon: Database, apiKind: "postgres" },
  {
    id: "bucket",
    label: "Object storage",
    icon: HardDrive,
    apiKind: "bucket",
  },
];

export type MovableResource = {
  id: string;
  name: string;
  slug: string;
  kind: string;
  projectId: string;
  projectName: string;
};

export function resourceKindMeta(kind: string): {
  tag: string;
  surface: "application" | "static" | "database";
} {
  switch (kind) {
    case "static_site":
      return { tag: "STATIC SITE", surface: "static" };
    case "postgres":
      return { tag: "POSTGRES", surface: "database" };
    case "redis":
      return { tag: "REDIS", surface: "database" };
    case "bucket":
      return { tag: "BUCKET", surface: "application" };
    case "compose":
      return { tag: "COMPOSE", surface: "application" };
    default:
      return { tag: "APP", surface: "application" };
  }
}

export async function loadMovableResources(
  kind: ProjectResourceKind,
  currentProjectId: string,
): Promise<MovableResource[]> {
  if (kind === "app") {
    const { applications } = await listApplications();
    return applications
      .filter((a) => a.project_id !== currentProjectId)
      .map(appRowToMovable);
  }
  if (kind === "postgres") {
    const { databases } = await listDatabases();
    return databases
      .filter((d) => d.project_id !== currentProjectId && d.kind === "postgres")
      .map(dbRowToMovable);
  }
  if (kind === "static_site") {
    const { projects } = await listProjects();
    const out: MovableResource[] = [];
    for (const p of projects) {
      if (p.id === currentProjectId) continue;
      const { services } = await listServices(p.id, "static_site");
      for (const s of services) {
        out.push(serviceToMovable(s, p));
      }
    }
    return out;
  }
  const storage = await getStorageStatus();
  const buckets = storage.buckets ?? [];
  return buckets
    .filter((b) => b.project_id !== currentProjectId)
    .map((b) => ({
      id: b.id,
      name: b.name,
      slug: b.slug,
      kind: "bucket",
      projectId: b.project_id,
      projectName: b.project_name,
    }));
}

export function movableToSelectOptions(
  rows: MovableResource[],
): SearchSelectOption[] {
  return rows.map((r) => ({
    value: r.id,
    label: r.name,
    description: `${r.projectName} · ${r.slug}`,
  }));
}

function appRowToMovable(a: ApplicationRow): MovableResource {
  return {
    id: a.id,
    name: a.name,
    slug: a.slug,
    kind: "app",
    projectId: a.project_id,
    projectName: a.project_name,
  };
}

function dbRowToMovable(d: DatabaseRow): MovableResource {
  return {
    id: d.id,
    name: d.name,
    slug: d.slug,
    kind: d.kind,
    projectId: d.project_id,
    projectName: d.project_name,
  };
}

function serviceToMovable(s: Service, project: Project): MovableResource {
  return {
    id: s.id,
    name: s.name,
    slug: s.slug,
    kind: s.kind,
    projectId: project.id,
    projectName: project.name,
  };
}

export function servicePublicHost(service: Service): string | undefined {
  const cfg = service.config as {
    publicHost?: string;
    domain?: string;
    hostname?: string;
  };
  return cfg.publicHost ?? cfg.domain ?? cfg.hostname;
}
