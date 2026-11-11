export type DeployBreakMountOptions = {
  releaseLabel?: string;
  serviceLabel?: string;
  initialBest?: number;
  deployProgress?: number;
  deployDone?: boolean;
  onScoreChange?: (score: number) => void;
  onBestChange?: (best: number) => void;
};

export function mountDeployBreak(
  canvas: HTMLCanvasElement,
  root: HTMLElement,
  options: DeployBreakMountOptions,
): { getScore: () => number; getBest: () => number };
