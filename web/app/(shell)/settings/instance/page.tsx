import { Suspense } from "react";
import { InstanceSettingsPage } from "@/components/dockyard/instance-settings-page";

export default function InstanceSettingsRoute() {
  return (
    <Suspense fallback={null}>
      <InstanceSettingsPage />
    </Suspense>
  );
}
