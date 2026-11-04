export const RESOURCE_DETAIL_CARD = "rounded-[20px] border px-6 py-5";

export function resourceDetailCardStyle() {
  return {
    borderColor: "var(--border)",
    background: "var(--surface)",
  } as const;
}

export function readInitialDetailTab<T extends string>(
  allowed: readonly T[],
  fallback: T,
): T {
  if (typeof window === "undefined") return fallback;
  const t = new URLSearchParams(window.location.search).get("tab");
  if (t && (allowed as readonly string[]).includes(t)) return t as T;
  return fallback;
}

export function persistDetailTab(tab: string) {
  const url = new URL(window.location.href);
  url.searchParams.set("tab", tab);
  window.history.replaceState(
    null,
    "",
    `${url.pathname}?${url.searchParams.toString()}`,
  );
}
