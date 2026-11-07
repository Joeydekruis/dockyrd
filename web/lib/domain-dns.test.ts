import { describe, expect, it } from "vitest";
import type { InstanceSettings } from "@/lib/api";
import {
  previewDefaultPlatformHostname,
  serverDomainForPreview,
} from "@/lib/domain-dns";

describe("serverDomainForPreview", () => {
  it("uses dashboard hostname from public_url", () => {
    const instance: InstanceSettings = {
      public_url: "https://dockyrd.derisory.nl",
      wildcard_domain: "*.dockyrd.derisory.nl",
    } as InstanceSettings;
    expect(serverDomainForPreview(instance)).toBe("dockyrd.derisory.nl");
  });

  it("strips wildcard and optional apps prefix when no public_url", () => {
    const instance = {
      wildcard_domain: "*.apps.dockyrd.derisory.nl",
    } as InstanceSettings;
    expect(serverDomainForPreview(instance)).toBe("dockyrd.derisory.nl");
  });
});

describe("previewDefaultPlatformHostname", () => {
  it("builds slug.server-domain without deploy segment or asterisk", () => {
    const instance = {
      public_url: "https://dockyrd.derisory.nl",
    } as InstanceSettings;
    expect(previewDefaultPlatformHostname("test-api", instance)).toBe(
      "test-api.dockyrd.derisory.nl",
    );
  });
});
