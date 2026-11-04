"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
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
          <Link
            href="/settings/updates"
            className="rounded-lg bg-[var(--primary)] px-4 py-2 text-sm font-medium text-[var(--primary-foreground)]"
            onClick={onClose}
          >
            Go to Updates
          </Link>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-[var(--border)] px-4 py-2 text-sm"
          >
            Close
          </button>
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
      <div className="mb-4 flex flex-wrap items-center justify-between gap-2 rounded-lg border border-amber-200 bg-amber-50 px-4 py-2 text-sm text-amber-950 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-100">
        <span>
          Update <span className="font-mono">{info.latest_tag}</span> is
          available.
        </span>
        <div className="flex gap-2">
          <Link
            href="/settings/updates"
            className="rounded-md border border-current px-3 py-1 text-xs font-medium"
          >
            Update
          </Link>
          <button
            type="button"
            className="rounded-md border border-current px-3 py-1 text-xs font-medium"
            onClick={() => setOpen(true)}
          >
            Release notes
          </button>
        </div>
      </div>
      <UpdateAvailableModal
        open={open}
        onClose={() => setOpen(false)}
        info={info}
      />
    </>
  );
}
