"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { SettingsPageBody } from "@/components/dockyard/settings-page-body";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  createRemoteBuilder,
  deleteRemoteBuilder,
  listBuilders,
  type RemoteBuilder,
  testRemoteBuilder,
} from "@/lib/api";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

export function ServerBuildersSettingsPage() {
  const [builders, setBuilders] = useState<RemoteBuilder[]>([]);
  const [name, setName] = useState("");
  const [endpoint, setEndpoint] = useState("");
  const [dockerRun, setDockerRun] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const refresh = useCallback(async () => {
    try {
      const { builders: rows } = await listBuilders();
      setBuilders(rows);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load builders");
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    setDockerRun(null);
    try {
      const created = await createRemoteBuilder({
        name: name.trim(),
        endpoint: endpoint.trim(),
      });
      setDockerRun(created.docker_run);
      setName("");
      setEndpoint("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <DashboardShell title="Builders" breadcrumb="Server settings">
      <SettingsPageBody>
        <section
          className={RESOURCE_DETAIL_CARD}
          style={resourceDetailCardStyle()}
        >
          <h1 className="text-2xl font-semibold tracking-tight">
            Remote BuildKit
          </h1>
          <p className="mt-2 text-sm text-[var(--muted-foreground)]">
            Register a mTLS BuildKit node on another machine. Services can
            target it under Build location.
          </p>
          <form onSubmit={(e) => void onCreate(e)} className="mt-6 space-y-3">
            <Field
              label="Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Office builder"
            />
            <Field
              label="Endpoint"
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder="builder.example.com:1234"
            />
            {error ? (
              <p className="text-sm text-[var(--danger)]">{error}</p>
            ) : null}
            <Button
              type="submit"
              disabled={busy || !name.trim() || !endpoint.trim()}
            >
              {busy ? "Creating…" : "Add builder"}
            </Button>
          </form>
          {dockerRun ? (
            <pre className="mt-4 overflow-x-auto rounded-lg border p-3 text-xs">
              {dockerRun}
            </pre>
          ) : null}
        </section>

        <section
          className={`${RESOURCE_DETAIL_CARD} mt-8`}
          style={resourceDetailCardStyle()}
        >
          <h2 className="text-lg font-medium">Registered builders</h2>
          {builders.length === 0 ? (
            <p className="mt-3 text-sm text-[var(--muted-foreground)]">
              No remote builders yet.
            </p>
          ) : (
            <ul className="mt-4 divide-y">
              {builders.map((b) => (
                <li
                  key={b.id}
                  className="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0"
                >
                  <div>
                    <p className="font-medium">{b.name}</p>
                    <p className="font-mono text-xs text-[var(--muted-foreground)]">
                      {b.endpoint}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <StatusBadge
                      label={b.status ?? "unknown"}
                      variant={
                        b.status === "online"
                          ? "ok"
                          : b.status === "error"
                            ? "err"
                            : "info"
                      }
                    />
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() =>
                        void testRemoteBuilder(b.id)
                          .then(refresh)
                          .catch((e) =>
                            setError(
                              e instanceof Error ? e.message : "Test failed",
                            ),
                          )
                      }
                    >
                      Test
                    </Button>
                    <Button
                      type="button"
                      variant="danger"
                      onClick={() =>
                        void deleteRemoteBuilder(b.id).then(refresh)
                      }
                    >
                      Remove
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>
      </SettingsPageBody>
    </DashboardShell>
  );
}
