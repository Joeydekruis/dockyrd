"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  DeployBreakGame,
  DeployBreakStatBar,
} from "@/components/dockyard/deploy-break-game";
import {
  type Deployment,
  type DeploymentBreakScoreSummary,
  fetchDeploymentBreakScore,
  putDeploymentBreakScore,
  type Service,
} from "@/lib/api";
import {
  deploymentDeployDone,
  deploymentProgressPercent,
} from "@/lib/deployment-status-progress";
import {
  deploymentReleaseTag,
  formatDeploymentStartedAt,
} from "@/lib/static-site-display";

type DeploymentDeployBreakSectionProps = {
  deployment: Deployment;
  service: Service;
};

export function DeploymentDeployBreakSection({
  deployment,
  service,
}: DeploymentDeployBreakSectionProps) {
  const [summary, setSummary] = useState<DeploymentBreakScoreSummary | null>(
    null,
  );
  const [sessionBest, setSessionBest] = useState(0);
  const [liveScore, setLiveScore] = useState(0);
  const lastSaved = useRef(0);

  const release = useMemo(
    () => deploymentReleaseTag(deployment) ?? "v0.1.1",
    [deployment],
  );

  const deployProgress = deploymentProgressPercent(deployment.status);
  const deployDone = deploymentDeployDone(deployment.status);

  const refreshScores = useCallback(async () => {
    try {
      const data = await fetchDeploymentBreakScore(deployment.id);
      setSummary(data);
      if (data.your_score != null) {
        lastSaved.current = Math.max(lastSaved.current, data.your_score);
      }
    } catch {
      /* optional */
    }
  }, [deployment.id]);

  useEffect(() => {
    void refreshScores();
  }, [refreshScores]);

  const persistScore = useCallback(
    async (score: number) => {
      if (score <= lastSaved.current) return;
      try {
        await putDeploymentBreakScore(deployment.id, score);
        lastSaved.current = score;
        await refreshScores();
      } catch {
        /* ignore */
      }
    },
    [deployment.id, refreshScores],
  );

  const onScoreChange = useCallback(
    (score: number) => {
      setLiveScore(score);
      if (score > sessionBest) {
        setSessionBest(score);
        void persistScore(score);
      }
    },
    [persistScore, sessionBest],
  );

  const onBestChange = useCallback((best: number) => {
    setSessionBest((prev) => Math.max(prev, best));
  }, []);

  const displayBest = Math.max(
    sessionBest,
    summary?.your_score ?? 0,
    summary?.deployment_best?.score ?? 0,
  );

  const allTime = summary?.all_time_best;

  return (
    <section className="w-full space-y-4" aria-label="Deploy Break">
      <div>
        <p className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-text)] m-0">
          While you wait
        </p>
        <h2 className="text-2xl sm:text-[32px] font-semibold tracking-tight mt-1 mb-2">
          Deploy Break
        </h2>
        <p className="text-sm text-[var(--muted-text)] max-w-2xl m-0">
          Keep the cube alive until {service.name} finishes deploying — or keep
          going once it&apos;s live.
        </p>
      </div>

      <DeployBreakStatBar
        deploys={liveScore.toLocaleString("en")}
        release={release}
        best={displayBest.toLocaleString("en")}
      />

      {allTime ? (
        <p className="text-sm text-[var(--muted-text)] m-0">
          All-time high:{" "}
          <span className="text-[var(--ink-navy)] font-medium">
            {allTime.score.toLocaleString("en")}
          </span>{" "}
          by {allTime.user_name} ({allTime.service_name} ·{" "}
          {formatDeploymentStartedAt(allTime.updated_at)})
        </p>
      ) : null}

      <DeployBreakGame
        deploymentId={deployment.id}
        serviceLabel={service.name}
        releaseLabel={release}
        deployProgress={deployProgress}
        deployDone={deployDone}
        initialBest={summary?.your_score ?? 0}
        onScoreChange={onScoreChange}
        onBestChange={onBestChange}
      />
    </section>
  );
}
