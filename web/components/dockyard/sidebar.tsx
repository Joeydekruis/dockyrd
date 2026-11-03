"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  mainNav,
  projectNav,
  serverNav,
  settingsNav,
} from "@/lib/nav";
import { DockyrdLogo } from "@/components/dockyard/dockyrd-logo";

function NavLink({ href, label }: { href: string; label: string }) {
  const pathname = usePathname();
  const active = pathname === href || (href !== "/" && pathname.startsWith(href));
  return (
    <Link
      href={href}
      aria-current={active ? "page" : undefined}
      className="block rounded-lg px-3 py-2 text-sm font-medium transition-colors"
      style={{
        background: active ? "var(--brand-soft)" : "transparent",
        color: active ? "var(--ink-navy)" : "var(--secondary-text)",
      }}
    >
      {label}
    </Link>
  );
}

export function Sidebar() {
  return (
    <aside
      className="flex w-64 shrink-0 flex-col border-r px-4 py-6"
      style={{ background: "var(--page-bg)", borderColor: "var(--border)" }}
    >
      <Link href="/" className="mb-8 inline-block" aria-label="Dockyrd home">
        <DockyrdLogo />
      </Link>
      <nav aria-label="Main" className="flex flex-col gap-1">
        {mainNav.map((item) => (
          <NavLink key={item.href} href={item.href} label={item.label} />
        ))}
      </nav>
      <div
        role="group"
        aria-label="Project sections"
        className="mt-6 flex flex-col gap-1"
      >
        {projectNav.map((item) => (
          <NavLink key={item.href} href={item.href} label={item.label} />
        ))}
      </div>
      <p
        className="mt-6 mb-2 font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        Server
      </p>
      <nav className="flex flex-col gap-1">
        {serverNav.map((item) => (
          <NavLink key={item.href} href={item.href} label={item.label} />
        ))}
      </nav>
      <div className="mt-6">
        <NavLink href={settingsNav.href} label={settingsNav.label} />
      </div>
      <div
        className="mt-auto rounded-[20px] border p-4 text-xs"
        style={{ borderColor: "var(--border)", background: "var(--surface)" }}
      >
        <div className="flex items-center gap-2">
          <span
            className="inline-block h-2 w-2 rounded-full"
            style={{ background: "var(--status-ok-dot)" }}
            aria-hidden
          />
          <span className="font-medium">Server healthy</span>
        </div>
        <p className="mt-1 font-mono text-[11px]" style={{ color: "var(--muted-text)" }}>
          v0.0.3-dev
        </p>
      </div>
    </aside>
  );
}
