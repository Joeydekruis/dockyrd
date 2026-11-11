"use client";

import { usePathname } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Snackbar } from "@/components/dockyard/ui/snackbar";
import { type ActiveDeployment, getActiveDeployment } from "@/lib/api";
import { activeDeploymentHref } from "@/lib/deployment-routes";
import { deploymentIsInProgress } from "@/lib/static-site-display";

const POLL_MS = 2500;

export function DeployActivitySnackbar() {
  const pathname = usePathname();
  const [active, setActive] = useState<ActiveDeployment | null>(null);
  const [visible, setVisible] = useState(false);
  const hadActive = useRef(false);

  const poll = useCallback(async () => {
    try {
      const { deployment } = await getActiveDeployment();
      if (deployment && deploymentIsInProgress(deployment.status)) {
        setActive(deployment);
        setVisible(true);
        hadActive.current = true;
        return;
      }
      if (hadActive.current) {
        setVisible(false);
        setActive(null);
        hadActive.current = false;
      } else {
        setActive(null);
        setVisible(false);
      }
    } catch {
      /* unauthenticated shell */
    }
  }, []);

  useEffect(() => {
    void poll();
    const id = window.setInterval(() => void poll(), POLL_MS);
    return () => window.clearInterval(id);
  }, [poll]);

  if (!visible || !active) return null;

  const href = activeDeploymentHref(active);
  const onDetail =
    pathname === href ||
    pathname.endsWith(`/deployments/${active.id}`);

  if (onDetail) return null;

  return (
    <div className="pointer-events-none fixed bottom-6 right-6 z-[60] flex justify-end px-4">
      <div className="pointer-events-auto w-full max-w-sm">
        <Snackbar
          elevated
          tone="info"
          title={`Deploying ${active.service_name}`}
          description={active.status.replaceAll("_", " ")}
          loading
          action={
            <Button href={href} variant="primary" className="h-9 px-4 text-xs">
              View deployment
            </Button>
          }
        />
      </div>
    </div>
  );
}
