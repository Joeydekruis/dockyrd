"use client";

import { Backdrop, ProgressBar } from "@/components/dockyard/ui";

export type UpdateApplyPhase = "applying" | "restarting" | "complete" | "error";

const phaseTitles: Record<UpdateApplyPhase, string> = {
  applying: "Installing update",
  restarting: "Restarting dockyrd",
  complete: "Update complete",
  error: "Update failed",
};

type UpdateApplyingOverlayProps = {
  open: boolean;
  phase: UpdateApplyPhase;
  detail?: string | null;
};

export function UpdateApplyingOverlay({
  open,
  phase,
  detail,
}: UpdateApplyingOverlayProps) {
  if (!open) {
    return null;
  }

  const progressVariant =
    phase === "complete"
      ? "complete"
      : phase === "error"
        ? "error"
        : "indeterminate";

  const defaultDetail =
    phase === "applying"
      ? "Downloading and verifying the release. This can take a few minutes."
      : phase === "restarting"
        ? "Your apps and proxy keep running. The dashboard will reconnect shortly."
        : phase === "complete"
          ? "Reloading the panel…"
          : "Something went wrong.";

  return (
    <>
      <Backdrop visible label="Update in progress" />
      <div
        className="pointer-events-none fixed inset-0 z-50 flex items-center justify-center p-4"
        aria-live="polite"
      >
        <div
          role="alertdialog"
          aria-modal="true"
          aria-labelledby="update-overlay-title"
          aria-busy={phase === "applying" || phase === "restarting"}
          className="pointer-events-auto w-full max-w-md rounded-[20px] border p-6 shadow-lg"
          style={{ background: "var(--surface)", borderColor: "var(--border)" }}
        >
          <h2 id="update-overlay-title" className="text-lg font-semibold">
            {phaseTitles[phase]}
          </h2>
          <p className="mt-2 text-sm" style={{ color: "var(--muted-text)" }}>
            {detail ?? defaultDetail}
          </p>
          <div className="mt-6">
            <ProgressBar
              label={phase === "complete" ? "Done" : "Progress"}
              variant={progressVariant}
              value={phase === "error" ? 100 : 0}
            />
          </div>
        </div>
      </div>
    </>
  );
}
