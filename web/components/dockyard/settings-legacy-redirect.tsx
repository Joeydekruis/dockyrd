"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

export function SettingsLegacyRedirect({ href }: { href: string }) {
  const router = useRouter();
  useEffect(() => {
    router.replace(href);
  }, [href, router]);
  return (
    <div
      className="flex min-h-dvh items-center justify-center text-sm"
      style={{ color: "var(--muted-text)" }}
    >
      Redirecting…
    </div>
  );
}
