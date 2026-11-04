import { Suspense } from "react";
import { IntegrationsSettingsPage } from "@/components/dockyard/integrations-settings-page";

export default function IntegrationsSettingsRoute() {
  return (
    <Suspense fallback={null}>
      <IntegrationsSettingsPage />
    </Suspense>
  );
}
