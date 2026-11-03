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
  return apiFetch<{ ok: boolean }>("/api/v1/auth/login", {
    method: "POST",
    json: { email, password, remember: true },
  });
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
  kind: "app" | "static_site";
  name: string;
  config?: Record<string, unknown>;
}) {
  return apiFetch<Service>("/api/v1/services", {
    method: "POST",
    json: { environment: "production", ...input },
  });
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
