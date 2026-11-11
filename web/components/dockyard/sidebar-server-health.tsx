"use client";

import { useEffect, useState } from "react";
import { ServerHealthCard } from "@/components/dockyard/ui/server-health-card";
import { getServerInfo, type ServerInfo } from "@/lib/api";
import { formatPanelVersion } from "@/lib/panel-version";

export function SidebarServerHealth() {
  const [info, setInfo] = useState<ServerInfo | null>(null);

  useEffect(() => {
    void getServerInfo()
      .then(setInfo)
      .catch(() => setInfo(null));
  }, []);

  const version = info ? formatPanelVersion(info.version) : "…";

  const statusLine = info?.update_available ? "Update available" : "Up to date";

  return (
    <div className="mt-auto">
      <ServerHealthCard healthy version={version} statusLine={statusLine} />
    </div>
  );
}
