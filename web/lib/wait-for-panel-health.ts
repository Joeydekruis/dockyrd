/** Poll until the panel responds after a service restart (e.g. self-update). */
export async function waitForPanelHealth(timeoutMs = 120_000): Promise<void> {
  const started = Date.now();
  await new Promise((resolve) => setTimeout(resolve, 800));

  while (Date.now() - started < timeoutMs) {
    try {
      const res = await fetch("/api/v1/health", { credentials: "same-origin" });
      if (res.ok) {
        return;
      }
    } catch {
      // Expected while dockyrd serve is down.
    }
    await new Promise((resolve) => setTimeout(resolve, 1500));
  }

  throw new Error(
    "The panel did not respond after the restart. Try reloading this page.",
  );
}
