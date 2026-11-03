import { defineConfig, devices } from "@playwright/test";

const port = process.env.E2E_PORT ?? "18080";
const baseURL = process.env.E2E_BASE_URL ?? `http://127.0.0.1:${port}`;

export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  use: {
    ...devices["Desktop Chrome"],
    baseURL,
  },
  webServer: {
    command: "bash ../scripts/e2e-webserver.sh",
    cwd: __dirname,
    url: `${baseURL}/api/v1/health`,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
