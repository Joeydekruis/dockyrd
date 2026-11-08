"use client";

import { ArrowLeft, Check, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { Chip } from "@/components/dockyard/ui/chip";
import { Spinner } from "@/components/dockyard/ui/spinner";
import {
  type Deployment,
  fetchDeploymentLog,
  getDeploymentForService,
} from "@/lib/api";
import {
  resolveStaticDeployStep,
  type StaticDeployStep,
  staticDeploySteps,
  stepVisualState,
} from "@/lib/deployment-progress";

export type StaticSiteCreateResult = {
  serviceId: string;
  deploymentId: string;
  siteName: string;
  slug: string;
  detailHref: string;
  repo: string;
  branch: string;
  installCmd: string;
  buildCmd: string;
  outputDir: string;
  defaultUrl: string;
};

type StaticSiteCreateProgressModalProps = {
  open: boolean;
  context: StaticSiteCreateResult | null;
  onClose: () => void;
  onEditSettings: () => void;
  onFinished: () => void;
};

export function StaticSiteCreateProgressModal({
  open,
  context,
  onClose,
  onEditSettings,
  onFinished,
}: StaticSiteCreateProgressModalProps) {
  const router = useRouter();
  const [deployment, setDeployment] = useState<Deployment | null>(null);
  const [log, setLog] = useState("");
  const finishedRef = useRef(false);
  const onFinishedRef = useRef(onFinished);
  onFinishedRef.current = onFinished;
  const detailHrefRef = useRef<string | undefined>(undefined);
  detailHrefRef.current = context?.detailHref;

  useEffect(() => {
    if (!open || !context) {
      setDeployment(null);
      setLog("");
      finishedRef.current = false;
      return;
    }

    let cancelled = false;

    async function tick() {
      if (!context || cancelled) return;
      try {
        const dep = await getDeploymentForService(
          context.serviceId,
          context.deploymentId,
        );
        if (cancelled) return;
        setDeployment(dep);
        const logText = await fetchDeploymentLog(context.deploymentId);
        if (!cancelled) setLog(logText);
        if (dep?.status === "live" && !finishedRef.current) {
          finishedRef.current = true;
          onFinishedRef.current();
          const href = detailHrefRef.current;
          if (href) {
            window.setTimeout(() => {
              if (!cancelled) router.push(href);
            }, 450);
          }
        }
      } catch {
        /* keep polling */
      }
    }

    void tick();
    const id = window.setInterval(tick, 1500);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, [open, context, router]);

  const steps = useMemo(
    () =>
      context
        ? staticDeploySteps({
            repo: context.repo,
            branch: context.branch,
            installCmd: context.installCmd,
            buildCmd: context.buildCmd,
            outputDir: context.outputDir,
            defaultUrl: context.defaultUrl,
          })
        : [],
    [context],
  );

  const activeStep = resolveStaticDeployStep(deployment, log);
  const isLive = deployment?.status === "live";
  const isFailed =
    deployment?.status === "failed" || deployment?.status === "cancelled";

  if (!open || !context) return null;

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
        aria-labelledby="static-create-progress-title"
        className="relative z-10 flex max-h-[90vh] w-full max-w-[920px] flex-col overflow-hidden rounded-[20px] border shadow-lg"
        style={{ background: "var(--surface)", borderColor: "var(--border)" }}
      >
        <header
          className="border-b px-8 py-6"
          style={{ borderColor: "var(--border)" }}
        >
          <div className="flex items-start justify-between gap-4">
            <div>
              <h2
                id="static-create-progress-title"
                className="text-[28px] font-semibold leading-tight tracking-[-0.03em]"
              >
                Create static site
              </h2>
              <p
                className="mt-1 text-sm"
                style={{ color: "var(--muted-text)" }}
              >
                Five short steps, then deploy.
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
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto px-8 py-6">
          <div className="mb-6">
            <Chip>{isFailed ? "Failed" : isLive ? "Live" : "Deploying"}</Chip>
            <h3 className="mt-4 text-2xl font-semibold tracking-[-0.03em]">
              {context.siteName} is being created
            </h3>
            <p
              className="mt-2 max-w-xl text-sm"
              style={{ color: "var(--muted-text)" }}
            >
              First deploy usually takes about a minute. You can close this
              window; it keeps running.
            </p>
            {isFailed && deployment?.error ? (
              <p className="mt-3 text-sm text-red-600" role="alert">
                {deployment.error}
              </p>
            ) : null}
          </div>

          <div className="grid gap-8 lg:grid-cols-2">
            <ol className="space-y-5">
              {steps.map((step, index) => (
                <DeployStepRow
                  key={step.id}
                  step={step}
                  state={stepVisualState(index, activeStep, deployment)}
                />
              ))}
            </ol>
            <div>
              <p
                className="mb-2 font-mono text-[11px] uppercase tracking-widest"
                style={{ color: "var(--muted-text)" }}
              >
                Build log
              </p>
              <pre className="create-progress-log max-h-[320px] overflow-y-auto rounded-[16px] p-4 text-[12px] leading-relaxed">
                {log.trim() || "Waiting for build output…"}
              </pre>
            </div>
          </div>
        </div>

        <footer
          className="flex flex-wrap items-center justify-between gap-3 border-t px-8 py-4"
          style={{ borderColor: "var(--border)" }}
        >
          <button
            type="button"
            className="inline-flex items-center gap-1.5 text-sm font-medium hover:underline"
            style={{ color: "var(--link-blue)" }}
            onClick={onEditSettings}
          >
            <ArrowLeft className="size-4" aria-hidden />
            Edit settings
          </button>
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Close
            </Button>
            {isLive ? (
              <Button href={context.detailHref}>Open static site</Button>
            ) : (
              <Button type="button" disabled>
                Open static site
              </Button>
            )}
          </div>
        </footer>
      </div>
    </div>
  );
}

function DeployStepRow({
  step,
  state,
}: {
  step: StaticDeployStep;
  state: ReturnType<typeof stepVisualState>;
}) {
  return (
    <li className="flex gap-3">
      <StepIndicator state={state} />
      <div>
        <p className="font-semibold">{step.title}</p>
        <p
          className="mt-0.5 text-[13px]"
          style={{ color: "var(--muted-text)" }}
        >
          {step.detail}
        </p>
      </div>
    </li>
  );
}

function StepIndicator({
  state,
}: {
  state: ReturnType<typeof stepVisualState>;
}) {
  if (state === "done") {
    return (
      <span
        className="flex size-7 shrink-0 items-center justify-center rounded-full"
        style={{ background: "#e6f4ec", color: "#176b3d" }}
      >
        <Check className="size-4" strokeWidth={2.5} aria-hidden />
      </span>
    );
  }
  if (state === "active") {
    return (
      <span className="flex size-7 shrink-0 items-center justify-center">
        <Spinner size="sm" label="In progress" />
      </span>
    );
  }
  if (state === "failed") {
    return (
      <span
        className="flex size-7 shrink-0 items-center justify-center rounded-full border-2"
        style={{ borderColor: "var(--danger)", color: "var(--danger)" }}
        aria-hidden
      >
        !
      </span>
    );
  }
  return (
    <span
      className="flex size-7 shrink-0 items-center justify-center rounded-full border-2"
      style={{ borderColor: "var(--input-border)" }}
      aria-hidden
    />
  );
}
