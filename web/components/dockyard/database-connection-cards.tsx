"use client";

import { Copy, Eye, EyeOff } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { Button } from "@/components/dockyard/ui/button";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import { KeyValueRow } from "@/components/dockyard/ui/key-value-row";
import {
  getServiceConnection,
  patchDatabasePublicAccess,
  type ServiceConnectionResponse,
} from "@/lib/api";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

const MASK = "••••••••";

function isSecretVariableKey(key: string): boolean {
  const upper = key.toUpperCase();
  return (
    upper.includes("PASSWORD") || key === "DATABASE_URL" || key === "REDIS_URL"
  );
}

function ConnectionValue({
  name,
  maskedValue,
  plainSecrets,
  revealed,
  onReveal,
  onHide,
}: {
  name: string;
  maskedValue: string;
  plainSecrets: Record<string, string> | null;
  revealed: boolean;
  onReveal: () => void;
  onHide: () => void;
}) {
  const secret = isSecretVariableKey(name);
  const plain = plainSecrets?.[name];
  const display = secret && !revealed ? MASK : (plain ?? maskedValue);
  const copyValue = secret ? (plain ?? "") : maskedValue;

  return (
    <span className="inline-flex max-w-full flex-wrap items-center gap-1 sm:gap-2">
      <span className="break-all font-mono text-sm">{display}</span>
      <span className="inline-flex shrink-0 items-center gap-0.5">
        {secret ? (
          <IconButton
            icon={revealed ? EyeOff : Eye}
            label={revealed ? `Hide ${name}` : `Reveal ${name}`}
            onClick={() => {
              if (revealed) onHide();
              else onReveal();
            }}
          />
        ) : null}
        <IconButton
          icon={Copy}
          label={`Copy ${name}`}
          disabled={secret && !plain}
          onClick={() => {
            if (secret && !plain) return;
            void navigator.clipboard.writeText(copyValue);
          }}
        />
      </span>
    </span>
  );
}

function ConnectionFields({
  variables,
  plainSecrets,
  revealedKeys,
  onRevealKey,
  onHideKey,
}: {
  variables: Record<string, string>;
  plainSecrets: Record<string, string> | null;
  revealedKeys: Set<string>;
  onRevealKey: (key: string) => void;
  onHideKey: (key: string) => void;
}) {
  const keys = Object.keys(variables).sort();
  if (keys.length === 0) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
        No connection variables.
      </p>
    );
  }
  return (
    <dl>
      {keys.map((key) => (
        <KeyValueRow
          key={key}
          label={key}
          value={
            <ConnectionValue
              name={key}
              maskedValue={variables[key] ?? ""}
              plainSecrets={plainSecrets}
              revealed={revealedKeys.has(key)}
              onReveal={() => onRevealKey(key)}
              onHide={() => onHideKey(key)}
            />
          }
        />
      ))}
    </dl>
  );
}

type DatabaseConnectionCardsProps = {
  serviceId: string;
  kind: string;
};

export function DatabaseConnectionCards({
  serviceId,
  kind,
}: DatabaseConnectionCardsProps) {
  const [conn, setConn] = useState<ServiceConnectionResponse | null>(null);
  const [plainSecrets, setPlainSecrets] = useState<Record<
    string,
    string
  > | null>(null);
  const [revealedKeys, setRevealedKeys] = useState<Set<string>>(
    () => new Set(),
  );
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [publicBusy, setPublicBusy] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    setPlainSecrets(null);
    setRevealedKeys(new Set());
    try {
      const data = await getServiceConnection(serviceId, false);
      setConn(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load connection");
    } finally {
      setLoading(false);
    }
  }, [serviceId]);

  useEffect(() => {
    void load();
  }, [load]);

  const revealKey = useCallback(
    async (key: string, scope: "internal" | "external") => {
      if (plainSecrets?.[key]) {
        setRevealedKeys((prev) => new Set(prev).add(key));
        return;
      }
      try {
        const data = await getServiceConnection(serviceId, true);
        setConn(data);
        const vars =
          scope === "external"
            ? (data.external?.variables ?? {})
            : (data.internal?.variables ?? data.variables);
        setPlainSecrets(vars);
        setRevealedKeys((prev) => new Set(prev).add(key));
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to reveal secret");
      }
    },
    [plainSecrets, serviceId],
  );

  async function togglePublicAccess(enable: boolean) {
    if (kind !== "postgres") return;
    setPublicBusy(true);
    setError(null);
    try {
      await patchDatabasePublicAccess(serviceId, { enabled: enable });
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Public access update failed");
    } finally {
      setPublicBusy(false);
    }
  }

  const internalVars = conn?.internal?.variables ?? conn?.variables ?? {};
  const external = conn?.external;

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <section
        className={RESOURCE_DETAIL_CARD}
        style={resourceDetailCardStyle()}
        aria-label="Internal connection"
      >
        <h2 className="text-lg font-semibold tracking-tight">
          Internal connection
        </h2>
        <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
          From other services on the same project network.
        </p>
        {loading ? (
          <DockyrdPageLoading className="py-2" cubeSize={28} />
        ) : (
          <div className="mt-4">
            <ConnectionFields
              variables={internalVars}
              plainSecrets={plainSecrets}
              revealedKeys={revealedKeys}
              onRevealKey={(k) => void revealKey(k, "internal")}
              onHideKey={(k) =>
                setRevealedKeys((prev) => {
                  const next = new Set(prev);
                  next.delete(k);
                  return next;
                })
              }
            />
          </div>
        )}
      </section>

      <section
        className={RESOURCE_DETAIL_CARD}
        style={resourceDetailCardStyle()}
        aria-label="External connection"
      >
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold tracking-tight">
              External connection
            </h2>
            <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
              Direct TCP to this server when public access is enabled. Not
              recommended for production without a firewall.
            </p>
          </div>
          {kind === "postgres" ? (
            external?.enabled ? (
              <Button
                type="button"
                variant="outline"
                disabled={publicBusy}
                onClick={() => void togglePublicAccess(false)}
              >
                Disable public access
              </Button>
            ) : (
              <Button
                type="button"
                variant="primary"
                disabled={publicBusy}
                onClick={() => void togglePublicAccess(true)}
              >
                Enable public access
              </Button>
            )
          ) : null}
        </div>
        {external?.enabled && external.variables ? (
          <div className="mt-4">
            <ConnectionFields
              variables={external.variables}
              plainSecrets={plainSecrets}
              revealedKeys={revealedKeys}
              onRevealKey={(k) => void revealKey(k, "external")}
              onHideKey={(k) =>
                setRevealedKeys((prev) => {
                  const next = new Set(prev);
                  next.delete(k);
                  return next;
                })
              }
            />
          </div>
        ) : (
          <p className="mt-4 text-sm" style={{ color: "var(--muted-text)" }}>
            {kind === "postgres"
              ? "Public access is off. Enable to expose a host port on this server."
              : "External access is not available for Redis in the panel yet."}
          </p>
        )}
      </section>
      {error ? (
        <p className="text-sm text-red-600 lg:col-span-2" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
