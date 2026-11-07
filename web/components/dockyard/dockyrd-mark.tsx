import { DockyrdCubeIcon } from "@/components/dockyard/dockyrd-cube-icon";

type DockyrdMarkProps = {
  className?: string;
};

/** Colored cube + white wordmark for auth gradient panel. */
export function DockyrdMark({ className = "" }: DockyrdMarkProps) {
  return (
    <span
      className={`dockyrd-logo inline-flex items-center gap-3 ${className}`.trim()}
    >
      <DockyrdCubeIcon size={40} dobber className="shrink-0" />
      <span className="text-[28px] font-semibold tracking-[-0.04em] text-white lowercase">
        dockyrd
      </span>
    </span>
  );
}
