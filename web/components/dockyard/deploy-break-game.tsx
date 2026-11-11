"use client";

import { useEffect, useRef } from "react";
import { mountDeployBreak } from "@/lib/deploy-break/engine.js";

type DeployBreakGameProps = {
  deploymentId: string;
  serviceLabel: string;
  releaseLabel: string;
  deployProgress: number;
  deployDone: boolean;
  initialBest?: number;
  onScoreChange?: (score: number) => void;
  onBestChange?: (best: number) => void;
};

export function DeployBreakGame({
  deploymentId,
  serviceLabel,
  releaseLabel,
  deployProgress,
  deployDone,
  initialBest = 0,
  onScoreChange,
  onBestChange,
}: DeployBreakGameProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const optsRef = useRef({
    releaseLabel,
    serviceLabel,
    initialBest,
    deployProgress,
    deployDone,
    onScoreChange,
    onBestChange,
  });

  optsRef.current.releaseLabel = releaseLabel;
  optsRef.current.serviceLabel = serviceLabel;
  optsRef.current.deployProgress = deployProgress;
  optsRef.current.deployDone = deployDone;
  optsRef.current.onScoreChange = onScoreChange;
  optsRef.current.onBestChange = onBestChange;

  // biome-ignore lint/correctness/useExhaustiveDependencies: remount canvas game when deployment id changes
  useEffect(() => {
    const root = rootRef.current;
    const canvas = canvasRef.current;
    if (!root || !canvas) return;
    optsRef.current.initialBest = initialBest;
    mountDeployBreak(canvas, root, optsRef.current);
  }, [deploymentId, initialBest]);

  return (
    <div
      ref={rootRef}
      className="deploy-break w-full"
      data-deployment-id={deploymentId}
    >
      <div
        className="rounded-xl border px-4 py-3 mb-3"
        style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        data-dy-id="deploy"
        id="deploy"
      >
        <div className="flex justify-between items-center gap-3 mb-2.5">
          <span
            className="inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-[11px] font-mono uppercase tracking-wide bg-[var(--accent-blue-soft)] text-[var(--link-blue)]"
            data-dy-id="badge"
            id="badge"
          >
            <i className="inline-block size-1.5 rounded-full bg-[var(--accent-blue)] animate-pulse" />
            <span data-dy-id="badgeText" id="badgeText">
              Deploying
            </span>
          </span>
          <span
            className="font-mono text-[13px] font-semibold"
            data-dy-id="pct"
            id="pct"
          >
            0%
          </span>
        </div>
        <div
          className="h-2.5 rounded-md overflow-hidden bg-[var(--hover-row)]"
          data-dy-id="bar"
          id="bar"
        >
          <b
            className="block h-full w-0 rounded-md bg-[var(--accent-blue)] transition-[width] duration-200"
            data-dy-id="barfill"
            id="barfill"
          />
        </div>
        <p
          className="mt-2 font-mono text-[13px] text-[var(--muted-text)]"
          data-dy-id="stage"
          id="stage"
        >
          Pulling image…
        </p>
      </div>

      <div
        className="relative rounded-[20px] border p-2.5"
        style={{ borderColor: "var(--border)", background: "var(--surface)" }}
      >
        <canvas
          ref={canvasRef}
          data-dy-id="c"
          id="c"
          width={640}
          height={480}
          className="block w-full aspect-[4/3] rounded-xl bg-white touch-none"
          aria-label="Breakout game: move the dy-proxy paddle to keep the container cube in play"
        />
        <div
          className="absolute inset-2.5 rounded-xl flex flex-col items-center justify-center text-center p-6 gap-2.5 bg-[rgba(250,249,248,0.92)]"
          data-dy-id="overlay"
          id="overlay"
        >
          <span className="inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-[11px] font-mono uppercase tracking-wide bg-[var(--accent-blue-soft)] text-[var(--link-blue)]">
            <i className="inline-block size-1.5 rounded-full bg-[var(--accent-blue)] animate-pulse" />
            Deploy in progress
          </span>
          <h2
            className="text-[28px] font-semibold tracking-tight m-0"
            data-dy-id="ovTitle"
            id="ovTitle"
          >
            Deploy Break
          </h2>
          <p
            className="text-sm text-[var(--muted-text)] max-w-md m-0"
            data-dy-id="ovText"
            id="ovText"
          >
            Bounce the container cube off dy-proxy and clear the pipeline. Your
            deploy keeps running in the background.
          </p>
          <div className="flex gap-2.5 flex-wrap justify-center mt-1">
            <button
              type="button"
              className="btn-primary h-11 px-5 rounded-[10px] text-xs font-mono uppercase"
              data-dy-id="ovBtn"
              id="ovBtn"
            >
              Start
            </button>
          </div>
          <p className="font-mono text-xs text-[var(--muted-text)] mt-1">
            Mouse, touch or ← → to move · Space or click to launch · P to pause
          </p>
        </div>
      </div>

      <div className="flex justify-end gap-2.5 mt-3 mb-3">
        <button
          type="button"
          className="btn-outline h-9 px-3.5 rounded-[10px] text-xs font-mono uppercase"
          data-dy-id="muteBtn"
          id="muteBtn"
        >
          Sound: on
        </button>
        <button
          type="button"
          className="btn-outline h-9 px-3.5 rounded-[10px] text-xs font-mono uppercase"
          data-dy-id="pauseBtn"
          id="pauseBtn"
        >
          Pause
        </button>
      </div>

      <div
        className="rounded-2xl px-4 py-3.5 font-mono text-xs leading-relaxed min-h-[7rem] overflow-hidden"
        style={{ background: "var(--ink-navy)", color: "var(--banner-blue)" }}
        data-dy-id="log"
        id="log"
        aria-live="off"
      />

      <p className="mt-4 font-mono text-xs text-[var(--muted-text)]">
        Power-ups: REPLICAS (3 balls), SCALE UP, CACHE HIT, ROLLBACK (+1
        uptime). Avoid OOM KILLED and CPU SPIKE.
      </p>

      <div className="sr-only" aria-hidden>
        <span data-dy-id="score" id="score" />
        <span data-dy-id="lives" id="lives" />
        <span data-dy-id="level" id="level" />
        <span data-dy-id="best" id="best" />
      </div>
    </div>
  );
}

/** Stats row rendered in React above the game (engine still updates hidden nodes for compatibility). */
export function DeployBreakStatBar({
  deploys,
  release,
  best,
  uptime = "3 / 3",
}: {
  deploys: number | string;
  release: string;
  best: number | string;
  uptime?: string;
}) {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-3">
      {[
        { label: "Deploys", value: deploys },
        { label: "Uptime", value: uptime },
        { label: "Release", value: release },
        { label: "Best", value: best },
      ].map((s) => (
        <div
          key={s.label}
          className="rounded-2xl border px-4 py-3.5"
          style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        >
          <p className="font-mono text-[11px] uppercase tracking-wide text-[var(--muted-text)] m-0">
            {s.label}
          </p>
          <p className="text-2xl font-semibold tracking-tight mt-1 m-0">
            {s.value}
          </p>
        </div>
      ))}
    </div>
  );
}
