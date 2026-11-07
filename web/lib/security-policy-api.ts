import { apiFetch } from "@/lib/api";

export type HeaderEntry = { enabled: boolean; value: string };

export type SecurityPolicy = {
  default_profile: string;
  lock: boolean;
  headers: Record<string, HeaderEntry>;
  header_catalog: string[];
  rl_default_rpm?: number;
  rl_login_rpm?: number;
  cdn_mode: string;
  cdn_ranges: string[];
  cdn_header: string;
  default_mem_mb: number;
  default_cpu: number;
  default_pids: number;
  proxy_rate_limit_installed: boolean;
  proxy_waf_installed: boolean;
};

export type ServiceSecurityStored = {
  profile: string;
  lock_exempt: boolean;
  csp?: string;
  csp_report_only: boolean;
  basic_user?: string;
  basic_auth_configured: boolean;
  ip_allowlist: string[];
  maintenance: boolean;
  rl_mode: string;
  waf_mode: string;
  mem_mb?: number;
  cpu?: number;
  pids?: number;
};

export type ServiceSecurityPayload = {
  stored: ServiceSecurityStored;
  effective: {
    profile: string;
    headers: Record<string, string>;
    lock_floor: boolean;
    instance_default_profile: string;
    mem_mb: number;
    cpu: number;
    pids: number;
    proxy_rate_limit_installed: boolean;
    proxy_waf_installed: boolean;
  };
};

export function getSecurityPolicy() {
  return apiFetch<SecurityPolicy>("/api/v1/security/policy");
}

export function putSecurityPolicy(
  body: Partial<SecurityPolicy> & { headers: Record<string, HeaderEntry> },
) {
  return apiFetch<SecurityPolicy>("/api/v1/security/policy", {
    method: "PUT",
    json: body,
  });
}

export function getServiceSecurity(serviceId: string) {
  return apiFetch<ServiceSecurityPayload>(
    `/api/v1/services/${serviceId}/security`,
  );
}

export function putServiceSecurity(
  serviceId: string,
  body: Record<string, unknown>,
) {
  return apiFetch<ServiceSecurityPayload>(
    `/api/v1/services/${serviceId}/security`,
    {
      method: "PUT",
      json: body,
    },
  );
}

export function setServiceBasicAuth(serviceId: string, username: string) {
  return apiFetch<{ password: string }>(
    `/api/v1/services/${serviceId}/security/basic-auth`,
    {
      method: "POST",
      json: { username },
    },
  );
}

export function clearServiceBasicAuth(serviceId: string) {
  return apiFetch<void>(`/api/v1/services/${serviceId}/security/basic-auth`, {
    method: "DELETE",
  });
}

export function listServiceSecurityEvents(serviceId: string) {
  return apiFetch<{ events: unknown[] }>(
    `/api/v1/services/${serviceId}/security/events`,
  );
}
