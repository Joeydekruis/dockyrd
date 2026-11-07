import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";
import { DockyrdLogo } from "@/components/dockyrd/dockyrd-logo";
import { DockyrdMark } from "@/components/dockyrd/dockyrd-mark";

type AuthLayoutProps = {
  children: ReactNode;
  title: string;
  subtitle?: string;
  backHref?: string;
  backLabel?: string;
};

export function AuthLayout({
  children,
  title,
  subtitle,
  backHref,
  backLabel = "Back to sign in",
}: AuthLayoutProps) {
  return (
    <div className="flex min-h-dvh">
      <div
        className="auth-brand-panel hidden w-1/2 flex-col justify-between p-12 text-white lg:flex"
        style={{ background: "var(--hero-gradient)" }}
      >
        <DockyrdMark />
        <div className="flex max-w-lg flex-col gap-6 py-8">
          <span className="auth-brand-badge">Open source • Self-hosted</span>
          <h2 className="auth-brand-headline">
            Deploy on your own
            <br />
            hardware. Without the
            <br />
            mess.
          </h2>
        </div>
        <p className="max-w-md text-sm leading-relaxed text-white/90">
          Your server. Your data. One lightweight binary.
        </p>
      </div>
      <main className="flex flex-1 flex-col justify-center px-6 py-12 lg:px-16">
        <div className="mx-auto w-full max-w-md">
          {backHref ? (
            <Link
              href={backHref}
              className="auth-back-link mb-8 inline-flex items-center gap-1.5 text-sm font-medium"
            >
              <ArrowLeft className="size-4 shrink-0" aria-hidden />
              {backLabel}
            </Link>
          ) : (
            <div className="mb-8 lg:hidden">
              <DockyrdLogo width={150} />
            </div>
          )}
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
