import type { ReactNode } from "react";
import { DockyrdLogo } from "@/components/dockyard/dockyrd-logo";

type AuthLayoutProps = {
  children: ReactNode;
  title: string;
};

export function AuthLayout({ children, title }: AuthLayoutProps) {
  return (
    <div className="flex min-h-dvh">
      <div
        className="hidden w-1/2 flex-col justify-between p-12 text-white lg:flex"
        style={{ background: "var(--hero-gradient)" }}
      >
        <DockyrdLogo variant="dark" width={170} />
        <p className="max-w-md text-sm text-white/90">
          Self-hosted deployments with a calm, reliable control plane.
        </p>
      </div>
      <main className="flex flex-1 flex-col justify-center px-6 py-12 lg:px-16">
        <div className="mx-auto w-full max-w-md">
          <h1 className="text-3xl font-semibold tracking-tight">{title}</h1>
          <div className="mt-8">{children}</div>
        </div>
      </main>
    </div>
  );
}
