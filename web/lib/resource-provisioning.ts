import type { ServerContainer, Service } from "@/lib/api";

export function normalizeResourceState(value: string): string {
  return value.trim().toLowerCase();
}

export function isRunningResourceState(value: string): boolean {
  const normalized = normalizeResourceState(value);
  return normalized === "running" || normalized === "live";
}

export function isFailedResourceState(value: string): boolean {
  const normalized = normalizeResourceState(value);
  return normalized === "failed" || normalized === "error";
}

/** 0 = create record, 1 = start container, 2 = ready (complete), -1 = failed */
export function resolveDatabaseProvisionStep(input: {
  service: Service | null;
  containers: ServerContainer[];
  serviceId: string;
  pollAttempts: number;
}): number {
  const { service, containers, serviceId, pollAttempts } = input;
  if (!service) return 0;
  if (isFailedResourceState(service.desired_state)) return -1;

  const step = 1;
  const hasRunningContainer = containers.some(
    (c) => c.service_id === serviceId && isRunningResourceState(c.state ?? ""),
  );
  if (hasRunningContainer) return 2;
  if (isRunningResourceState(service.desired_state) && pollAttempts >= 3) {
    return 2;
  }
  return step;
}
