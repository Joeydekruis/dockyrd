const apiBase = "";

type FetchOptions = RequestInit & { json?: unknown };

export async function apiFetch<T>(
  path: string,
  options: FetchOptions = {},
): Promise<T> {
  const { json, headers, ...rest } = options;
  const res = await fetch(`${apiBase}${path}`, {
    ...rest,
    credentials: "same-origin",
    headers: {
      ...(json ? { "Content-Type": "application/json" } : {}),
      ...headers,
    },
    body: json ? JSON.stringify(json) : rest.body,
  });
  if (!res.ok) {
    const err = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(err.error ?? res.statusText);
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return res.json() as Promise<T>;
}

export type MeResponse = {
  id: string;
  email: string;
  name: string;
  role: string;
  totp_enabled?: boolean;
};

export type AuthProbeState =
  | { status: "authenticated"; user: MeResponse }
  | { status: "mfa_pending" }
  | { status: "anonymous" };

export type SetupStatusResponse = {
  complete: boolean;
};

export async function probeAuthState(): Promise<AuthProbeState> {
  const res = await fetch(`${apiBase}/api/v1/auth/me`, {
    credentials: "same-origin",
  });
  if (res.ok) {
    const user = (await res.json()) as MeResponse;
    return { status: "authenticated", user };
  }
  const body = (await res.json().catch(() => ({}))) as {
    mfa_required?: boolean;
  };
  if (body.mfa_required) {
    return { status: "mfa_pending" };
  }
  return { status: "anonymous" };
}

export function getMe() {
  return apiFetch<MeResponse>("/api/v1/auth/me");
}

export function patchAccount(input: { name?: string; email?: string }) {
  return apiFetch<MeResponse>("/api/v1/auth/account", {
    method: "PATCH",
    json: input,
  });
}

export function changePassword(input: {
  current_password: string;
  new_password: string;
}) {
  return apiFetch<{ ok: boolean }>("/api/v1/auth/password/change", {
    method: "POST",
    json: input,
  });
}

export type NotificationPrefs = {
  deployments: boolean;
  security_alerts: boolean;
  product_updates: boolean;
};

export function getNotificationPrefs() {
  return apiFetch<NotificationPrefs>("/api/v1/auth/notifications");
}

export function patchNotificationPrefs(prefs: NotificationPrefs) {
  return apiFetch<NotificationPrefs>("/api/v1/auth/notifications", {
    method: "PATCH",
    json: prefs,
  });
}

export type UserLocalePrefsApi = {
  timezone: string;
  time_format: "24h" | "12h";
};

export function getLocalePrefs() {
  return apiFetch<UserLocalePrefsApi>("/api/v1/auth/locale");
}

export function patchLocalePrefs(prefs: UserLocalePrefsApi) {
  return apiFetch<UserLocalePrefsApi>("/api/v1/auth/locale", {
    method: "PATCH",
    json: prefs,
  });
}

export function getSetupStatus() {
  return apiFetch<SetupStatusResponse>("/api/v1/setup/status");
}

export function login(email: string, password: string) {
  return apiFetch<{ ok?: boolean; mfa_required?: boolean }>(
    "/api/v1/auth/login",
    {
      method: "POST",
      json: { email, password, remember: true },
    },
  );
}

export function submitMfa(code: string, recoveryCode?: string) {
  return apiFetch<{ ok: boolean }>("/api/v1/auth/mfa", {
    method: "POST",
    json: recoveryCode ? { recovery_code: recoveryCode } : { code },
  });
}

export function getAuthSecurity() {
  return apiFetch<{ totp_enabled: boolean; recovery_remaining: number }>(
    "/api/v1/auth/security",
  );
}

export function setupTotp() {
  return apiFetch<{ secret: string; otpauth_url: string }>(
    "/api/v1/auth/totp/setup",
    {
      method: "POST",
      json: {},
    },
  );
}

export function enableTotp(input: { secret: string; code: string }) {
  return apiFetch<{ ok: boolean; recovery_codes?: string[] }>(
    "/api/v1/auth/totp/enable",
    {
      method: "POST",
      json: input,
    },
  );
}

export type AuthSession = {
  id: string;
  current: boolean;
  user_agent?: string;
  ip?: string;
  mfa_passed: boolean;
  created_at: number;
  last_seen_at: number;
  expires_at: number;
};

export function listAuthSessions() {
  return apiFetch<{ sessions: AuthSession[] }>("/api/v1/auth/sessions");
}

export function revokeAuthSession(id: string) {
  return apiFetch<void>(`/api/v1/auth/sessions/${id}`, { method: "DELETE" });
}

export type APITokenEntry = {
  id: string;
  name: string;
  scope: string;
  project_id: string;
  last_used_at?: number;
  last_used_ip?: string;
  created_at: number;
  token_prefix?: string;
};

export function listAPITokens() {
  return apiFetch<{ tokens: APITokenEntry[] }>("/api/v1/tokens");
}

export function createAPIToken(body: {
  name: string;
  scope: string;
  project_id: string;
}) {
  return apiFetch<{ token: string; entry: APITokenEntry }>("/api/v1/tokens", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function revokeAPIToken(id: string) {
  return apiFetch<void>(`/api/v1/tokens/${id}`, { method: "DELETE" });
}

export type AuditEntry = {
  id: string;
  user_id?: string;
  action: string;
  target_kind?: string;
  target_id?: string;
  created_at: number;
};

export function listAuditLog() {
  return apiFetch<{ entries: AuditEntry[] }>("/api/v1/audit");
}

export function logout() {
  return apiFetch<void>("/api/v1/auth/logout", { method: "POST" });
}

export function completeSetup(input: {
  token: string;
  name: string;
  email: string;
  password: string;
}) {
  return apiFetch<{ ok: boolean }>("/api/v1/setup", {
    method: "POST",
    json: input,
  });
}

export type Project = {
  id: string;
  server_id: string;
  name: string;
  slug: string;
  description?: string;
  links_mode?: "open" | "strict";
  created_at: number;
  updated_at: number;
};

export function listProjects() {
  return apiFetch<{ projects: Project[] }>("/api/v1/projects");
}

export function createProject(input: { name: string; description?: string }) {
  return apiFetch<Project>("/api/v1/projects", {
    method: "POST",
    json: input,
  });
}

export function patchProject(
  projectId: string,
  input: {
    name?: string;
    description?: string;
    links_mode?: "open" | "strict";
  },
) {
  return apiFetch<Project>(`/api/v1/projects/${projectId}`, {
    method: "PATCH",
    json: input,
  });
}

export type Service = {
  id: string;
  environment_id: string;
  kind: string;
  name: string;
  slug: string;
  desired_state: string;
  config: Record<string, unknown>;
  created_at: number;
  updated_at: number;
};

export function listServices(projectId: string, kind?: string) {
  const q = new URLSearchParams({ project_id: projectId });
  if (kind) q.set("kind", kind);
  return apiFetch<{ services: Service[] }>(`/api/v1/services?${q}`);
}

export function createService(input: {
  project_id: string;
  environment?: string;
  kind: "app" | "static_site" | "postgres" | "redis" | "bucket" | "compose";
  name: string;
  config?: Record<string, unknown>;
}) {
  return apiFetch<Service>("/api/v1/services", {
    method: "POST",
    json: { environment: "production", ...input },
  });
}

export function moveServiceToProject(serviceId: string, projectId: string) {
  return patchService(serviceId, { project_id: projectId });
}

export function getService(serviceId: string) {
  return apiFetch<Service>(`/api/v1/services/${serviceId}`);
}

export function patchService(
  serviceId: string,
  input: {
    project_id?: string;
    name?: string;
    config?: Record<string, unknown>;
  },
) {
  return apiFetch<Service>(`/api/v1/services/${serviceId}`, {
    method: "PATCH",
    json: input,
  });
}

export type DatabaseRow = {
  id: string;
  kind: string;
  name: string;
  slug: string;
  desired_state: string;
  config: Record<string, unknown>;
  project_id: string;
  project_name: string;
  project_slug: string;
};

export function listDatabases() {
  return apiFetch<{ databases: DatabaseRow[] }>("/api/v1/databases");
}

export function getServiceConnection(serviceId: string, reveal = false) {
  const q = reveal ? "?reveal=1" : "";
  return apiFetch<ServiceConnectionResponse>(
    `/api/v1/services/${serviceId}/connection${q}`,
  );
}

export type ServiceConnectionEndpoint = {
  enabled?: boolean;
  host?: string;
  port?: number;
  variables: Record<string, string>;
};

export type ServiceConnectionResponse = {
  variables: Record<string, string>;
  internal: ServiceConnectionEndpoint;
  external: ServiceConnectionEndpoint;
};

export function patchDatabasePublicAccess(
  serviceId: string,
  input: { enabled: boolean; port?: number },
) {
  return apiFetch<{ enabled: boolean; port: number }>(
    `/api/v1/services/${serviceId}/public-access`,
    { method: "PATCH", json: input },
  );
}

export function deleteDatabaseService(
  serviceId: string,
  input: { confirm_slug: string; delete_volume?: boolean },
) {
  return apiFetch<void>(`/api/v1/services/${serviceId}`, {
    method: "DELETE",
    json: {
      confirm_slug: input.confirm_slug,
      delete_volume: input.delete_volume ?? true,
    },
  });
}

export function deletePanelService(
  serviceId: string,
  input: { confirm_slug: string; delete_volume?: boolean },
) {
  return deleteDatabaseService(serviceId, input);
}

export function getServiceDeployTokenStatus(serviceId: string) {
  return apiFetch<{ configured: boolean; service_id: string }>(
    `/api/v1/services/${serviceId}/deploy-token`,
  );
}

export function createServiceDeployToken(serviceId: string) {
  return apiFetch<{ token: string; scope: string; service_id: string }>(
    `/api/v1/services/${serviceId}/deploy-token`,
    { method: "POST", json: {} },
  );
}

export function getServiceGitHubWorkflow(serviceId: string) {
  return apiFetch<{
    workflow: string;
    yaml: string;
    secrets: { name: string; value?: string; description?: string }[];
  }>(`/api/v1/services/${serviceId}/github-workflow`);
}

export function restartDatabaseService(serviceId: string) {
  return apiFetch<void>(`/api/v1/services/${serviceId}/restart`, {
    method: "POST",
  });
}

export type DatabaseBackupPolicy = {
  enabled: boolean;
  interval: "hourly" | "daily" | "weekly";
  retention_days: number;
  local_enabled: boolean;
  remote_enabled: boolean;
  run_at_minute?: number;
  run_weekday?: number;
  last_run_at?: number;
  next_run_at?: number;
};

export function getDatabaseBackupPolicy(serviceId: string) {
  return apiFetch<DatabaseBackupPolicy>(
    `/api/v1/services/${serviceId}/backup-policy`,
  );
}

export function patchDatabaseBackupPolicy(
  serviceId: string,
  input: Partial<DatabaseBackupPolicy>,
) {
  return apiFetch<DatabaseBackupPolicy>(
    `/api/v1/services/${serviceId}/backup-policy`,
    { method: "PATCH", json: input },
  );
}

export function triggerDatabaseBackup(serviceId: string) {
  return apiFetch<{ queued: boolean }>(
    `/api/v1/services/${serviceId}/backup-runs`,
    { method: "POST" },
  );
}

export function listDatabaseBackupRuns(serviceId: string) {
  return apiFetch<{ runs: DatabaseBackupRun[] }>(
    `/api/v1/services/${serviceId}/backup-runs`,
  );
}

export type DatabaseBackupRun = {
  id: string;
  status: string;
  destination: string;
  object_key?: string;
  size_bytes?: number;
  sha256?: string;
  verified?: boolean;
  verified_at?: number;
  started_at?: number;
  finished_at?: number;
  error?: string;
};

export type DatabaseRestoreRun = {
  id: string;
  backup_run_id: string;
  mode: string;
  status: string;
  pre_snapshot_run_id?: string;
  log_tail?: string;
  error?: string;
  created_at: number;
  started_at?: number;
  finished_at?: number;
};

export type DatabaseQueryResult = {
  columns: string[];
  rows: string[][];
  row_count: number;
  truncated: boolean;
  duration_ms: number;
};

export function runDatabaseQuery(
  serviceId: string,
  sql: string,
  maxRows = 200,
) {
  return apiFetch<DatabaseQueryResult>(`/api/v1/services/${serviceId}/query`, {
    method: "POST",
    json: { sql, max_rows: maxRows },
  });
}

export type DatabaseSchemaTable = {
  schema: string;
  name: string;
  columns: { name: string; data_type: string }[];
};

export function getDatabaseSchema(serviceId: string) {
  return apiFetch<{ tables: DatabaseSchemaTable[] }>(
    `/api/v1/services/${serviceId}/schema`,
  );
}

export function restoreDatabaseBackup(
  serviceId: string,
  runId: string,
  input: { mode?: "in_place"; confirm: string },
) {
  return apiFetch<{ restore_id: string }>(
    `/api/v1/services/${serviceId}/backup-runs/${runId}/restore`,
    {
      method: "POST",
      json: { mode: input.mode ?? "in_place", confirm: input.confirm },
    },
  );
}

export function listDatabaseRestoreRuns(serviceId: string) {
  return apiFetch<{ runs: DatabaseRestoreRun[] }>(
    `/api/v1/services/${serviceId}/restore-runs`,
  );
}

export function getDatabaseRestoreRun(serviceId: string, restoreId: string) {
  return apiFetch<DatabaseRestoreRun>(
    `/api/v1/services/${serviceId}/restore-runs/${restoreId}`,
  );
}

export type ServerBackupTarget = {
  configured: boolean;
  provider: string;
  endpoint: string;
  bucket: string;
  prefix: string;
  region: string;
  has_keys: boolean;
  managed_garage?: boolean;
  last_test_at?: number;
  last_test_error?: string;
};

export function getServerBackupTarget() {
  return apiFetch<ServerBackupTarget>("/api/v1/settings/server/backup-target");
}

export function patchServerBackupTarget(input: {
  endpoint?: string;
  bucket: string;
  prefix?: string;
  region?: string;
  access_key?: string;
  secret_key?: string;
}) {
  return apiFetch<ServerBackupTarget>("/api/v1/settings/server/backup-target", {
    method: "PATCH",
    json: input,
  });
}

export function configureGarageBackupTarget() {
  return apiFetch<ServerBackupTarget>(
    "/api/v1/settings/server/backup-target/garage",
    { method: "POST" },
  );
}

export function testServerBackupTarget() {
  return apiFetch<ServerBackupTarget>(
    "/api/v1/settings/server/backup-target/test",
    { method: "POST" },
  );
}

export async function fetchServiceLogs(
  serviceId: string,
  tail = "200",
  stream: "runtime" | "access" = "runtime",
  composeService?: string,
) {
  const params = new URLSearchParams({ tail, stream });
  if (composeService?.trim()) {
    params.set("compose_service", composeService.trim());
  }
  const res = await fetch(
    `/api/v1/services/${serviceId}/logs?${params.toString()}`,
    { credentials: "same-origin" },
  );
  if (!res.ok) {
    const err = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(err.error ?? res.statusText);
  }
  return res.text();
}

export type ServiceAnalyticsBucket = {
  ts: number;
  requests: number;
  bytes: number;
  status_2xx: number;
  status_3xx: number;
  status_4xx: number;
  status_5xx: number;
  latency_avg_ms: number;
  latency_p95_ms: number;
};

export type ServiceAnalytics = {
  range: string;
  requests_total: number;
  bytes_total: number;
  status_2xx: number;
  status_3xx: number;
  status_4xx: number;
  status_5xx: number;
  series: ServiceAnalyticsBucket[];
};

export function getServiceAnalytics(
  serviceId: string,
  range: "1h" | "24h" | "7d" | "30d" = "24h",
) {
  const q = new URLSearchParams({ range });
  return apiFetch<ServiceAnalytics>(
    `/api/v1/services/${serviceId}/analytics?${q}`,
  );
}

export type StorageStatus = {
  installed: boolean;
  provider?: string;
  version?: string;
  internal_endpoint: string;
  buckets?: {
    id: string;
    name: string;
    slug: string;
    project_id: string;
    project_name: string;
    config: Record<string, unknown>;
  }[];
};

export function getStorageStatus() {
  return apiFetch<StorageStatus>("/api/v1/storage");
}

export function installGarage() {
  return apiFetch<StorageStatus>("/api/v1/storage/install", {
    method: "POST",
    json: { provider: "garage" },
  });
}

export type BucketKey = {
  id: string;
  name?: string;
  access_key_id: string;
  secret?: string;
};

export function listBucketKeys(serviceId: string, reveal = false) {
  const q = reveal ? "?reveal=1" : "";
  return apiFetch<{ keys: BucketKey[] }>(
    `/api/v1/services/${serviceId}/keys${q}`,
  );
}

export type Deployment = {
  id: string;
  service_id: string;
  status: string;
  trigger: string;
  image_ref?: string;
  release_path?: string;
  commit_sha?: string;
  commit_message?: string;
  commit_author?: string;
  branch?: string;
  build_log_path?: string;
  error?: string;
  mem_peak_bytes?: number;
  swap_peak_bytes?: number;
  cpu_seconds?: number;
  waited_ms?: number;
  frozen_ms?: number;
  budget_bytes?: number;
  builder?: string;
  failure_kind?: string;
  queued_at: number;
  started_at?: number;
  finished_at?: number;
};

export type BuildBudgetSnapshot = {
  ram_budget: number;
  swap_budget: number;
  apps_ws: number;
  system_ws: number;
  kernel_reserve: number;
  mem_total: number;
  isolation_active: boolean;
};

export type RemoteBuilder = {
  id: string;
  name: string;
  kind: string;
  endpoint: string;
  platforms?: string | null;
  status?: string | null;
  last_seen_at?: number | null;
  created_at: number;
};

export function getBuildBudget() {
  return apiFetch<BuildBudgetSnapshot>("/api/v1/server/build-budget");
}

export function listBuilders() {
  return apiFetch<{ builders: RemoteBuilder[] }>("/api/v1/builders");
}

export function createRemoteBuilder(body: { name: string; endpoint: string }) {
  return apiFetch<{
    id: string;
    name: string;
    endpoint: string;
    docker_run: string;
    server_certs: Record<string, string>;
  }>("/api/v1/builders", { method: "POST", json: body });
}

export function deleteRemoteBuilder(id: string) {
  return apiFetch<void>(`/api/v1/builders/${id}`, { method: "DELETE" });
}

export function testRemoteBuilder(id: string) {
  return apiFetch<{ status: string; detail: string }>(
    `/api/v1/builders/${id}/test`,
    { method: "POST", json: {} },
  );
}

export function patchServiceBuildLocation(
  serviceId: string,
  buildLocation: string,
) {
  return apiFetch<{ build_location: string }>(
    `/api/v1/services/${serviceId}/build-location`,
    {
      method: "PATCH",
      json: { build_location: buildLocation },
    },
  );
}

export type ServicePreviewConfig = {
  enabled: boolean;
};

export type ServicePreview = {
  id: string;
  name: string;
  slug: string;
  preview_pr: number;
  preview_head_sha?: string;
  preview_last_push_at?: number;
  primary_hostname?: string;
  live_deployment_id?: string;
};

export function listServicePreviews(serviceId: string) {
  return apiFetch<{ previews: ServicePreview[] }>(
    `/api/v1/services/${serviceId}/previews`,
  );
}

export function deleteServicePreview(serviceId: string, previewId: string) {
  return apiFetch<void>(`/api/v1/services/${serviceId}/previews/${previewId}`, {
    method: "DELETE",
  });
}

export function patchServicePreviewConfig(
  serviceId: string,
  config: ServicePreviewConfig,
) {
  return apiFetch<{ preview: ServicePreviewConfig }>(
    `/api/v1/services/${serviceId}/preview-config`,
    { method: "PATCH", json: config },
  );
}

export function listDeployments(serviceId: string) {
  return apiFetch<{ deployments: Deployment[] }>(
    `/api/v1/services/${serviceId}/deployments`,
  );
}

export function createDeployment(serviceId: string) {
  return apiFetch<Deployment>(`/api/v1/services/${serviceId}/deployments`, {
    method: "POST",
    json: {},
  });
}

export async function fetchDeploymentLog(
  deploymentId: string,
): Promise<string> {
  const res = await fetch(`/api/v1/deployments/${deploymentId}/log`, {
    credentials: "same-origin",
  });
  if (!res.ok) {
    return "";
  }
  return res.text();
}

export async function getDeploymentForService(
  serviceId: string,
  deploymentId: string,
): Promise<Deployment | null> {
  const { deployments } = await listDeployments(serviceId);
  return deployments.find((d) => d.id === deploymentId) ?? null;
}

export function promoteDeployment(deploymentId: string) {
  return apiFetch<Deployment>(`/api/v1/deployments/${deploymentId}/promote`, {
    method: "POST",
    json: {},
  });
}

export type ServiceDomain = {
  id: string;
  service_id: string;
  hostname: string;
  is_platform: boolean;
  is_primary: boolean;
  redirect_to?: string;
  dns_status?: string;
  cert_status?: string;
  target_service?: string;
  target_port?: number;
};

export function listServiceDomains(serviceId: string) {
  return apiFetch<{ domains: ServiceDomain[] }>(
    `/api/v1/services/${serviceId}/domains`,
  );
}

export function createServiceDomain(
  serviceId: string,
  input: {
    hostname?: string;
    platform?: boolean;
    primary?: boolean;
    target_service?: string;
    target_port?: number;
  },
) {
  return apiFetch<ServiceDomain>(`/api/v1/services/${serviceId}/domains`, {
    method: "POST",
    json: input,
  });
}

export function deleteServiceDomain(domainId: string) {
  return apiFetch<void>(`/api/v1/domains/${domainId}`, { method: "DELETE" });
}

export function patchServiceDomain(
  domainId: string,
  body: {
    primary?: true;
    target_service?: string;
    target_port?: number;
  },
) {
  return apiFetch<ServiceDomain>(`/api/v1/domains/${domainId}`, {
    method: "PATCH",
    json: body,
  });
}

export function checkServiceDomain(domainId: string) {
  return apiFetch<ServiceDomain>(`/api/v1/domains/${domainId}/check`, {
    method: "POST",
  });
}

export type ComposeValidateResult = {
  ok: boolean;
  errors: { rule: string; message: string }[];
  warnings: { rule: string; message: string }[];
  services: string[];
  exposable_ports: { service: string; port: number }[];
};

export function validateComposeYaml(serviceId: string, yaml: string) {
  return apiFetch<ComposeValidateResult>(
    `/api/v1/services/${serviceId}/compose/validate`,
    { method: "POST", json: { yaml } },
  );
}

export type ComposeContainerStatus = {
  name: string;
  service: string;
  state: string;
  health?: string;
};

export function listComposeContainers(serviceId: string) {
  return apiFetch<{ containers: ComposeContainerStatus[] }>(
    `/api/v1/services/${serviceId}/compose/containers`,
  );
}

export type ServiceVariable = {
  key: string;
  is_secret: boolean;
  at_build: boolean;
  at_runtime: boolean;
  value?: string;
  link_id?: string;
};

export function listServiceVariables(serviceId: string, reveal = false) {
  const q = reveal ? "?reveal=1" : "";
  return apiFetch<{ variables: ServiceVariable[] }>(
    `/api/v1/services/${serviceId}/variables${q}`,
  );
}

export function putServiceVariables(
  serviceId: string,
  variables: {
    key: string;
    value: string;
    is_secret?: boolean;
    at_build?: boolean;
    at_runtime?: boolean;
  }[],
) {
  return apiFetch<void>(`/api/v1/services/${serviceId}/variables`, {
    method: "PUT",
    json: { variables },
  });
}

export type ServiceLinkPeer = {
  id: string;
  name: string;
  kind: string;
  slug: string;
  status: string;
};

export type ServiceLink = {
  id: string;
  from_service_id: string;
  to_service_id: string;
  peer: ServiceLinkPeer;
  variables_count: number;
  created_at: number;
};

export type LinkableCandidate = {
  service: ServiceLinkPeer;
  suggested_variables: {
    key: string;
    ref: string;
    at_build: boolean;
    at_runtime: boolean;
    default_on: boolean;
    is_secret: boolean;
  }[];
  internal_host: string;
};

export function listServiceLinks(serviceId: string) {
  return apiFetch<{ outgoing: ServiceLink[]; incoming: ServiceLink[] }>(
    `/api/v1/services/${serviceId}/links`,
  );
}

export function listLinkableServices(serviceId: string) {
  return apiFetch<{ candidates: LinkableCandidate[] }>(
    `/api/v1/services/${serviceId}/linkable`,
  );
}

export function createServiceLink(
  serviceId: string,
  body: {
    to_service_id?: string;
    from_service_id?: string;
    variables?: {
      key: string;
      ref: string;
      at_build?: boolean;
      at_runtime?: boolean;
      is_secret?: boolean;
    }[];
  },
) {
  return apiFetch<{ id: string }>(`/api/v1/services/${serviceId}/links`, {
    method: "POST",
    json: body,
  });
}

export function deleteServiceLink(linkId: string, removeVariables = true) {
  const q = removeVariables
    ? "?remove_variables=true"
    : "?remove_variables=false";
  return apiFetch<void>(`/api/v1/links/${linkId}${q}`, { method: "DELETE" });
}

export type ProjectLinkEdge = {
  id: string;
  from_service_id: string;
  to_service_id: string;
  from_name: string;
  to_name: string;
  from_kind: string;
  to_kind: string;
  from_slug: string;
  to_slug: string;
};

export function listProjectLinks(projectId: string) {
  return apiFetch<{
    links: ProjectLinkEdge[];
    outgoing_counts: Record<string, number>;
  }>(`/api/v1/projects/${projectId}/links`);
}

export type GitHubStatus = {
  linked: boolean;
  slug?: string;
  installation_id?: number;
  installed_on?: string;
  repository_count?: number;
  webhook_active?: boolean;
  public_url?: string;
  install_url?: string;
  manage_url?: string;
  app_configured?: boolean;
  installation_ready?: boolean;
  missing_webhook_events?: string[];
  webhook_events_warning?: string;
};

export function getGitHubStatus() {
  return apiFetch<GitHubStatus>("/api/v1/github/status");
}

export function disconnectGitHub() {
  return apiFetch<void>("/api/v1/github/connection", { method: "DELETE" });
}

export type GitHubManifestLaunch = {
  action: string;
  state: string;
  manifest: string;
};

export function startGitHubAppLink(opts?: {
  app_name?: string;
  public_url?: string;
  organization?: string;
}) {
  return apiFetch<GitHubManifestLaunch>("/api/v1/github/app/manifest", {
    method: "POST",
    json: opts ?? {},
  });
}

export type GitHubRepo = {
  full_name: string;
  default_branch: string;
  private: boolean;
};

export function listGitHubRepos(query?: string) {
  const q = query ? `?q=${encodeURIComponent(query)}` : "";
  return apiFetch<{ repos: GitHubRepo[] }>(`/api/v1/github/repos${q}`);
}

export type MetricRange = "1h" | "24h" | "7d" | "30d";

export type HostMetricPoint = {
  ts: number;
  cpu: number;
  mem_used: number;
  mem_total: number;
  disk_used: number;
  disk_total: number;
  net_rx: number;
  net_tx: number;
  load1: number;
};

export type ServerMetricsResponse = {
  range: MetricRange;
  res_sec: number;
  points: HostMetricPoint[];
};

export function getServerMetrics(range: MetricRange = "24h") {
  const q = new URLSearchParams({ range });
  return apiFetch<ServerMetricsResponse>(`/api/v1/server/metrics?${q}`);
}

export type dockyrdEvent = {
  id: string;
  level: string;
  category: string;
  title: string;
  detail?: string;
  project_id?: string;
  service_id?: string;
  deployment_id?: string;
  acknowledged_at?: number;
  created_at: number;
  action_href?: string;
};

export function listEvents(opts?: {
  limit?: number;
  category?: string;
  needsAttention?: boolean;
}) {
  const q = new URLSearchParams();
  if (opts?.limit) q.set("limit", String(opts.limit));
  if (opts?.category) q.set("category", opts.category);
  if (opts?.needsAttention) q.set("needs_attention", "1");
  const suffix = q.size ? `?${q}` : "";
  return apiFetch<{ events: dockyrdEvent[] }>(`/api/v1/events${suffix}`);
}

export function ackEvent(id: string) {
  return apiFetch<void>(`/api/v1/events/${id}/ack`, { method: "POST" });
}

export type TeamUser = {
  id: string;
  email: string;
  name: string;
  role: string;
  created_at: number;
};

export type TeamInvite = {
  id: string;
  email: string;
  role: string;
  expires_at: number;
  token?: string;
  email_sent?: boolean;
};

export function listTeamUsers() {
  return apiFetch<{ users: TeamUser[] }>("/api/v1/users");
}

export function listInvites() {
  return apiFetch<{ invites: TeamInvite[] }>("/api/v1/invites");
}

export function createInvite(input: { email: string; role: string }) {
  return apiFetch<TeamInvite>("/api/v1/invites", {
    method: "POST",
    json: input,
  });
}

export function previewInvite(token: string) {
  const q = new URLSearchParams({ token });
  return apiFetch<{ email: string; role: string; expires_at: number }>(
    `/api/v1/invites/preview?${q}`,
  );
}

export function acceptInvite(input: {
  token: string;
  name: string;
  password: string;
}) {
  return apiFetch<{ ok: boolean }>("/api/v1/invites/accept", {
    method: "POST",
    json: input,
  });
}

export type SearchResult = {
  kind: string;
  label: string;
  href: string;
};

export function globalSearch(q: string) {
  const params = new URLSearchParams({ q });
  return apiFetch<{ results: SearchResult[] }>(`/api/v1/search?${params}`);
}

export type ServerContainer = {
  id: string;
  name: string;
  state: string;
  kind?: string;
  service_id?: string;
  created_at?: number;
  project_label?: string;
  cpu_percent?: number;
  memory_bytes?: number;
  memory_limit?: number;
};

export type CleanupSummary = {
  build_cache_bytes: number;
  dangling_images_bytes: number;
  stopped_unmanaged_containers: number;
  old_releases_bytes: number;
  docker_available: boolean;
};

export function getServerContainers() {
  return apiFetch<{ containers: ServerContainer[] }>(
    "/api/v1/server/containers",
  );
}

export function getServerCleanup() {
  return apiFetch<CleanupSummary>("/api/v1/server/cleanup");
}

export type ServerInfo = {
  version: string;
  uptime_sec: number;
  hostname?: string;
  public_ipv4?: string;
  cpu_count?: number;
  swap_used?: number;
  swap_total?: number;
  cores?: { core: number; cpu: number }[];
  update_available: boolean;
  latest_tag?: string;
  release_notes?: string;
};

export type ContainerMetricSnapshot = {
  service_id: string;
  name?: string;
  project_label?: string;
  ts: number;
  cpu_percent: number;
  memory_bytes: number;
  memory_limit: number;
  net_rx: number;
  net_tx: number;
};

export function getServerContainerMetrics(range: MetricRange = "24h") {
  const q = new URLSearchParams({ range });
  return apiFetch<{ range: MetricRange; metrics: ContainerMetricSnapshot[] }>(
    `/api/v1/server/metrics/containers?${q}`,
  );
}

export function getServerInfo() {
  return apiFetch<ServerInfo>("/api/v1/server");
}

export type UpdatesStatus = {
  version: string;
  update_available: boolean;
  latest_tag?: string;
  release_notes?: string;
  check_ok: boolean;
  check_error?: string;
  checked_at?: number;
  from_cache?: boolean;
  apply_supported: boolean;
  apply_unsupported_reason?: string;
};

export function getUpdatesStatus(options?: { refresh?: boolean }) {
  const query = options?.refresh ? "?refresh=1" : "";
  return apiFetch<UpdatesStatus>(`/api/v1/updates${query}`);
}

export function applydockyrdUpdate() {
  return apiFetch<{
    ok: boolean;
    installed_tag?: string;
    message: string;
    restarting: boolean;
  }>("/api/v1/updates/apply", { method: "POST" });
}

export type PublicSettings = {
  smtp_configured: boolean;
};

export function getPublicSettings() {
  return apiFetch<PublicSettings>("/api/v1/settings/public");
}

export type InstanceSettings = {
  public_url: string;
  acme_email: string;
  wildcard_domain: string;
  wildcard_verified_at?: number;
  wildcard_verified?: boolean;
  wildcard_dns_status?: string;
  wildcard_probe_hostname?: string;
  suggested_apps_domain?: string;
  public_ipv4?: string;
  require_2fa_all: boolean;
  terminal_allow_host?: boolean;
  smtp: {
    host: string;
    port: number;
    username: string;
    from: string;
    use_tls: boolean;
    configured: boolean;
  };
};

export function getInstanceSettings() {
  return apiFetch<InstanceSettings>("/api/v1/settings/instance");
}

export function verifyInstanceWildcard() {
  return apiFetch<{
    dns_status: string;
    wildcard_verified: boolean;
    wildcard_verified_at?: number;
    wildcard_probe_hostname?: string;
    expected_ipv4?: string;
  }>("/api/v1/settings/instance/verify-wildcard", { method: "POST" });
}

export function patchInstanceSettings(body: {
  public_url?: string;
  acme_email?: string;
  wildcard_domain?: string;
  require_2fa_all?: boolean;
  terminal_allow_host?: boolean;
  terminal_host_mfa_code?: string;
  smtp?: {
    host: string;
    port: number;
    username: string;
    from: string;
    use_tls: boolean;
    password?: string;
  };
}) {
  return apiFetch<InstanceSettings>("/api/v1/settings/instance", {
    method: "PATCH",
    json: body,
  });
}

export function testSMTP(to: string) {
  return apiFetch<{ ok: boolean }>("/api/v1/settings/smtp/test", {
    method: "POST",
    json: { to },
  });
}

export type EmailLogEntry = {
  id: string;
  kind: string;
  from: string;
  to: string;
  subject: string;
  status: "sent" | "failed";
  response: string;
  detail?: Record<string, unknown>;
  created_at: number;
  created_at_iso: string;
};

export function listEmailLog(limit = 100) {
  return apiFetch<{ entries: EmailLogEntry[] }>(
    `/api/v1/settings/email/log?limit=${limit}`,
  );
}

export function getEmailLog(id: string) {
  return apiFetch<EmailLogEntry>(`/api/v1/settings/email/log/${id}`);
}

export function postServerCleanup(targets: string[]) {
  return apiFetch<{ reclaimed: Record<string, number> }>(
    "/api/v1/server/cleanup",
    {
      method: "POST",
      json: { targets },
    },
  );
}

export function requestPasswordReset(email: string) {
  return apiFetch<{
    ok: boolean;
    smtp_configured?: boolean;
    email_sent?: boolean;
  }>("/api/v1/auth/password/forgot", {
    method: "POST",
    json: { email },
  });
}

export function resetPassword(token: string, password: string) {
  return apiFetch<{ ok: boolean }>("/api/v1/auth/password/reset", {
    method: "POST",
    json: { token, password },
  });
}

export type ApplicationRow = {
  id: string;
  name: string;
  slug: string;
  desired_state: string;
  project_id: string;
  project_name: string;
  project_slug: string;
};

export function listApplications() {
  return apiFetch<{ applications: ApplicationRow[] }>("/api/v1/applications");
}

export type InboxNotification = {
  id: string;
  category: "deploys" | "usage" | "system";
  level: "error" | "warning" | "success" | "info";
  title: string;
  body: string;
  source_label: string;
  cta_label?: string;
  cta_href?: string;
  read_at?: number;
  deployment_id?: string;
  service_id?: string;
  project_id?: string;
  created_at: number;
};

export type InboxSummary = {
  unread: number;
  total: number;
};

export function getInboxSummary() {
  return apiFetch<InboxSummary>("/api/v1/inbox/summary");
}

export function listInboxNotifications(params?: {
  category?: string;
  unread?: boolean;
  limit?: number;
  before?: number;
}) {
  const q = new URLSearchParams();
  if (params?.category) q.set("category", params.category);
  if (params?.unread) q.set("unread", "1");
  if (params?.limit) q.set("limit", String(params.limit));
  if (params?.before) q.set("before", String(params.before));
  const suffix = q.toString() ? `?${q.toString()}` : "";
  return apiFetch<{ notifications: InboxNotification[] }>(
    `/api/v1/inbox/notifications${suffix}`,
  );
}

export function markInboxNotificationRead(id: string) {
  return apiFetch<void>(`/api/v1/inbox/notifications/${id}/read`, {
    method: "POST",
  });
}

export function markAllInboxNotificationsRead() {
  return apiFetch<void>("/api/v1/inbox/notifications/read-all", {
    method: "POST",
  });
}

export type NotificationChannel = {
  id: string;
  kind: "slack" | "discord" | "email" | "webhook";
  name: string;
  enabled: boolean;
  config: Record<string, unknown>;
  secret_set: boolean;
  secret_hint?: string;
  min_level: string;
  event_types: string[];
  project_id?: string | null;
  status: string;
  disabled_reason?: string;
  consecutive_failures: number;
  last_success_at?: number;
  last_failure_at?: number;
  last_error?: string;
  created_at: number;
};

export type NotificationSettingsPayload = {
  settings: {
    disk_warn_pct: number;
    disk_crit_pct: number;
    cert_warn_pct: number;
    reminder_hours: number;
    instance_name: string;
    allow_private_targets: boolean;
    heartbeat_url?: string;
  };
  event_types: string[];
  smtp_configured: boolean;
};

export function getNotificationSettings() {
  return apiFetch<NotificationSettingsPayload>(
    "/api/v1/settings/notifications",
  );
}

export function patchNotificationSettings(body: {
  disk_warn_pct?: number;
  disk_crit_pct?: number;
  cert_warn_pct?: number;
  reminder_hours?: number;
  instance_name?: string;
  allow_private_targets?: boolean;
  heartbeat_url?: string;
}) {
  return apiFetch<NotificationSettingsPayload>(
    "/api/v1/settings/notifications",
    { method: "PATCH", json: body },
  );
}

export function listNotificationChannels() {
  return apiFetch<{ channels: NotificationChannel[] }>(
    "/api/v1/settings/notifications/channels",
  );
}

export function createNotificationChannel(body: {
  kind: string;
  name: string;
  enabled?: boolean;
  config?: Record<string, unknown>;
  secret?: Record<string, unknown>;
  min_level?: string;
  event_types?: string[];
  project_id?: string;
}) {
  return apiFetch<NotificationChannel>(
    "/api/v1/settings/notifications/channels",
    { method: "POST", json: body },
  );
}

export function patchNotificationChannel(
  id: string,
  body: {
    name?: string;
    enabled?: boolean;
    config?: Record<string, unknown>;
    secret?: Record<string, unknown>;
    min_level?: string;
    event_types?: string[];
    project_id?: string | null;
  },
) {
  return apiFetch<NotificationChannel>(
    `/api/v1/settings/notifications/channels/${id}`,
    { method: "PATCH", json: body },
  );
}

export function deleteNotificationChannel(id: string) {
  return apiFetch<void>(`/api/v1/settings/notifications/channels/${id}`, {
    method: "DELETE",
  });
}

export function testNotificationChannel(id: string) {
  return apiFetch<{
    ok: boolean;
    http_status?: number;
    latency_ms?: number;
    error?: string;
  }>(`/api/v1/settings/notifications/channels/${id}/test`, {
    method: "POST",
  });
}

export type NotificationDelivery = {
  id: string;
  event_id: string;
  status: string;
  attempts: number;
  http_status?: number;
  last_error?: string;
  created_at: number;
  sent_at?: number;
};

export function listNotificationDeliveries(id: string, limit = 50) {
  return apiFetch<{ deliveries: NotificationDelivery[] }>(
    `/api/v1/settings/notifications/channels/${id}/deliveries?limit=${limit}`,
  );
}

export function retryNotificationDelivery(id: string) {
  return apiFetch<{ ok: boolean }>(
    `/api/v1/settings/notifications/deliveries/${id}/retry`,
    { method: "POST" },
  );
}

export type TerminalTarget = {
  service_id: string;
  service_name: string;
  kind: string;
  service_kind: string;
  compose_service?: string;
  container_name: string;
};

export function listTerminalTargets() {
  return apiFetch<{ targets: TerminalTarget[] }>("/api/v1/terminal/targets");
}

export type TerminalMeta = {
  session_kinds: {
    container: { available: boolean };
    host: {
      available: boolean;
      allowed_by_instance: boolean;
      requires_mfa_if_enabled: boolean;
    };
    test: { available: boolean };
  };
};

export function getTerminalMeta() {
  return apiFetch<TerminalMeta>("/api/v1/terminal/meta");
}

export function getTerminalPreferences() {
  return apiFetch<{ theme_id: string }>("/api/v1/terminal/preferences");
}

export function patchTerminalPreferences(body: { theme_id: string }) {
  return apiFetch<{ theme_id: string }>("/api/v1/terminal/preferences", {
    method: "PATCH",
    json: body,
  });
}

export function createTerminalSession(input: {
  session_kind?: "container" | "host" | "test";
  project_id?: string;
  service_id?: string;
  compose_service?: string;
  mode?: string;
  exec_user?: string;
  mfa_code?: string;
}) {
  return apiFetch<{ ticket: string; expires_at: number }>(
    "/api/v1/terminal/sessions",
    { method: "POST", json: input },
  );
}
