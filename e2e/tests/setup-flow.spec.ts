import { readFileSync } from "node:fs";
import { expect, test } from "@playwright/test";

const tokenFile =
  process.env.E2E_SETUP_TOKEN_FILE ?? "/tmp/dockyard-e2e/etc/e2e-setup-token";

test("setup completes via API", async ({ request }) => {
  let token = process.env.E2E_SETUP_TOKEN ?? "";
  if (!token) {
    try {
      token = readFileSync(tokenFile, "utf8").trim();
    } catch {
      test.skip(true, "no setup token (instance already initialized)");
    }
  }
  if (!token) {
    test.skip(true, "empty setup token");
  }

  const setup = await request.post("/api/v1/setup", {
    data: {
      token,
      name: "E2E Owner",
      email: "e2e-owner@dockyard.test",
      password: "longpassword12",
    },
  });
  expect(setup.ok()).toBeTruthy();

  const login = await request.post("/api/v1/auth/login", {
    data: {
      email: "e2e-owner@dockyard.test",
      password: "longpassword12",
    },
  });
  expect(login.ok()).toBeTruthy();

  const status = await request.get("/api/v1/setup/status");
  expect(status.ok()).toBeTruthy();
  const body = await status.json();
  expect(body.complete).toBe(true);

  const server = await request.get("/api/v1/server");
  expect(server.ok()).toBeTruthy();
  const info = await server.json();
  expect(info.version).toBeTruthy();
  expect(typeof info.uptime_sec).toBe("number");
});
