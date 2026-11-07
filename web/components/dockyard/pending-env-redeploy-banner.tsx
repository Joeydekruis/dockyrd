"use client";

import { Alert } from "@/components/dockyard/ui/alert";
import { Button } from "@/components/dockyard/ui/button";

type PendingEnvRedeployBannerProps = {
  onDeploy?: () => void;
  deployBusy?: boolean;
  onDismiss?: () => void;
};

export function PendingEnvRedeployBanner({
  onDeploy,
  deployBusy,
  onDismiss,
}: PendingEnvRedeployBannerProps) {
  return (
    <Alert
      tone="warning"
      title="Redeploy to apply changes"
      description="Environment variables or service connections changed. Running containers still use the previous values until you deploy again."
      onDismiss={onDismiss}
      action={
        onDeploy ? (
          <Button
            type="button"
            variant="outline"
            disabled={deployBusy}
            onClick={onDeploy}
          >
            {deployBusy ? "Deploying…" : "Deploy now"}
          </Button>
        ) : undefined
      }
    />
  );
}
