import { describe, expect, it } from "vitest";
import type { ServerContainer, Service } from "@/lib/api";
import { resolveDatabaseProvisionStep } from "@/lib/resource-provisioning";

const baseService: Service = {
  id: "svc-1",
  environment_id: "env-1",
  kind: "postgres",
  name: "db",
  slug: "db",
  desired_state: "running",
  config: {},
  created_at: 0,
  updated_at: 0,
};

describe("resolveDatabaseProvisionStep", () => {
  it("returns 0 when service is missing", () => {
    expect(
      resolveDatabaseProvisionStep({
        service: null,
        containers: [],
        serviceId: "svc-1",
        pollAttempts: 0,
      }),
    ).toBe(0);
  });

  it("returns 2 when a running container exists for the service", () => {
    const containers: ServerContainer[] = [
      {
        id: "c1",
        name: "db",
        state: "running",
        service_id: "svc-1",
      },
    ];
    expect(
      resolveDatabaseProvisionStep({
        service: baseService,
        containers,
        serviceId: "svc-1",
        pollAttempts: 1,
      }),
    ).toBe(2);
  });

  it("returns 2 after enough polls when desired state is running", () => {
    expect(
      resolveDatabaseProvisionStep({
        service: baseService,
        containers: [],
        serviceId: "svc-1",
        pollAttempts: 3,
      }),
    ).toBe(2);
  });
});
