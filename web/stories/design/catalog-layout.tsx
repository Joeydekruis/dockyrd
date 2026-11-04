import type { ReactNode } from "react";

export function CatalogCanvas({ children }: { children: ReactNode }) {
  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-8 p-8">{children}</div>
  );
}

export function CatalogRow({ children }: { children: ReactNode }) {
  return <div className="flex flex-wrap items-center gap-3">{children}</div>;
}

export function CatalogGrid({ children }: { children: ReactNode }) {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">{children}</div>
  );
}
