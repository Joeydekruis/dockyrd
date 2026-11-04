import { DockyrdCubeIcon } from "@/components/dockyard/dockyrd-cube-icon";

const LOADER_CUBE_SLOTS = ["a", "b", "c", "d", "e"] as const;

type DockyrdPageLoadingProps = {
  label?: string;
  /** Full viewport centering vs in-page placeholder. */
  layout?: "page" | "inline";
  className?: string;
  cubeSize?: number;
};

export function DockyrdPageLoading({
  label = "Loading…",
  layout = "inline",
  className = "",
  cubeSize = 36,
}: DockyrdPageLoadingProps) {
  const loader = (
    <>
      <div className="dockyrd-loader" aria-hidden>
        {LOADER_CUBE_SLOTS.map((slot) => (
          <DockyrdCubeIcon
            key={slot}
            size={cubeSize}
            className="dockyrd-loader-cube"
          />
        ))}
      </div>
      <span className="sr-only">{label}</span>
      {layout === "page" ? (
        <p className="mt-6 text-sm" style={{ color: "var(--muted-text)" }}>
          {label}
        </p>
      ) : null}
    </>
  );

  if (layout === "page") {
    return (
      <div
        className="flex min-h-dvh flex-col items-center justify-center"
        role="status"
        aria-live="polite"
      >
        {loader}
      </div>
    );
  }

  return (
    <div
      className={`flex flex-col items-center py-8 ${className}`.trim()}
      role="status"
      aria-live="polite"
    >
      {loader}
    </div>
  );
}
