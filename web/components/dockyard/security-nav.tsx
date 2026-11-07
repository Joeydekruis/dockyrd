"use client";

import Link from "next/link";

export type SecurityNavTab = "overview" | "addons" | "policy";

export function SecurityNav({ active }: { active: SecurityNavTab }) {
  const linkClass = (tab: SecurityNavTab) =>
    `-mb-px border-b-2 pb-3 text-sm font-medium ${
      active === tab
        ? "border-[var(--ink-navy)] text-[var(--ink-navy)]"
        : "border-transparent text-[var(--muted-text)]"
    }`;
  return (
    <nav
      className="flex gap-6 border-b border-[var(--border)]"
      aria-label="Security sections"
    >
      <Link href="/server/security" className={linkClass("overview")}>
        Overview
      </Link>
      <Link href="/server/security/addons" className={linkClass("addons")}>
        Add-ons
      </Link>
      <Link href="/server/security/policy" className={linkClass("policy")}>
        Policy
      </Link>
    </nav>
  );
}
