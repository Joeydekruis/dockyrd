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
  kind: "app" | "static_site" | "postgres" | "redis" | "bucket";
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
  return apiFetch<{ variables: Record<string, string> }>(
    `/api/v1/services/${serviceId}/connection${q}`,
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
  branch?: string;
  build_log_path?: string;
  error?: string;
  queued_at: number;
  started_at?: number;
  finished_at?: number;
};

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
};

export function listServiceDomains(serviceId: string) {
  return apiFetch<{ domains: ServiceDomain[] }>(
    `/api/v1/services/${serviceId}/domains`,
  );
}

export function createServiceDomain(
  serviceId: string,
  input: { hostname?: string; platform?: boolean; primary?: boolean },
) {
  return apiFetch<ServiceDomain>(`/api/v1/services/${serviceId}/domains`, {
    method: "POST",
    json: input,
  });
}

export function deleteServiceDomain(domainId: string) {
  return apiFetch<void>(`/api/v1/domains/${domainId}`, { method: "DELETE" });
}

export type ServiceVariable = {
  key: string;
  is_secret: boolean;
  at_build: boolean;
  at_runtime: boolean;
  value?: string;
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

export type DockyardEvent = {
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
  return apiFetch<{ events: DockyardEvent[] }>(`/api/v1/events${suffix}`);
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

export function applyDockyardUpdate() {
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
  require_2fa_all: boolean;
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

export function patchInstanceSettings(body: {
  public_url?: string;
  acme_email?: string;
  wildcard_domain?: string;
  require_2fa_all?: boolean;
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
