"use client";

import { useEffect, useState } from "react";
import { Alert } from "@/components/dockyard/ui/alert";
import { Button } from "@/components/dockyard/ui/button";
import { getServerInfo, type ServerInfo } from "@/lib/api";

type Props = {
  open: boolean;
  onClose: () => void;
  info: ServerInfo | null;
};

export function UpdateAvailableModal({ open, onClose, info }: Props) {
  if (!open || !info?.update_available) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="update-modal-title"
    >
      <div className="max-h-[85vh] w-full max-w-lg overflow-y-auto rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6 shadow-lg">
        <h2 id="update-modal-title" className="text-lg font-medium">
          Update available
        </h2>
        <p className="mt-2 text-sm text-[var(--muted-foreground)]">
          Running{" "}
          <span className="font-mono text-[var(--foreground)]">
            {info.version}
          </span>
          . Latest release: <span className="font-mono">{info.latest_tag}</span>
        </p>
        {info.release_notes ? (
          <div className="mt-4 max-h-48 overflow-y-auto rounded-lg border border-[var(--border)] bg-[var(--background)] p-3 text-sm whitespace-pre-wrap">
            {info.release_notes}
          </div>
        ) : null}
        <p className="mt-4 text-sm text-[var(--muted-foreground)]">
          Open Settings → Updates to review release notes and install from the
          dashboard.
        </p>
        <div className="mt-6 flex flex-wrap gap-2">
          <Button href="/settings/server/updates" onClick={onClose}>
            Go to Updates
          </Button>
          <Button type="button" variant="outline" onClick={onClose}>
            Close
          </Button>
        </div>
      </div>
    </div>
  );
}

/** Admin/owner banner + modal when a newer GitHub release exists. */
export function UpdateAvailableNotifier() {
  const [info, setInfo] = useState<ServerInfo | null>(null);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    void getServerInfo()
      .then(setInfo)
      .catch(() => setInfo(null));
  }, []);

  if (!info?.update_available) return null;

  return (
    <>
      <div className="mb-4">
        <Alert
          tone="info"
          title={
            <>
              Update <span className="font-mono">{info.latest_tag}</span> is
              available
            </>
          }
          description="Review release notes and install from Settings → Updates."
          action={
            <>
              <Button href="/settings/server/updates" variant="outline">
                Update
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(true)}
              >
                Release notes
              </Button>
            </>
          }
        />
      </div>
      <UpdateAvailableModal
        open={open}
        onClose={() => setOpen(false)}
        info={info}
      />
    </>
  );
}
