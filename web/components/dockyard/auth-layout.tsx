import type { ReactNode } from "react";
import { DockyrdLogo } from "@/components/dockyard/dockyrd-logo";

type AuthLayoutProps = {
  children: ReactNode;
  title: string;
  subtitle?: string;
};

export function AuthLayout({ children, title, subtitle }: AuthLayoutProps) {
  return (
    <div className="flex min-h-dvh">
      <div
        className="hidden w-1/2 flex-col justify-between p-12 text-white lg:flex"
        style={{ background: "var(--hero-gradient)" }}
      >
        <DockyrdLogo variant="dark" width={170} />
        <p className="max-w-md text-sm leading-relaxed text-white/90">
          Deploy to your own servers. Without the mess — a calm control plane
          for applications, databases, and static sites.
        </p>
      </div>
      <main className="flex flex-1 flex-col justify-center px-6 py-12 lg:px-16">
        <div className="mx-auto w-full max-w-md">
          <div className="mb-8 lg:hidden">
            <DockyrdLogo width={150} />
          </div>
          <h1 className="text-[44px] font-semibold leading-[1.04] tracking-[-0.04em]">
            {title}
          </h1>
          {subtitle ? (
            <p className="mt-2 text-sm" style={{ color: "var(--muted-text)" }}>
              {subtitle}
            </p>
          ) : null}
          <div className="mt-8">{children}</div>
        </div>
      </main>
    </div>
  );
}
