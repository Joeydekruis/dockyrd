import { useEffect } from "react";

/** Poll while `active` is true (e.g. deployment still building). */
export function useDeploymentPoll(
  active: boolean,
  onTick: () => void | Promise<void>,
  intervalMs = 2000,
): void {
  useEffect(() => {
    if (!active) return;
    void onTick();
    const id = window.setInterval(() => void onTick(), intervalMs);
    return () => window.clearInterval(id);
  }, [active, onTick, intervalMs]);
}
