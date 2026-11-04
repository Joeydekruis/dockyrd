import type { Deployment } from "@/lib/api";

export type StaticDeployStep = {
  id: string;
  title: string;
  detail: string;
};

export function staticDeploySteps(input: {
  repo: string;
  branch: string;
  installCmd: string;
  buildCmd: string;
  outputDir: string;
  defaultUrl: string;
}): StaticDeployStep[] {
  return [
    {
      id: "clone",
      title: "Clone repository",
      detail: `${input.repo} · ${input.branch}`,
    },
    {
      id: "install",
      title: "Install dependencies",
      detail: input.installCmd.trim() || "npm install",
    },
    {
      id: "build",
      title: "Build",
      detail: input.buildCmd.trim() || "npm run build",
    },
    {
      id: "publish",
      title: "Publish",
      detail: `Serve from ${input.outputDir.replace(/^\//, "")}/`,
    },
    {
      id: "certificate",
      title: "Issue certificate",
      detail: input.defaultUrl,
    },
  ];
}

/** Returns active step index 0–4, or 5 when finished, -1 on failure. */
export function resolveStaticDeployStep(
  deployment: Deployment | null,
  log: string,
): number {
  if (!deployment) return 0;
  const status = deployment.status.toLowerCase();
  if (status === "failed" || status === "cancelled") return -1;
  if (status === "live") return 5;

  const l = log.toLowerCase();
  const hasRelease = Boolean(deployment.release_path);

  let step = 0;
  if (status !== "queued") step = 1;
  if (l.includes("cloning") || l.includes("clone into"))
    step = Math.max(step, 1);
  if (l.includes("added") && l.includes("package")) step = Math.max(step, 2);
  if (
    l.includes("npm run build") ||
    l.includes("▶") ||
    (status === "building" && l.includes("build"))
  ) {
    step = Math.max(step, 2);
  }
  if (hasRelease) step = Math.max(step, 3);
  if (status === "building" && l.length > 80 && step < 2) step = 2;

  return Math.min(step, 4);
}

export type StepVisualState = "done" | "active" | "pending" | "failed";

export function stepVisualState(
  index: number,
  activeIndex: number,
  deployment: Deployment | null,
): StepVisualState {
  if (!deployment) return index === 0 ? "active" : "pending";
  const status = deployment.status.toLowerCase();
  if (status === "failed" || status === "cancelled") {
    return index <= Math.max(0, activeIndex) ? "failed" : "pending";
  }
  if (activeIndex >= 5) return "done";
  if (index < activeIndex) return "done";
  if (index === activeIndex) return "active";
  return "pending";
}
