import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { PathnameContext } from "next/dist/shared/lib/hooks-client-context.shared-runtime.js";
import type { ReactNode } from "react";
import { useState } from "react";
import { AppHeader } from "@/components/dockyrd/app-header";
import { AuthLayout } from "@/components/dockyrd/auth-layout";
import { Sidebar } from "@/components/dockyrd/sidebar";
import { TopBar } from "@/components/dockyrd/top-bar";
import {
  Button,
  ChartPanel,
  ContainerTable,
  DangerZone,
  DnsRecordBlock,
  DomainRow,
  Field,
  PageHeader,
  SearchPalette,
  ServerHealthCard,
  SessionRow,
  TemplateCard,
} from "@/components/dockyrd/ui";
import { CatalogCanvas, CatalogGrid, CatalogRow } from "./catalog-layout";

const meta = {
  title: "Design/3 · Organisms",
} satisfies Meta;

export default meta;

function Section({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) {
  return (
    <section
      className="flex flex-col gap-6 rounded-[20px] border p-6"
      style={{ borderColor: "var(--border)", background: "var(--surface)" }}
    >
      <div>
        <h2 className="text-2xl font-semibold tracking-tight">{title}</h2>
        {description ? (
          <p className="mt-1.5 text-sm" style={{ color: "var(--muted-text)" }}>
            {description}
          </p>
        ) : null}
      </div>
      {children}
    </section>
  );
}

function SidebarPreview({
  pathname,
  label,
}: {
  pathname: string;
  label: string;
}) {
  return (
    <div className="flex flex-col gap-2">
      <span
        className="font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {label}
      </span>
      <PathnameContext.Provider value={pathname}>
        <div
          className="relative h-[560px] w-[232px] overflow-hidden rounded-2xl border [&_aside]:relative [&_aside]:inset-auto [&_aside]:h-full [&_aside]:w-full"
          style={{ borderColor: "var(--border)" }}
        >
          <Sidebar />
        </div>
      </PathnameContext.Provider>
    </div>
  );
}

export const Sidebars: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section
        title="Sidebars"
        description="Main navigation and the settings variant (swaps all items)."
      >
        <CatalogRow>
          <SidebarPreview pathname="/" label="Main · pathname /" />
          <SidebarPreview
            pathname="/settings/account/profile"
            label="Account settings · /settings/account/profile"
          />
        </CatalogRow>
      </Section>
    </CatalogCanvas>
  ),
};

export const PageHeaderTopBarHero: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Page header, top bar & hero">
        <TopBar />
        <AppHeader
          breadcrumb="Home"
          title="Dashboard"
          user={{
            id: "1",
            name: "Joey de Kruis",
            email: "joey@derisory.nl",
            role: "owner",
            totp_enabled: true,
          }}
          onOpenSearch={() => undefined}
        />
        <PageHeader
          eyebrow="Project"
          title="Client X"
          subtitle="Production workloads for the marketing site and API."
          action={<Button>Create service</Button>}
        />
        <div
          className="flex min-h-[120px] max-w-2xl items-end rounded-[20px] p-8 text-white"
          style={{ background: "var(--hero-gradient)" }}
        >
          <p className="max-w-md text-sm leading-relaxed text-white/90">
            Deploy to your own servers. Without the mess — a calm control plane
            for applications, databases, and static sites.
          </p>
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const CommandPalette: StoryObj = {
  render: () => {
    const [query, setQuery] = useState("app");
    const results = [
      { kind: "page" as const, label: "Applications", href: "/applications" },
      {
        kind: "project" as const,
        label: "marketing",
        href: "/projects/marketing",
        meta: "Static site · Client X",
      },
    ].filter((row) => row.label.toLowerCase().includes(query.toLowerCase()));
    return (
      <CatalogCanvas>
        <Section title="Command palette">
          <div className="mx-auto max-w-[740px]">
            <SearchPalette
              query={query}
              onQueryChange={setQuery}
              results={results}
              activeIndex={1}
              onClose={() => undefined}
            />
          </div>
        </Section>
      </CatalogCanvas>
    );
  },
};

export const Charts: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Charts">
        <CatalogGrid>
          <ChartPanel title="Request rate" subtitle="Last 24 hours" />
          <ChartPanel
            title="Error rate"
            subtitle="5xx responses"
            height={140}
          />
        </CatalogGrid>
      </Section>
    </CatalogCanvas>
  ),
};

export const FeaturePanels: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Feature panels">
        <CatalogGrid>
          <ServerHealthCard version="v0.0.27" statusLine="Update available" />
          <TemplateCard
            name="Next.js"
            description="Static export or Node server with zero-downtime deploys."
            stack="Node 22"
            recommended
          />
        </CatalogGrid>
      </Section>
    </CatalogCanvas>
  ),
};

export const SettingsDomainPanels: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Settings & domain panels">
        <div className="max-w-2xl space-y-6">
          <div
            className="rounded-[20px] border p-4"
            style={{ borderColor: "var(--border)" }}
          >
            <SessionRow
              device="MacBook Pro · Chrome"
              location="Amsterdam, NL"
              lastActive="Active now"
              current
            />
            <SessionRow
              device="iPhone · Safari"
              location="Amsterdam, NL"
              lastActive="2 days ago"
              onRevoke={() => undefined}
            />
          </div>
          <div
            className="rounded-[20px] border px-4"
            style={{ borderColor: "var(--border)" }}
          >
            <DomainRow
              hostname="api.clientx.nl"
              statusLabel="Verified"
              statusVariant="ok"
              primary
            />
            <DomainRow
              hostname="www.clientx.nl"
              statusLabel="Pending"
              statusVariant="warn"
            />
          </div>
          <DnsRecordBlock
            type="CNAME"
            name="api"
            value="edge.dockyrd.dev"
            ttl="3600"
          />
          <DangerZone
            title="Delete project"
            description="This permanently removes all services, volumes, and deployment history."
            action={
              <Button variant="outline" style={{ color: "var(--danger)" }}>
                Delete project
              </Button>
            }
          />
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const TablesAndLayouts: StoryObj = {
  render: () => (
    <>
      <CatalogCanvas>
        <Section title="Tables & layouts">
          <ContainerTable
            rows={[
              {
                id: "1",
                name: "marketing-web",
                image: "ghcr.io/acme/marketing:latest",
                statusLabel: "Running",
                statusVariant: "ok",
                cpu: "12%",
                memory: "256 MiB",
              },
              {
                id: "2",
                name: "api-worker",
                image: "ghcr.io/acme/api:sha-a1b2",
                statusLabel: "Starting",
                statusVariant: "info",
                cpu: "—",
                memory: "—",
              },
            ]}
          />
          <p className="text-sm" style={{ color: "var(--muted-text)" }}>
            Auth layout (full bleed below)
          </p>
        </Section>
      </CatalogCanvas>
      <AuthLayout title="Sign in" subtitle="Use your dockyrd account.">
        <Field label="Email" type="email" defaultValue="you@example.com" />
        <div className="mt-4">
          <Field label="Password" type="password" />
        </div>
        <div className="mt-6">
          <Button>Continue</Button>
        </div>
      </AuthLayout>
    </>
  ),
  parameters: { layout: "fullscreen" },
};
