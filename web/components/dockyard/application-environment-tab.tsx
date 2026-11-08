"use client";

import { Eye, EyeOff, Link2 } from "lucide-react";
import { useMemo, useState } from "react";
import { AddConnectionModal } from "@/components/dockyard/add-connection-modal";
import { AddEnvVariableModal } from "@/components/dockyard/add-env-variable-modal";
import { PendingEnvRedeployBanner } from "@/components/dockyard/pending-env-redeploy-banner";
import { Button } from "@/components/dockyard/ui/button";
import { Chip } from "@/components/dockyard/ui/chip";
import { CopyIconButton } from "@/components/dockyard/ui/copy-icon-button";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import type { Service, ServiceVariable } from "@/lib/api";
import {
  envVarScopeLabel,
  hasPendingEnvRedeploy,
  isBrokenEnvRef,
  parseEnvVarRef,
} from "@/lib/env-var-ref";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type ApplicationEnvironmentTabProps = {
  service: Service;
  variables: ServiceVariable[];
  linkedSlugs: ReadonlySet<string>;
  liveDeploymentAt?: number;
  revealedSecrets: boolean;
  onToggleRevealSecrets: () => void;
  onAddVariables: (
    entries: {
      key: string;
      value: string;
      is_secret: boolean;
      at_build: boolean;
      at_runtime: boolean;
    }[],
  ) => Promise<void>;
  onRefresh: () => Promise<void>;
  onDeploy?: () => void;
  deployBusy?: boolean;
  bannerDismissed?: boolean;
  onDismissBanner?: () => void;
};

export function ApplicationEnvironmentTab({
  service,
  variables,
  linkedSlugs,
  liveDeploymentAt,
  revealedSecrets,
  onToggleRevealSecrets,
  onAddVariables,
  onRefresh,
  onDeploy,
  deployBusy,
  bannerDismissed,
  onDismissBanner,
}: ApplicationEnvironmentTabProps) {
  const [addVarOpen, setAddVarOpen] = useState(false);
  const [connectOpen, setConnectOpen] = useState(false);
  const [rowReveal, setRowReveal] = useState<Record<string, boolean>>({});

  const showBanner = useMemo(
    () =>
      !bannerDismissed &&
      hasPendingEnvRedeploy({
        variables,
        linkedSlugs,
        serviceUpdatedAt: service.updated_at,
        liveDeploymentAt,
      }),
    [
      bannerDismissed,
      linkedSlugs,
      liveDeploymentAt,
      service.updated_at,
      variables,
    ],
  );

  return (
    <section className={RESOURCE_DETAIL_CARD} style={resourceDetailCardStyle()}>
      {showBanner ? (
        <div className="mb-6">
          <PendingEnvRedeployBanner
            onDeploy={onDeploy}
            deployBusy={deployBusy}
            onDismiss={onDismissBanner}
          />
        </div>
      ) : null}

      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">
            Environment variables
          </h2>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            Available to every process. Changes apply on the next deployment.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            variant="outline"
            iconLeft={revealedSecrets ? EyeOff : Eye}
            onClick={onToggleRevealSecrets}
          >
            {revealedSecrets ? "Hide secrets" : "Reveal secrets"}
          </Button>
          <Button type="button" onClick={() => setAddVarOpen(true)}>
            Add variable
          </Button>
        </div>
      </div>

      <div
        className="mt-4 hidden border-b py-2.5 text-[11px] font-medium uppercase tracking-widest md:grid md:grid-cols-[minmax(120px,1fr)_80px_120px_minmax(0,2fr)_auto]"
        style={{ borderColor: "var(--border)", color: "var(--muted-text)" }}
      >
        <span>Key</span>
        <span>Kind</span>
        <span>Scope</span>
        <span>Value</span>
        <span className="text-right">Actions</span>
      </div>

      <ul className="mt-0 list-none p-0">
        {variables.map((v) => {
          const ref = v.value ? parseEnvVarRef(v.value) : null;
          const broken = isBrokenEnvRef(v.value, linkedSlugs);
          const revealed = v.is_secret
            ? revealedSecrets || rowReveal[v.key]
            : true;
          const canCopy = Boolean(v.value) && (!v.is_secret || revealed);
          const display =
            v.is_secret && !revealed ? "••••••••" : (v.value ?? "");

          return (
            <li
              key={v.key}
              className="border-t py-4 first:border-t-0 md:grid md:grid-cols-[minmax(120px,1fr)_80px_120px_minmax(0,2fr)_auto] md:items-center md:gap-3"
              style={{ borderColor: "var(--border)" }}
            >
              <span
                className="font-mono text-[13px] font-semibold"
                style={{ color: "var(--ink-navy)" }}
              >
                {v.key}
              </span>
              <span
                className="mt-2 text-[13px] md:mt-0"
                style={{ color: "var(--muted-text)" }}
              >
                {v.is_secret ? "Secret" : "Plain"}
              </span>
              <span
                className="mt-1 text-[13px] md:mt-0"
                style={{ color: "var(--muted-text)" }}
              >
                {envVarScopeLabel(v)}
              </span>
              <div className="mt-2 min-w-0 md:mt-0">
                <p
                  className="truncate font-mono text-[13px]"
                  style={{ color: "var(--muted-text)" }}
                >
                  {display}
                </p>
                <div className="mt-1 flex flex-wrap gap-1">
                  {v.link_id ? <Chip>via link</Chip> : null}
                  {ref ? (
                    <Chip>
                      <Link2 className="mr-1 inline size-3" aria-hidden />
                      {ref.slug} · {ref.key}
                    </Chip>
                  ) : null}
                  {broken ? <Chip>Not linked</Chip> : null}
                </div>
              </div>
              <div className="mt-2 flex flex-wrap items-center justify-end gap-2 md:mt-0">
                {v.is_secret ? (
                  <IconButton
                    icon={revealed ? EyeOff : Eye}
                    label={revealed ? "Hide value" : "Show value"}
                    onClick={() =>
                      setRowReveal((prev) => ({
                        ...prev,
                        [v.key]: !prev[v.key],
                      }))
                    }
                  />
                ) : null}
                {broken ? (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setConnectOpen(true)}
                  >
                    Connect
                  </Button>
                ) : null}
                <CopyIconButton
                  value={v.value ?? ""}
                  label={`Copy ${v.key}`}
                  message={`Copied ${v.key}`}
                  disabled={!canCopy}
                />
              </div>
            </li>
          );
        })}
      </ul>

      {variables.length === 0 ? (
        <p
          className="py-8 text-center text-sm"
          style={{ color: "var(--muted-text)" }}
        >
          No environment variables yet.
        </p>
      ) : null}

      <AddEnvVariableModal
        open={addVarOpen}
        existingKeys={variables.map((v) => v.key)}
        onClose={() => setAddVarOpen(false)}
        onAdd={onAddVariables}
      />

      <AddConnectionModal
        open={connectOpen}
        onOpenChange={setConnectOpen}
        serviceId={service.id}
        onCreated={() => void onRefresh()}
      />
    </section>
  );
}
