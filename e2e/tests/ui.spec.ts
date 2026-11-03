import { expect, test } from "@playwright/test";

test("health endpoint responds", async ({ request }) => {
  const res = await request.get("/api/v1/health");
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  expect(body.status).toBe("ok");
});

test("sign-in page shows heading", async ({ page }) => {
  await page.goto("/sign-in/");
  await expect(page.getByRole("heading", { name: /sign in/i })).toBeVisible();
});
