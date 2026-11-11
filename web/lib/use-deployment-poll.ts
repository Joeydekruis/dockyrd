import { useEffect, useRef } from "react";

/** Poll while `active` is true (e.g. deployment still building). */
export function useDeploymentPoll(
  active: boolean,
  onTick: () => void | Promise<void>,
  intervalMs = 2000,
): void {
  const wasActive = useRef(false);

  useEffect(() => {
    if (wasActive.current && !active) {
      void onTick();
    }
    wasActive.current = active;
  }, [active, onTick]);

  useEffect(() => {
    if (!active) return;
    void onTick();
    const id = window.setInterval(() => void onTick(), intervalMs);
    return () => window.clearInterval(id);
  }, [active, onTick, intervalMs]);
}
