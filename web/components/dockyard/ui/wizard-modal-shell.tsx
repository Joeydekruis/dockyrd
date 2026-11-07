"use client";

import { Check, X } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "@/components/dockyard/ui/button";

export type WizardStepDef = {
  id: string;
  label: string;
};

type WizardModalShellProps = {
  open: boolean;
  title: string;
  subtitle: string;
  steps: WizardStepDef[];
  step: number;
  onClose: () => void;
  onBack?: () => void;
  onNext?: () => void;
  nextLabel?: string;
  nextDisabled?: boolean;
  busy?: boolean;
  cancelLabel?: string;
  error?: string | null;
  children: ReactNode;
};

export function WizardModalShell({
  open,
  title,
  subtitle,
  steps,
  step,
  onClose,
  onBack,
  onNext,
  nextLabel = "Next",
  nextDisabled,
  busy,
  cancelLabel = "Cancel",
  error,
  children,
}: WizardModalShellProps) {
  if (!open) return null;

  const total = steps.length;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        className="absolute inset-0 border-0"
        style={{ background: "rgba(11, 31, 58, 0.45)" }}
        aria-label="Close dialog"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="wizard-modal-title"
        className="relative z-10 flex max-h-[90vh] w-full max-w-[840px] flex-col overflow-hidden rounded-[20px] border shadow-lg"
        style={{ background: "var(--surface)", borderColor: "var(--border)" }}
      >
        <header
          className="border-b px-8 pt-7 pb-0"
          style={{ borderColor: "var(--border)" }}
        >
          <div className="flex items-start justify-between gap-4">
            <div>
              <h2
                id="wizard-modal-title"
                className="text-[28px] font-semibold leading-tight tracking-[-0.03em]"
              >
                {title}
              </h2>
              <p
                className="mt-1 text-sm"
                style={{ color: "var(--muted-text)" }}
              >
                {subtitle}
              </p>
            </div>
            <button
              type="button"
              className="btn-ghost rounded-lg p-2"
              aria-label="Close"
              onClick={onClose}
            >
              <X className="size-5" strokeWidth={1.8} />
            </button>
          </div>
          <nav
            aria-label="Progress"
            className="wizard-stepper mt-8 flex gap-6 overflow-x-auto pb-0"
          >
            {steps.map((s, index) => {
              const n = index + 1;
              const done = n < step;
              const active = n === step;
              return (
                <div
                  key={s.id}
                  className={`wizard-stepper-item shrink-0 pb-4 ${active ? "wizard-stepper-item-active" : ""}`}
                >
                  <div className="flex items-center gap-2">
                    <span
                      className="flex size-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold"
                      style={
                        done
                          ? { background: "#e6f4ec", color: "#176b3d" }
                          : active
                            ? { background: "#102a43", color: "#fff" }
                            : {
                                background: "#eceef1",
                                color: "var(--muted-text)",
                              }
                      }
                    >
                      {done ? (
                        <Check
                          className="size-4"
                          strokeWidth={2.5}
                          aria-hidden
                        />
                      ) : (
                        n
                      )}
                    </span>
                    <span
                      className="text-sm font-medium whitespace-nowrap"
                      style={
                        active
                          ? { color: "var(--ink-navy)" }
                          : { color: "var(--muted-text)" }
                      }
                    >
                      {s.label}
                    </span>
                  </div>
                </div>
              );
            })}
          </nav>
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto px-8 py-6">
          {error ? (
            <p className="mb-4 text-sm text-red-600" role="alert">
              {error}
            </p>
          ) : null}
          {children}
        </div>

        <footer
          className="flex flex-wrap items-center justify-between gap-3 border-t px-8 py-4"
          style={{ borderColor: "var(--border)" }}
        >
          <p
            className="font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Step {step} of {total}
          </p>
          <div className="flex flex-wrap gap-2">
            {step > 1 && onBack ? (
              <Button
                type="button"
                variant="outline"
                onClick={onBack}
                disabled={busy}
              >
                Back
              </Button>
            ) : (
              <Button
                type="button"
                variant="outline"
                onClick={onClose}
                disabled={busy}
              >
                {cancelLabel}
              </Button>
            )}
            {onNext ? (
              <Button
                type="button"
                onClick={onNext}
                disabled={nextDisabled || busy}
              >
                {nextLabel}
              </Button>
            ) : null}
          </div>
        </footer>
      </div>
    </div>
  );
}
