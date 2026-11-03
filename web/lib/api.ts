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
};

export type SetupStatusResponse = {
  complete: boolean;
};

export function getMe() {
  return apiFetch<MeResponse>("/api/v1/auth/me");
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
  kind: "app" | "static_site" | "postgres" | "redis" | "bucket";
  name: string;
  config?: Record<string, unknown>;
}) {
  return apiFetch<Service>("/api/v1/services", {
    method: "POST",
    json: { environment: "production", ...input },
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

export function promoteDeployment(deploymentId: string) {
  return apiFetch<Deployment>(`/api/v1/deployments/${deploymentId}/promote`, {
    method: "POST",
    json: {},
  });
}

export type GitHubStatus = {
  linked: boolean;
  slug?: string;
  installation_id?: number;
};

export function getGitHubStatus() {
  return apiFetch<GitHubStatus>("/api/v1/github/status");
}

export function startGitHubAppLink() {
  return apiFetch<{ url: string; state: string }>(
    "/api/v1/github/app/manifest",
    { method: "POST", json: {} },
  );
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

export type MetricRange = "1h" | "24h" | "7d";

export type HostMetricPoint = {
  ts: number;
  cpu: number;
  mem_used: number;
  mem_total: number;
  disk_used: number;
  disk_total: number;
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
  update_available: boolean;
  latest_tag?: string;
};

export function getServerInfo() {
  return apiFetch<ServerInfo>("/api/v1/server");
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
  return apiFetch<{ ok: boolean; reset_path?: string }>(
    "/api/v1/auth/password/forgot",
    {
      method: "POST",
      json: { email },
    },
  );
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
