import type { ServiceVariable } from "@/lib/api";

export const ENV_VAR_REF_PATTERN =
  /\$\{\{([a-z0-9][a-z0-9-]*)\.([A-Z][A-Z0-9_]*)\}\}/;

export function parseEnvVarRef(
  value: string,
): { slug: string; key: string } | null {
  const m = ENV_VAR_REF_PATTERN.exec(value);
  if (!m) return null;
  return { slug: m[1], key: m[2] };
}

export function envVarScopeLabel(v: {
  at_build: boolean;
  at_runtime: boolean;
}): string {
  if (v.at_build && v.at_runtime) return "Runtime + build";
  if (v.at_build && !v.at_runtime) return "Build";
  return "Runtime";
}

export function isBrokenEnvRef(
  value: string | undefined,
  linkedSlugs: ReadonlySet<string>,
): boolean {
  if (!value?.includes("${{")) return false;
  const ref = parseEnvVarRef(value);
  if (!ref) return value.includes("${{");
  return !linkedSlugs.has(ref.slug);
}

export type PendingEnvRedeployInput = {
  variables: ServiceVariable[];
  linkedSlugs: ReadonlySet<string>;
  /** Service row `updated_at` (seconds). */
  serviceUpdatedAt?: number;
  /** Latest live deployment `finished_at` or `started_at` (seconds). */
  liveDeploymentAt?: number;
};

export function hasPendingEnvRedeploy(input: PendingEnvRedeployInput): boolean {
  for (const v of input.variables) {
    if (isBrokenEnvRef(v.value, input.linkedSlugs)) return true;
  }
  const live = input.liveDeploymentAt ?? 0;
  const updated = input.serviceUpdatedAt ?? 0;
  if (live > 0 && updated > live + 1) {
    return true;
  }
  if (live === 0 && updated > 0 && input.variables.length > 0) {
    return true;
  }
  return false;
}
