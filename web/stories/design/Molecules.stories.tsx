import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { Home, Settings } from "lucide-react";
import type { ReactNode } from "react";
import { useState } from "react";
import {
  AddCard,
  Alert,
  DeploymentRow,
  EmptyState,
  EnvVarRow,
  EventFeedItem,
  FilterChips,
  KeyValueRow,
  ListCard,
  LogLine,
  LogStatusSummary,
  NavMenuItem,
  ProjectCard,
  ResourceRow,
  ServiceCard,
  StatCard,
  Tabs,
  TimeRangeToggle,
} from "@/components/dockyard/ui";
import { CatalogCanvas, CatalogGrid, CatalogRow } from "./catalog-layout";

const meta = {
  title: "Design/2 · Molecules",
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

export const CardsAndTiles: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Cards & tiles">
        <CatalogGrid>
          <ListCard
            title="Recently updated"
            description="Applications and services with activity in the last 24 hours."
            stat="4"
          />
          <StatCard label="Active services" value={12} hint="+2 this week" />
          <AddCard
            title="Create project"
            description="Group services under one repo."
          />
        </CatalogGrid>
      </Section>
    </CatalogCanvas>
  ),
};

export const ResourceRows: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Resource rows">
        <div
          className="max-w-2xl space-y-0 rounded-[20px] border p-4"
          style={{ borderColor: "var(--border)" }}
        >
          <ResourceRow
            name="marketing"
            meta="marketing.example.com"
            statusLabel="Live"
            statusVariant="ok"
            sha="a1b2c3d4e5f6"
            commitMessage="Fix cache headers"
          />
          <ResourceRow
            name="api-worker"
            meta="api.example.com"
            statusLabel="Deploying"
            statusVariant="info"
          />
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const DeploymentServiceProjectCards: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Deployment, service & project cards">
        <CatalogGrid>
          <ProjectCard
            name="Client X"
            repo="github.com/acme/client-x"
            servicesCount={4}
            statusLabel="Healthy"
            statusVariant="ok"
          />
          <ServiceCard
            name="marketing"
            domain="marketing.example.com"
            statusLabel="Live"
            statusVariant="ok"
          />
        </CatalogGrid>
        <div
          className="mt-4 max-w-2xl rounded-[20px] border p-4"
          style={{ borderColor: "var(--border)" }}
        >
          <DeploymentRow
            name="Production deploy"
            meta="Triggered by push to main"
            statusLabel="Live"
            statusVariant="ok"
            sha="a1b2c3d"
            commitMessage="Fix cache headers"
          />
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const NavigationAndSwitching: StoryObj = {
  render: () => {
    const [tab, setTab] = useState("overview");
    const [range, setRange] = useState<"1h" | "24h" | "7d" | "30d">("24h");
    const [filter, setFilter] = useState("all");
    return (
      <CatalogCanvas>
        <Section title="Navigation & switching">
          <CatalogRow>
            <NavMenuItem label="Home" icon={Home} active />
            <NavMenuItem label="Settings" icon={Settings} />
          </CatalogRow>
          <Tabs
            activeId={tab}
            onChange={setTab}
            tabs={[
              { id: "overview", label: "Overview" },
              { id: "deployments", label: "Deployments" },
              { id: "logs", label: "Logs" },
            ]}
          />
          <TimeRangeToggle value={range} onChange={setRange} />
          <FilterChips
            active={filter}
            onChange={setFilter}
            options={[
              { value: "all", label: "All" },
              { value: "live", label: "Live" },
              { value: "deploying", label: "Deploying" },
            ]}
          />
        </Section>
      </CatalogCanvas>
    );
  },
};

export const Feedback: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Feedback">
        <div className="flex max-w-xl flex-col gap-4">
          <Alert
            tone="info"
            message="Deploy queued — building image marketing:sha-a1b2."
          />
          <EmptyState
            title="No buckets yet"
            description="Create a bucket to store assets for your static sites."
            actionLabel="Create bucket"
            actionHref="#"
          />
          <LogStatusSummary counts={{ "2xx": 1240, "4xx": 3, "5xx": 1 }} />
          <EventFeedItem
            time="14:32 UTC"
            title="Deploy finished"
            detail="marketing · a1b2c3d"
            actor="joey@example.com"
          />
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const DataRows: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Data rows">
        <div className="max-w-2xl space-y-6">
          <div
            className="rounded-[20px] border p-4"
            style={{ borderColor: "var(--border)" }}
          >
            <EnvVarRow name="NODE_ENV" value="production" />
            <EnvVarRow name="API_TOKEN" value="sk_live_••••••••" secret />
          </div>
          <dl
            className="rounded-[20px] border px-4"
            style={{ borderColor: "var(--border)" }}
          >
            <KeyValueRow label="Region" value="eu-central-1" />
            <KeyValueRow label="Plan" value="Self-hosted" />
          </dl>
          <LogLine
            timestamp="14:32:01"
            status={200}
            method="GET"
            path="/health"
          />
        </div>
      </Section>
    </CatalogCanvas>
  ),
};
