"use client";

import { Check, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Chip } from "@/components/dockyard/ui/chip";
import { Spinner } from "@/components/dockyard/ui/spinner";
import { listDatabases } from "@/lib/api";

export type ResourceProvisionContext = {
  kind: "database" | "application" | "bucket";
  serviceId: string;
  name: string;
  label: string;
};

type ResourceProvisioningModalProps = {
  open: boolean;
  context: ResourceProvisionContext | null;
  onClose: () => void;
  onFinished: () => void;
};

const STEPS: Record<ResourceProvisionContext["kind"], string[]> = {
  database: ["Create service", "Start container", "Ready for connections"],
  application: ["Create service", "Pull image", "Start container"],
  bucket: ["Create bucket", "Register with storage", "Ready"],
};

export function ResourceProvisioningModal({
  open,
  context,
  onClose,
  onFinished,
}: ResourceProvisioningModalProps) {
  const [activeStep, setActiveStep] = useState(0);
  const [done, setDone] = useState(false);
  const [failed, setFailed] = useState(false);
  const finishedRef = useRef(false);

  useEffect(() => {
    if (!open || !context) {
      setActiveStep(0);
      setDone(false);
      setFailed(false);
      finishedRef.current = false;
      return;
    }

    let cancelled = false;
    const steps = STEPS[context.kind];
    setActiveStep(0);
    setDone(false);
    setFailed(false);
    finishedRef.current = false;

    const markDone = () => {
      if (finishedRef.current || cancelled) return;
      finishedRef.current = true;
      setDone(true);
      setActiveStep(steps.length);
      onFinished();
    };

    const stepTimer = window.setInterval(() => {
      setActiveStep((s) => {
        if (s >= steps.length - 1) return s;
        return s + 1;
      });
    }, 1400);

    const finishTimer = window.setTimeout(() => {
      if (context.kind !== "database") markDone();
    }, 4500);

    async function pollDatabase() {
      if (context?.kind !== "database") return;
      try {
        const { databases } = await listDatabases();
        const row = databases.find((d) => d.id === context.serviceId);
        if (row?.desired_state === "running") markDone();
        if (row?.desired_state === "failed") setFailed(true);
      } catch {
        /* retry on next tick */
      }
    }

    void pollDatabase();
    const pollTimer =
      context.kind === "database"
        ? window.setInterval(pollDatabase, 2000)
        : undefined;

    return () => {
      cancelled = true;
      window.clearInterval(stepTimer);
      window.clearTimeout(finishTimer);
      if (pollTimer) window.clearInterval(pollTimer);
    };
  }, [open, context, onFinished]);

  if (!open || !context) return null;

  const steps = STEPS[context.kind];

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
        className="relative z-10 w-full max-w-lg rounded-[20px] border p-6 shadow-lg"
        style={{ background: "var(--surface)", borderColor: "var(--border)" }}
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <Chip>{failed ? "Failed" : done ? "Ready" : "Creating"}</Chip>
            <h2 className="mt-3 text-xl font-semibold tracking-tight">
              {context.name} is being created
            </h2>
            <p className="mt-2 text-sm" style={{ color: "var(--muted-text)" }}>
              {context.label}
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

        <ol className="mt-6 space-y-4">
          {steps.map((label, index) => {
            const complete = done || index < activeStep;
            const active = !done && !failed && index === activeStep;
            return (
              <li key={label} className="flex items-center gap-3">
                {complete ? (
                  <span
                    className="flex size-7 items-center justify-center rounded-full"
                    style={{ background: "#e6f4ec", color: "#176b3d" }}
                  >
                    <Check className="size-4" strokeWidth={2.5} aria-hidden />
                  </span>
                ) : active ? (
                  <span className="flex size-7 items-center justify-center">
                    <Spinner size="sm" label="In progress" />
                  </span>
                ) : (
                  <span
                    className="size-7 shrink-0 rounded-full border-2"
                    style={{ borderColor: "var(--input-border)" }}
                  />
                )}
                <span className="text-sm font-medium">{label}</span>
              </li>
            );
          })}
        </ol>

        <div className="mt-8 flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            Close
          </Button>
        </div>
      </div>
    </div>
  );
}
