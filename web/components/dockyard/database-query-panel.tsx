"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import {
  type DatabaseQueryResult,
  type DatabaseSchemaTable,
  getDatabaseSchema,
  runDatabaseQuery,
} from "@/lib/api";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type DatabaseQueryPanelProps = {
  serviceId: string;
};

export function DatabaseQueryPanel({ serviceId }: DatabaseQueryPanelProps) {
  const [sql, setSql] = useState(
    "SELECT tablename FROM pg_tables WHERE schemaname = 'public' LIMIT 20",
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<DatabaseQueryResult | null>(null);
  const [tables, setTables] = useState<DatabaseSchemaTable[]>([]);

  useEffect(() => {
    void getDatabaseSchema(serviceId)
      .then(({ tables: t }) => setTables(t))
      .catch(() => setTables([]));
  }, [serviceId]);

  function insertTableQuery(t: DatabaseSchemaTable) {
    const q =
      t.schema === "public"
        ? `SELECT * FROM "${t.name}" LIMIT 20`
        : `SELECT * FROM "${t.schema}"."${t.name}" LIMIT 20`;
    setSql(q);
  }

  async function onRun(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const res = await runDatabaseQuery(serviceId, sql);
      setResult(res);
    } catch (err) {
      setResult(null);
      setError(err instanceof Error ? err.message : "Query failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section
      className={RESOURCE_DETAIL_CARD}
      style={resourceDetailCardStyle()}
      aria-label="Query studio"
    >
      <h2 className="text-lg font-semibold tracking-tight">Query studio</h2>
      <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
        Read-only SQL against this Postgres instance. Runs on the server via
        psql; mutations are blocked.
      </p>
      <div className="mt-4 flex flex-col gap-6 lg:flex-row">
        {tables.length > 0 ? (
          <aside
            className="w-full shrink-0 lg:w-56"
            aria-label="Schema browser"
          >
            <p
              className="mb-2 font-mono text-[11px] uppercase tracking-widest"
              style={{ color: "var(--muted-text)" }}
            >
              Tables
            </p>
            <ul className="max-h-64 space-y-1 overflow-y-auto text-[13px]">
              {tables.map((t) => (
                <li key={`${t.schema}.${t.name}`}>
                  <button
                    type="button"
                    className="w-full rounded-md px-2 py-1.5 text-left hover:underline"
                    style={{ color: "var(--link-blue)" }}
                    onClick={() => insertTableQuery(t)}
                  >
                    {t.schema}.{t.name}
                  </button>
                </li>
              ))}
            </ul>
          </aside>
        ) : null}
        <div className="min-w-0 flex-1">
          <form className="space-y-4" onSubmit={(e) => void onRun(e)}>
            <Field label="SQL">
              <textarea
                id="database-query-sql"
                rows={8}
                value={sql}
                onChange={(e) => setSql(e.target.value)}
                className="field-input font-mono text-[13px]"
              />
            </Field>
            <Button type="submit" disabled={busy || !sql.trim()}>
              {busy ? "Running…" : "Run query"}
            </Button>
          </form>
          {error ? (
            <p className="mt-4 text-sm text-red-600" role="alert">
              {error}
            </p>
          ) : null}
          {result ? (
            <div className="mt-6 overflow-x-auto">
              <p
                className="mb-2 text-xs"
                style={{ color: "var(--muted-text)" }}
              >
                {result.row_count} row(s)
                {result.truncated ? " · truncated" : ""} · {result.duration_ms}{" "}
                ms
              </p>
              <table className="w-full min-w-[480px] border-collapse text-left text-[13px]">
                <thead>
                  <tr style={{ borderBottom: "1px solid var(--border)" }}>
                    {result.columns.map((col) => (
                      <th key={col} className="px-2 py-2 font-semibold">
                        {col}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {result.rows.map((row) => {
                    const rowKey = result.columns
                      .map((col, j) => `${col}:${row[j] ?? ""}`)
                      .join("|");
                    return (
                      <tr
                        key={rowKey}
                        style={{ borderBottom: "1px solid var(--border)" }}
                      >
                        {result.columns.map((col, j) => (
                          <td
                            key={col}
                            className="max-w-xs truncate px-2 py-2 font-mono"
                            title={row[j]}
                          >
                            {row[j]}
                          </td>
                        ))}
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          ) : null}
        </div>
      </div>
    </section>
  );
}
