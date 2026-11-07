"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import { Tabs } from "@/components/dockyard/ui/tabs";
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

type StudioMode = "console" | "browse";
type BrowseTab = "data" | "structure";

const BROWSE_TABS = [
  { id: "data", label: "Data" },
  { id: "structure", label: "Structure" },
] as const;

function tableKey(t: DatabaseSchemaTable) {
  return `${t.schema}.${t.name}`;
}

function selectTableSQL(t: DatabaseSchemaTable, limit = 200) {
  if (t.schema === "public") {
    return `SELECT * FROM "${t.name}" LIMIT ${limit}`;
  }
  return `SELECT * FROM "${t.schema}"."${t.name}" LIMIT ${limit}`;
}

function QueryResultTable({
  result,
  columnTypes,
}: {
  result: DatabaseQueryResult;
  columnTypes?: Map<string, string>;
}) {
  return (
    <div className="overflow-x-auto">
      <p className="mb-2 text-xs" style={{ color: "var(--muted-text)" }}>
        {result.row_count} row(s)
        {result.truncated ? " · truncated" : ""} · {result.duration_ms} ms
      </p>
      <table className="w-full min-w-[480px] border-collapse text-left text-[13px]">
        <thead>
          <tr style={{ borderBottom: "1px solid var(--border)" }}>
            {result.columns.map((col) => (
              <th key={col} className="px-2 py-2 align-bottom font-semibold">
                <div>{col}</div>
                {columnTypes?.get(col) ? (
                  <div
                    className="mt-0.5 font-normal font-mono text-[11px]"
                    style={{ color: "var(--muted-text)" }}
                  >
                    {columnTypes.get(col)}
                  </div>
                ) : null}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {result.rows.length === 0 ? (
            <tr>
              <td
                colSpan={Math.max(result.columns.length, 1)}
                className="px-2 py-6 text-center"
                style={{ color: "var(--muted-text)" }}
              >
                No rows
              </td>
            </tr>
          ) : (
            result.rows.map((row) => {
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
                      {row[j] ?? ""}
                    </td>
                  ))}
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
}

export function DatabaseQueryPanel({ serviceId }: DatabaseQueryPanelProps) {
  const [mode, setMode] = useState<StudioMode>("browse");
  const [browseTab, setBrowseTab] = useState<BrowseTab>("data");
  const [tableSearch, setTableSearch] = useState("");
  const [schemaLoading, setSchemaLoading] = useState(true);
  const [tables, setTables] = useState<DatabaseSchemaTable[]>([]);
  const [selectedTable, setSelectedTable] =
    useState<DatabaseSchemaTable | null>(null);

  const [sql, setSql] = useState(
    "SELECT tablename FROM pg_tables WHERE schemaname = 'public' LIMIT 20",
  );
  const [consoleBusy, setConsoleBusy] = useState(false);
  const [consoleError, setConsoleError] = useState<string | null>(null);
  const [consoleResult, setConsoleResult] =
    useState<DatabaseQueryResult | null>(null);

  const [browseBusy, setBrowseBusy] = useState(false);
  const [browseError, setBrowseError] = useState<string | null>(null);
  const [browseResult, setBrowseResult] = useState<DatabaseQueryResult | null>(
    null,
  );

  const loadSchema = useCallback(async () => {
    setSchemaLoading(true);
    try {
      const { tables: t } = await getDatabaseSchema(serviceId);
      setTables(t);
      setSelectedTable((prev) => {
        if (!prev) return t[0] ?? null;
        const still = t.find((x) => tableKey(x) === tableKey(prev));
        return still ?? t[0] ?? null;
      });
    } catch {
      setTables([]);
      setSelectedTable(null);
    } finally {
      setSchemaLoading(false);
    }
  }, [serviceId]);

  useEffect(() => {
    void loadSchema();
  }, [loadSchema]);

  const filteredTables = useMemo(() => {
    const q = tableSearch.trim().toLowerCase();
    if (!q) return tables;
    return tables.filter((t) => tableKey(t).toLowerCase().includes(q));
  }, [tables, tableSearch]);

  const columnTypeMap = useMemo(() => {
    if (!selectedTable) return undefined;
    return new Map(
      selectedTable.columns.map((c) => [c.name, c.data_type] as const),
    );
  }, [selectedTable]);

  const loadBrowseData = useCallback(async () => {
    if (!selectedTable || browseTab !== "data") return;
    setBrowseBusy(true);
    setBrowseError(null);
    try {
      const res = await runDatabaseQuery(
        serviceId,
        selectTableSQL(selectedTable),
      );
      setBrowseResult(res);
    } catch (err) {
      setBrowseResult(null);
      setBrowseError(
        err instanceof Error ? err.message : "Could not load rows",
      );
    } finally {
      setBrowseBusy(false);
    }
  }, [serviceId, selectedTable, browseTab]);

  useEffect(() => {
    if (mode !== "browse" || browseTab !== "data" || !selectedTable) return;
    void loadBrowseData();
  }, [mode, browseTab, selectedTable, loadBrowseData]);

  function selectTable(t: DatabaseSchemaTable) {
    setSelectedTable(t);
    setMode("browse");
    setBrowseTab("data");
  }

  async function onRunConsole(e: React.FormEvent) {
    e.preventDefault();
    setConsoleBusy(true);
    setConsoleError(null);
    try {
      const res = await runDatabaseQuery(serviceId, sql);
      setConsoleResult(res);
    } catch (err) {
      setConsoleResult(null);
      setConsoleError(err instanceof Error ? err.message : "Query failed");
    } finally {
      setConsoleBusy(false);
    }
  }

  return (
    <section
      className={RESOURCE_DETAIL_CARD}
      style={resourceDetailCardStyle()}
      aria-label="Query studio"
    >
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">Query studio</h2>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            Browse tables or run read-only SQL. Mutations are blocked on the
            server.
          </p>
        </div>
        <fieldset
          className="inline-flex rounded-lg border-0 p-0.5"
          style={{ background: "var(--surface-muted, #f4f4f5)" }}
          aria-label="Studio mode"
        >
          <button
            type="button"
            className="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
            style={{
              background: mode === "browse" ? "var(--ink-navy)" : "transparent",
              color: mode === "browse" ? "#fff" : "var(--muted-text)",
            }}
            onClick={() => setMode("browse")}
          >
            Schema
          </button>
          <button
            type="button"
            className="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
            style={{
              background:
                mode === "console" ? "var(--ink-navy)" : "transparent",
              color: mode === "console" ? "#fff" : "var(--muted-text)",
            }}
            onClick={() => setMode("console")}
          >
            SQL console
          </button>
        </fieldset>
      </div>

      <div className="mt-4 flex flex-col gap-6 lg:flex-row">
        <aside
          className="flex w-full shrink-0 flex-col lg:w-60"
          aria-label="Schema browser"
        >
          <div className="flex gap-2">
            <input
              type="search"
              placeholder="Search tables…"
              value={tableSearch}
              onChange={(e) => setTableSearch(e.target.value)}
              className="field-input min-w-0 flex-1 text-[13px]"
              aria-label="Search tables"
            />
            <Button
              type="button"
              variant="outline"
              disabled={schemaLoading}
              onClick={() => void loadSchema()}
              aria-label="Refresh schema"
            >
              ↻
            </Button>
          </div>
          <ul
            className="mt-3 max-h-[min(420px,50vh)] flex-1 space-y-0.5 overflow-y-auto rounded-md border text-[13px]"
            style={{ borderColor: "var(--border)" }}
          >
            {schemaLoading ? (
              <li className="px-3 py-4" style={{ color: "var(--muted-text)" }}>
                Loading…
              </li>
            ) : filteredTables.length === 0 ? (
              <li className="px-3 py-4" style={{ color: "var(--muted-text)" }}>
                {tables.length === 0 ? "No tables found" : "No matches"}
              </li>
            ) : (
              filteredTables.map((t) => {
                const active =
                  selectedTable && tableKey(selectedTable) === tableKey(t);
                return (
                  <li key={tableKey(t)}>
                    <button
                      type="button"
                      className="flex w-full items-center gap-2 px-3 py-2 text-left transition-colors"
                      style={{
                        background: active ? "var(--ink-navy)" : "transparent",
                        color: active ? "#fff" : "var(--ink-navy)",
                      }}
                      onClick={() => selectTable(t)}
                    >
                      <span aria-hidden className="opacity-60">
                        ⊞
                      </span>
                      <span className="truncate font-mono">
                        {t.schema === "public" ? t.name : tableKey(t)}
                      </span>
                    </button>
                  </li>
                );
              })
            )}
          </ul>
        </aside>

        <div className="min-w-0 flex-1">
          {mode === "console" ? (
            <>
              <form
                className="space-y-4"
                onSubmit={(e) => void onRunConsole(e)}
              >
                <Field label="SQL">
                  <textarea
                    id="database-query-sql"
                    rows={8}
                    value={sql}
                    onChange={(e) => setSql(e.target.value)}
                    className="field-input font-mono text-[13px]"
                  />
                </Field>
                <Button type="submit" disabled={consoleBusy || !sql.trim()}>
                  {consoleBusy ? "Running…" : "Run query"}
                </Button>
              </form>
              {consoleError ? (
                <p className="mt-4 text-sm text-red-600" role="alert">
                  {consoleError}
                </p>
              ) : null}
              {consoleResult ? (
                <div className="mt-6">
                  <QueryResultTable result={consoleResult} />
                </div>
              ) : null}
            </>
          ) : selectedTable ? (
            <>
              <div className="flex flex-wrap items-center justify-between gap-3">
                <p className="font-mono text-sm font-semibold">
                  {tableKey(selectedTable)}
                </p>
                <div className="flex items-center gap-2">
                  {browseTab === "data" ? (
                    <Button
                      type="button"
                      variant="outline"
                      disabled={browseBusy}
                      onClick={() => void loadBrowseData()}
                    >
                      Refresh rows
                    </Button>
                  ) : null}
                </div>
              </div>
              <div className="mt-4">
                <Tabs
                  tabs={[...BROWSE_TABS]}
                  activeId={browseTab}
                  onChange={(id) => setBrowseTab(id as BrowseTab)}
                />
              </div>
              <div className="mt-4">
                {browseTab === "structure" ? (
                  <table className="w-full border-collapse text-left text-[13px]">
                    <thead>
                      <tr style={{ borderBottom: "1px solid var(--border)" }}>
                        <th className="px-2 py-2 font-semibold">Column</th>
                        <th className="px-2 py-2 font-semibold">Type</th>
                      </tr>
                    </thead>
                    <tbody>
                      {selectedTable.columns.map((col) => (
                        <tr
                          key={col.name}
                          style={{ borderBottom: "1px solid var(--border)" }}
                        >
                          <td className="px-2 py-2 font-mono">{col.name}</td>
                          <td
                            className="px-2 py-2 font-mono"
                            style={{ color: "var(--muted-text)" }}
                          >
                            {col.data_type}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                ) : browseBusy && !browseResult ? (
                  <p style={{ color: "var(--muted-text)" }}>Loading rows…</p>
                ) : browseError ? (
                  <p className="text-sm text-red-600" role="alert">
                    {browseError}
                  </p>
                ) : browseResult ? (
                  <QueryResultTable
                    result={browseResult}
                    columnTypes={columnTypeMap}
                  />
                ) : null}
              </div>
            </>
          ) : (
            <p className="py-8 text-sm" style={{ color: "var(--muted-text)" }}>
              Select a table from the list to browse data and structure.
            </p>
          )}
        </div>
      </div>
    </section>
  );
}
