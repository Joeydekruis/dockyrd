"use client";

import { useCallback, useEffect, useState } from "react";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { KeyValueRow } from "@/components/dockyard/ui/key-value-row";
import { RevealableCopyField } from "@/components/dockyard/ui/revealable-copy-field";
import { getServiceConnection } from "@/lib/api";
import { isSecretConnectionVariableKey } from "@/lib/connection-secret";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

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
  const secret = isSecretConnectionVariableKey(name);
  const plain = plainSecrets?.[name];
  const display = plain ?? maskedValue;
  const copyValue = secret ? (plain ?? "") : maskedValue;
  const canCopy = !(secret && !plain);

  return (
    <RevealableCopyField
      displayValue={display}
      copyValue={copyValue}
      copyLabel={`Copy ${name}`}
      copyMessage={`Copied ${name}`}
      canCopy={canCopy}
      revealable={secret}
      revealed={revealed}
      onToggleReveal={() => {
        if (revealed) onHide();
        else onReveal();
      }}
      revealLabel={`Reveal ${name}`}
      hideLabel={`Hide ${name}`}
    />
  );
}

type ServiceConnectionPanelProps = {
  serviceId: string;
  title?: string;
  description?: string;
};

export function ServiceConnectionPanel({
  serviceId,
  title = "Connection",
  description = "Reachable only for linked applications. Reveal a secret to copy it.",
}: ServiceConnectionPanelProps) {
  const [variables, setVariables] = useState<Record<string, string>>({});
  const [plainSecrets, setPlainSecrets] = useState<Record<
    string,
    string
  > | null>(null);
  const [revealedKeys, setRevealedKeys] = useState<Set<string>>(
    () => new Set(),
  );
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [revealing, setRevealing] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    setPlainSecrets(null);
    setRevealedKeys(new Set());
    void getServiceConnection(serviceId, false)
      .then(({ variables: vars }) => {
        if (!cancelled) setVariables(vars);
      })
      .catch((e) => {
        if (!cancelled) {
          setError(
            e instanceof Error ? e.message : "Failed to load connection",
          );
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [serviceId]);

  const revealKey = useCallback(
    async (key: string) => {
      if (plainSecrets?.[key]) {
        setRevealedKeys((prev) => new Set(prev).add(key));
        return;
      }
      setRevealing(true);
      setError(null);
      try {
        const { variables: vars } = await getServiceConnection(serviceId, true);
        setPlainSecrets(vars);
        setRevealedKeys((prev) => new Set(prev).add(key));
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to reveal secret");
      } finally {
        setRevealing(false);
      }
    },
    [plainSecrets, serviceId],
  );

  function hideKey(key: string) {
    setRevealedKeys((prev) => {
      const next = new Set(prev);
      next.delete(key);
      return next;
    });
  }

  const keys = Object.keys(variables).sort();

  return (
    <section
      className={RESOURCE_DETAIL_CARD}
      style={resourceDetailCardStyle()}
      aria-label={title}
    >
      <div>
        <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
        <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
          {description}
        </p>
      </div>
      {error ? (
        <p className="mt-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : loading ? (
        <DockyrdPageLoading className="py-2" cubeSize={28} />
      ) : (
        <dl className="mt-4">
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
                  onReveal={() => void revealKey(key)}
                  onHide={() => hideKey(key)}
                />
              }
            />
          ))}
          {revealing ? (
            <p className="text-xs" style={{ color: "var(--muted-text)" }}>
              Loading secret…
            </p>
          ) : null}
          {keys.length === 0 ? (
            <p className="text-sm" style={{ color: "var(--muted-text)" }}>
              No connection variables for this service.
            </p>
          ) : null}
        </dl>
      )}
    </section>
  );
}
