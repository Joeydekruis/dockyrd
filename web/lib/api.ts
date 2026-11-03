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
