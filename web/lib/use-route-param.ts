"use client";

import { usePathname } from "next/navigation";
import { useMemo } from "react";

/** Reads a dynamic segment from the URL (static export uses `_` at build time). */
export function useRouteParam(name: string, indexFromEnd = 0): string {
  const pathname = usePathname();
  return useMemo(() => {
    const parts = pathname.split("/").filter(Boolean);
    if (parts.length === 0) {
      return "";
    }
    if (name === "slug" && parts[0] === "projects" && parts[1]) {
      return parts[1] === "_" ? "" : parts[1];
    }
    const idx = parts.length - 1 - indexFromEnd;
    if (idx < 0) {
      return "";
    }
    const val = parts[idx];
    return val === "_" ? "" : val;
  }, [pathname, name, indexFromEnd]);
}
