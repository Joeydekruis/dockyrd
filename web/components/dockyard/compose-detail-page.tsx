"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { ServiceDomainsTab } from "@/components/dockyard/service-domains-tab";
import { ServiceLogsPanel } from "@/components/dockyard/service-logs-panel";
import { Button } from "@/components/dockyard/ui/button";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { Field } from "@/components/dockyard/ui/field";
import { SelectField } from "@/components/dockyard/ui/select-field";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import { Tabs } from "@/components/dockyard/ui/tabs";
import {
  createDeployment,
  type Deployment,
  listComposeContainers,
  listDeployments,
  listProjects,
  listServiceDomains,
  listServices,
  type Project,
  type Service,
  type ServiceDomain,
  validateComposeYaml,
} from "@/lib/api";
import {
  projectPathLabel,
  resourceStatusDisplay,
} from "@/lib/resource-dashboard";
import {
  persistDetailTab,
  RESOURCE_DETAIL_CARD,
  readInitialDetailTab,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";
import {
  deploymentStatusDisplay,
  formatDeploymentWhen,
} from "@/lib/static-site-display";
import { useRouteParam } from "@/lib/use-route-param";

const TABS = [
  { id: "overview", label: "Overview" },
  { id: "deployments", label: "Deployments" },
  { id: "logs", label: "Logs" },
  { id: "domains", label: "Domains" },
  { id: "settings", label: "Settings" },
] as const;

type TabId = (typeof TABS)[number]["id"];

export function ComposeDetailPage() {
  const slug = useRouteParam("slug");
  const [tab, setTab] = useState<TabId>(() =>
    readInitialDetailTab(
      TABS.map((t) => t.id),
      "overview",
    ),
  );
  const [project, setProject] = useState<Project | null>(null);
  const [service, setService] = useState<Service | null>(null);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [domains, setDomains] = useState<ServiceDomain[]>([]);
  const [containers, setContainers] = useState<
    Awaited<ReturnType<typeof listComposeContainers>>["containers"]
  >([]);
  const [logService, setLogService] = useState("");
  const [domainTargetService, setDomainTargetService] = useState("web");
  const [domainTargetPort, setDomainTargetPort] = useState("80");
  const [settingsYaml, setSettingsYaml] = useState("");
  const [validateResult, setValidateResult] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [busyDeploy, setBusyDeploy] = useState(false);

  const refresh = useCallback(async () => {
    if (!slug) return;
    setError(null);
    const { projects } = await listProjects();
    for (const p of projects) {
      const { services } = await listServices(p.id, "compose");
      const match = services.find((s) => s.slug === slug);
      if (match) {
        setProject(p);
        setService(match);
        const src = match.config?.source as { yaml?: string } | undefined;
        setSettingsYaml(src?.yaml ?? "");
        const [deps, dom, cont] = await Promise.all([
          listDeployments(match.id),
          listServiceDomains(match.id),
          listComposeContainers(match.id).catch(() => ({ containers: [] })),
        ]);
        setDeployments(deps.deployments ?? []);
        setDomains(dom.domains ?? []);
        setContainers(cont.containers ?? []);
        if (!logService && cont.containers[0]?.service) {
          setLogService(cont.containers[0].service);
        }
        return;
      }
    }
    setProject(null);
    setService(null);
  }, [slug, logService]);

  useEffect(() => {
    setLoading(true);
    void refresh()
      .catch((e) => setError(e instanceof Error ? e.message : "Load failed"))
      .finally(() => setLoading(false));
  }, [refresh]);

  useEffect(() => {
    setTab(
      readInitialDetailTab(
        TABS.map((t) => t.id),
        "overview",
      ),
    );
  }, []);

  function selectTab(next: TabId) {
    setTab(next);
    persistDetailTab(next);
  }

  const status = useMemo(() => {
    if (!service) return null;
    return resourceStatusDisplay(service.desired_state, "application");
  }, [service]);

  const title = service?.name ?? "Compose stack";

  async function onRedeploy() {
    if (!service) return;
    setBusyDeploy(true);
    try {
      await createDeployment(service.id);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Deploy failed");
    } finally {
      setBusyDeploy(false);
    }
  }

  async function onValidateYaml() {
    if (!service) return;
    setValidateResult(null);
    try {
      const res = await validateComposeYaml(service.id, settingsYaml);
      const lines = [
        res.ok ? "Validation passed." : "Validation failed.",
        ...res.errors.map((e) => `Error: ${e.message}`),
        ...res.warnings.map((w) => `Warning: ${w.message}`),
      ];
      setValidateResult(lines.join("\n"));
    } catch (e) {
      setValidateResult(e instanceof Error ? e.message : "Validate failed");
    }
  }

  if (loading) return <DockyrdPageLoading />;
  if (!service || !project) {
    return (
      <DashboardShell title="Compose stack" breadcrumb="Stacks">
        <EmptyState title="Stack not found" description="No compose service matches this URL." />
      </DashboardShell>
    );
  }

  const composeTarget = {
    service: domainTargetService.trim() || "web",
    port: Number(domainTargetPort) || 80,
  };

  return (
    <DashboardShell title={title} breadcrumb="Stacks">
      <div className="mx-auto max-w-[960px] space-y-6 px-6 py-8">
        <div className="space-y-2">
          <p className="font-mono text-[11px] uppercase tracking-widest" style={{ color: "var(--muted-text)" }}>
            {projectPathLabel(project.name)} · Compose stack
          </p>
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-semibold tracking-tight">{service.name}</h1>
            {status ? <StatusBadge label={status.label} variant={status.variant} /> : null}
          </div>
          <p className="text-sm" style={{ color: "var(--muted-text)" }}>
            Zero-downtime deploy is not available for Compose stacks. Host{" "}
            <code className="font-mono text-xs">ports:</code> are stripped — attach HTTPS domains instead.
          </p>
        </div>

        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}

        <Tabs
          tabs={[...TABS]}
          activeId={tab}
          onChange={(id) => selectTab(id as TabId)}
        />

        {tab === "overview" ? (
          <section className={RESOURCE_DETAIL_CARD} style={resourceDetailCardStyle()}>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <h2 className="text-lg font-semibold">Containers</h2>
              <Button type="button" variant="outline" onClick={() => void refresh()}>
                Refresh
              </Button>
            </div>
            {containers.length === 0 ? (
              <p className="mt-4 text-sm" style={{ color: "var(--muted-text)" }}>
                No containers yet — deploy the stack to start services.
              </p>
            ) : (
              <ul className="mt-4 space-y-2 text-sm">
                {containers.map((c) => (
                  <li key={c.name} className="flex flex-wrap gap-2 font-mono text-xs">
                    <span>{c.service}</span>
                    <span style={{ color: "var(--muted-text)" }}>{c.state}</span>
                    {c.health ? <span>{c.health}</span> : null}
                  </li>
                ))}
              </ul>
            )}
          </section>
        ) : null}

        {tab === "deployments" ? (
          <section className={RESOURCE_DETAIL_CARD} style={resourceDetailCardStyle()}>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <h2 className="text-lg font-semibold">Deployments</h2>
              <Button type="button" disabled={busyDeploy} onClick={() => void onRedeploy()}>
                {busyDeploy ? "Deploying…" : "Deploy now"}
              </Button>
            </div>
            <ul className="mt-4 divide-y text-sm" style={{ borderColor: "var(--border)" }}>
              {deployments.map((d) => {
                const st = deploymentStatusDisplay(d.status);
                return (
                  <li key={d.id} className="flex flex-wrap items-center justify-between gap-2 py-3">
                    <div>
                      <Link href={`/deployments/${d.id}`} className="font-mono text-xs hover:underline">
                        {d.id.slice(0, 8)}
                      </Link>
                      <span className="ml-2" style={{ color: "var(--muted-text)" }}>
                        {formatDeploymentWhen(d.finished_at ?? d.queued_at)}
                      </span>
                    </div>
                    <StatusBadge label={st.label} variant={st.variant} />
                  </li>
                );
              })}
            </ul>
          </section>
        ) : null}

        {tab === "logs" ? (
          <div className="space-y-4">
            <SelectField
              label="Compose service"
              value={logService}
              onChange={setLogService}
              options={containers.map((c) => ({
                value: c.service,
                label: c.service,
              }))}
              placeholder="Select service"
            />
            <ServiceLogsPanel
              serviceId={service.id}
              composeService={logService}
              description="Logs for the selected compose service container."
            />
          </div>
        ) : null}

        {tab === "domains" ? (
          <div className="space-y-4">
            <div
              className="grid gap-4 rounded-[20px] border p-4 sm:grid-cols-2"
              style={{ borderColor: "var(--border)" }}
            >
              <Field
                label="Target service"
                value={domainTargetService}
                onChange={(e) => setDomainTargetService(e.target.value)}
              />
              <Field
                label="Target port"
                value={domainTargetPort}
                onChange={(e) => setDomainTargetPort(e.target.value)}
                inputMode="numeric"
              />
            </div>
            <ServiceDomainsTab
              serviceId={service.id}
              domains={domains}
              layout="compose"
              composeTarget={composeTarget}
              onRefresh={async () => {
                const dom = await listServiceDomains(service.id);
                setDomains(dom.domains ?? []);
              }}
              onError={setError}
            />
          </div>
        ) : null}

        {tab === "settings" ? (
          <section className={RESOURCE_DETAIL_CARD} style={resourceDetailCardStyle()}>
            <h2 className="text-lg font-semibold">Compose YAML</h2>
            <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
              Validate before redeploying. Saving inline YAML to the service is coming in a follow-up; redeploy uses the last created config until then.
            </p>
            <Field label="YAML">
              <textarea
                className="field-input mt-1 min-h-[240px] font-mono text-xs"
                value={settingsYaml}
                onChange={(e) => setSettingsYaml(e.target.value)}
                spellCheck={false}
              />
            </Field>
            <div className="mt-4 flex flex-wrap gap-2">
              <Button type="button" variant="outline" onClick={() => void onValidateYaml()}>
                Validate
              </Button>
            </div>
            {validateResult ? (
              <pre className="mt-4 whitespace-pre-wrap font-mono text-xs">{validateResult}</pre>
            ) : null}
          </section>
        ) : null}
      </div>
    </DashboardShell>
  );
}
