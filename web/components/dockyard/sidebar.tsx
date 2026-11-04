"use client";

import type { LucideIcon } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { DockyrdLogo } from "@/components/dockyard/dockyrd-logo";
import {
  accountSettingsBackNav,
  accountSettingsNav,
  isAccountSettingsPath,
  isServerSettingsPath,
  isSettingsPath,
  mainNav,
  projectNav,
  type SettingsNavItem,
  serverNav,
  serverSettingsBackNav,
  serverSettingsNav,
  serverSettingsSidebarNav,
} from "@/lib/nav";

function NavLink({
  href,
  label,
  icon: Icon,
}: {
  href: string;
  label: string;
  icon: LucideIcon;
}) {
  const pathname = usePathname();
  const active =
    pathname === href || (href !== "/" && pathname.startsWith(href));
  return (
    <Link
      href={href}
      aria-current={active ? "page" : undefined}
      className="dockyard-nav-link"
    >
      <Icon strokeWidth={1.8} className="size-[18px] shrink-0" aria-hidden />
      <span>{label}</span>
    </Link>
  );
}

function SettingsNavLink({ href, label, icon: Icon }: SettingsNavItem) {
  const pathname = usePathname();
  const active = pathname === href || pathname.startsWith(`${href}/`);
  return (
    <Link
      href={href}
      aria-current={active ? "page" : undefined}
      className="dockyard-nav-link"
    >
      <Icon strokeWidth={1.8} className="size-[18px] shrink-0" aria-hidden />
      <span>{label}</span>
    </Link>
  );
}

function SettingsSidebarShell({
  back,
  groupLabel,
  items,
}: {
  back: { href: string; label: string; icon: LucideIcon };
  groupLabel: string;
  items: SettingsNavItem[];
}) {
  const BackIcon = back.icon;
  return (
    <nav aria-label="Settings" className="flex flex-col gap-1">
      <Link href={back.href} className="dockyard-nav-link">
        <BackIcon
          strokeWidth={1.8}
          className="size-[18px] shrink-0"
          aria-hidden
        />
        <span>{back.label}</span>
      </Link>
      <p
        className="mb-2 mt-6 font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {groupLabel}
      </p>
      {items.map((item) => (
        <SettingsNavLink key={item.href} {...item} />
      ))}
    </nav>
  );
}

function AccountSettingsSidebar() {
  return (
    <SettingsSidebarShell
      back={accountSettingsBackNav}
      groupLabel="Account"
      items={accountSettingsNav}
    />
  );
}

function ServerSettingsSidebar() {
  return (
    <SettingsSidebarShell
      back={serverSettingsBackNav}
      groupLabel="Server"
      items={serverSettingsSidebarNav}
    />
  );
}

function MainSidebar() {
  return (
    <>
      <nav aria-label="Main" className="flex flex-col gap-1">
        {mainNav.map((item) => (
          <NavLink key={item.href} {...item} />
        ))}
      </nav>
      <fieldset
        aria-label="Project sections"
        className="mb-7 mt-6 m-0 flex flex-col gap-1 border-0 p-0"
      >
        {projectNav.map((item) => (
          <NavLink key={item.href} {...item} />
        ))}
      </fieldset>
      <p
        className="mb-2 font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        Server
      </p>
      <nav className="flex flex-col gap-1">
        {serverNav.map((item) => (
          <NavLink key={item.href} {...item} />
        ))}
        <NavLink {...serverSettingsNav} />
      </nav>
    </>
  );
}

export function Sidebar() {
  const pathname = usePathname();
  const inSettings = isSettingsPath(pathname);

  let settingsSidebar: React.ReactNode = null;
  if (isAccountSettingsPath(pathname)) {
    settingsSidebar = <AccountSettingsSidebar />;
  } else if (isServerSettingsPath(pathname)) {
    settingsSidebar = <ServerSettingsSidebar />;
  }

  return (
    <aside
      className="fixed bottom-0 left-0 z-40 flex w-[232px] shrink-0 flex-col overflow-y-auto border-r px-4 py-6 top-[var(--dockyard-announcement-height,44px)]"
      style={{ background: "var(--page-bg)", borderColor: "var(--border)" }}
    >
      <Link href="/" className="mb-8 inline-block" aria-label="Dockyrd home">
        <DockyrdLogo />
      </Link>
      {inSettings && settingsSidebar ? settingsSidebar : <MainSidebar />}
      {!inSettings ? (
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
          <p
            className="mt-1 font-mono text-[11px]"
            style={{ color: "var(--muted-text)" }}
          >
            v0.0.3-dev
          </p>
        </div>
      ) : null}
    </aside>
  );
}
