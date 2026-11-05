"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { formatBytes } from "@/lib/analytics-stats";
import { getBuildBudget, type BuildBudgetSnapshot } from "@/lib/api";

export function BuildCapacityCard() {
  const [budget, setBudget] = useState<BuildBudgetSnapshot | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void getBuildBudget()
      .then(setBudget)
      .catch((e) => {
        setError(e instanceof Error ? e.message : "Could not load build budget");
      });
  }, []);

  if (error) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
        {error}
      </p>
    );
  }
  if (!budget) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
        Loading build capacity…
      </p>
    );
  }

  const total = budget.ram_budget + budget.swap_budget;

  return (
    <article
      className="rounded-[20px] border p-5 sm:col-span-2 xl:col-span-4"
      style={{
        background: "var(--surface)",
        borderColor: "var(--border)",
      }}
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Build capacity
          </p>
          <p className="mt-2 text-[28px] font-semibold leading-none tabular-nums">
            {formatBytes(total)}
          </p>
          <p className="mt-2 text-[13px]" style={{ color: "var(--muted-text)" }}>
            RAM {formatBytes(budget.ram_budget)} + swap{" "}
            {formatBytes(budget.swap_budget)} per build · server{" "}
            {formatBytes(budget.mem_total)}
            {budget.isolation_active ? "" : " · isolation inactive"}
          </p>
        </div>
        <Link
          href="/settings/server/builders"
          className="text-sm underline"
          style={{ color: "var(--link-blue)" }}
        >
          Remote builders
        </Link>
      </div>
    </article>
  );
}
