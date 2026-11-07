import { apiFetch } from "@/lib/api";

export type SecurityCheck = {
  id: string;
  title: string;
  status: "ok" | "warn" | "fail";
  severity: string;
  detail: string;
  fix: "none" | "host_action" | "link";
  fix_kind?: string;
  fix_href?: string;
  last_run_at: number;
  extra?: unknown;
};

export type SecurityChecksReport = {
  ran_at: number;
  checks: SecurityCheck[];
  cached?: boolean;
  stale?: boolean;
};

export type SecurityAddon = {
  id: string;
  title: string;
  description: string;
  kind: string;
  installable: boolean;
  state: string;
  version?: string;
  last_error?: string;
  coming_soon?: string;
  host_action?: string;
};

export function getSecurityChecks(refresh?: boolean) {
  const q = refresh ? "?refresh=1" : "";
  return apiFetch<SecurityChecksReport>(`/api/v1/security/checks${q}`);
}

export function runSecurityChecks() {
  return apiFetch<SecurityChecksReport>("/api/v1/security/checks/run", {
    method: "POST",
  });
}

export function fixSecurityCheck(id: string) {
  return apiFetch<{
    host_action_id: string;
    state: string;
    confirm_deadline?: { Int64: number; Valid: boolean };
  }>(`/api/v1/security/checks/${id}/fix`, { method: "POST" });
}

export function listSecurityAddons() {
  return apiFetch<{ addons: SecurityAddon[] }>("/api/v1/security/addons");
}

export function planSecurityAddon(id: string) {
  return apiFetch<{ id: string; plan: Record<string, unknown> }>(
    `/api/v1/security/addons/${id}/plan`,
    { method: "POST" },
  );
}

export function installSecurityAddon(id: string) {
  return apiFetch<{ addon_id: string; job_id: string; state: string }>(
    `/api/v1/security/addons/${id}/install`,
    { method: "POST" },
  );
}

export function uninstallSecurityAddon(id: string) {
  return apiFetch<{ addon_id: string; state: string }>(
    `/api/v1/security/addons/${id}/uninstall`,
    { method: "POST" },
  );
}

export function confirmHostAction(id: string) {
  return apiFetch<{ ok: boolean; id: string; state: string }>(
    `/api/v1/security/host-actions/${id}/confirm`,
    { method: "POST" },
  );
}

export function revertHostAction(id: string) {
  return apiFetch<{ ok: boolean; id: string; state: string }>(
    `/api/v1/security/host-actions/${id}/revert`,
    { method: "POST" },
  );
}
