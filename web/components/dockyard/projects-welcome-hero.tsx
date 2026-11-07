"use client";

import type { LucideIcon } from "lucide-react";
import { Box, Database, Globe, HardDrive } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/dockyard/ui/button";

type ServiceKind = "app" | "db" | "site" | "bucket";

const KINDS: {
  id: ServiceKind;
  icon: LucideIcon;
  title: string;
  description: string;
}[] = [
  {
    id: "app",
    icon: Box,
    title: "App",
    description:
      "Deploy from GitHub on every push, from a Dockerfile or from an image. Zero-downtime deploys and one-click rollback.",
  },
  {
    id: "db",
    icon: Database,
    title: "Database",
    description:
      "Managed Postgres or Redis with connection strings injected into your apps. Backups and metrics included.",
  },
  {
    id: "site",
    icon: Globe,
    title: "Static site",
    description:
      "Build from Git on push or upload a folder. Automatic HTTPS and global CDN-style caching at the edge.",
  },
  {
    id: "bucket",
    icon: HardDrive,
    title: "Object storage",
    description:
      "S3-compatible buckets on your server for assets, uploads, and backups.",
  },
];

type ProjectsWelcomeHeroProps = {
  onCreateProject: () => void;
};

export function ProjectsWelcomeHero({
  onCreateProject,
}: ProjectsWelcomeHeroProps) {
  const [kind, setKind] = useState<ServiceKind>("app");
  const active = KINDS.find((k) => k.id === kind) ?? KINDS[0];

  return (
    <section
      aria-label="Welcome"
      className="relative mb-10 overflow-hidden rounded-[20px] p-8 text-white lg:p-10"
      style={{ background: "var(--hero-gradient)" }}
    >
      <div className="flex flex-col gap-8 lg:flex-row lg:items-stretch lg:justify-between">
        <div className="max-w-xl flex flex-col gap-4 justify-center">
          <p className="auth-brand-badge">Self-hosted · Open source</p>
          <h1 className="auth-brand-headline mt-4 text-white">
            Deploy to your own servers. Without the mess.
          </h1>
          <p className="mt-4 max-w-lg text-sm leading-relaxed text-white/90">
            Connect a server, connect GitHub and push code. Apps, databases and
            static sites with domains, HTTPS, logs and metrics.
          </p>
        </div>

        <section
          className="w-full max-w-md shrink-0 rounded-[20px] border p-6 text-[var(--ink-navy)] lg:max-w-[380px]"
          style={{ background: "var(--surface)", borderColor: "var(--border)" }}
          aria-labelledby="projects-welcome-service-type"
        >
          <p
            id="projects-welcome-service-type"
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Add service
          </p>
          <fieldset className="mt-4 border-0 p-0">
            <legend className="sr-only">
              Choose a service type to preview
            </legend>
            <div className="flex flex-wrap gap-2">
              {KINDS.map((item) => {
                const selected = item.id === kind;
                const Icon = item.icon;
                return (
                  <button
                    key={item.id}
                    type="button"
                    aria-pressed={selected}
                    aria-label={item.title}
                    onClick={() => setKind(item.id)}
                    className="flex size-11 items-center justify-center rounded-full border-2 transition-colors"
                    style={{
                      borderColor: selected
                        ? "var(--accent-blue)"
                        : "var(--border)",
                      background: selected
                        ? "var(--brand-soft)"
                        : "transparent",
                      color: selected
                        ? "var(--link-blue)"
                        : "var(--secondary-text)",
                    }}
                  >
                    <Icon strokeWidth={1.8} className="size-5" aria-hidden />
                  </button>
                );
              })}
            </div>
          </fieldset>
          <h2 className="mt-5 text-lg font-semibold">{active.title}</h2>
          <p
            className="mt-2 text-sm leading-relaxed"
            style={{ color: "var(--muted-text)" }}
          >
            {active.description}
          </p>
          <Button
            type="button"
            variant="link"
            className="mt-4 px-0"
            onClick={onCreateProject}
          >
            Create a project →
          </Button>
        </section>
      </div>
    </section>
  );
}
