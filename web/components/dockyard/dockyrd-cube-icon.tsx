import { useId } from "react";

const WAVE_CLIP_D =
  "M-88 46 q6 7 12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 t12 0 V0 H-88 Z";

const SIDE_DOBBER_D = "M32 8 L56 22 V70 H8 V22 Z";
const TOP_D = "M32 8 L56 22 L32 36 L8 22 Z";
const SIDE_STATIC_D =
  "M32 8 L56 22 V46 q-6 -7 -12 0 t-12 0 t-12 0 t-12 0 V22 Z";
const MASK_OUTLINE_D = "M8 22 L32 36 L56 22 M32 36 V58";
const MASK_OUTLINE_STATIC_D = "M8 22 L32 36 L56 22 M32 36 V52";

type DockyrdCubeIconProps = {
  size?: number;
  className?: string;
  /** Wave clip + taller sides for logo hover “dobber”. */
  dobber?: boolean;
  idPrefix?: string;
};

export function DockyrdCubeIcon({
  size = 64,
  className = "",
  dobber = false,
  idPrefix,
}: DockyrdCubeIconProps) {
  const reactId = useId().replace(/:/g, "");
  const prefix = idPrefix ?? reactId;
  const maskId = `${prefix}-dockyrd-mask`;
  const clipId = `${prefix}-dockyrd-clip`;
  const viewBox = dobber ? "6 6 52 46" : "0 0 64 64";

  return (
    <svg
      viewBox={viewBox}
      width={size}
      height={size}
      aria-hidden
      className={className}
    >
      <defs>
        <mask
          id={maskId}
          maskUnits="userSpaceOnUse"
          x="0"
          y="0"
          width="64"
          height="64"
        >
          <rect width="64" height="64" fill="#fff" />
          <path
            d={dobber ? MASK_OUTLINE_D : MASK_OUTLINE_STATIC_D}
            fill="none"
            stroke="#000"
            strokeWidth="4.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </mask>
        {dobber ? (
          <clipPath id={clipId}>
            <path className="dockyrd-wave" d={WAVE_CLIP_D} />
          </clipPath>
        ) : null}
      </defs>
      <g className={dobber ? "dockyrd-mark" : undefined}>
        <g mask={`url(#${maskId})`} strokeWidth={3} strokeLinejoin="round">
          {dobber ? (
            <g clipPath={`url(#${clipId})`}>
              <path d={SIDE_DOBBER_D} fill="#8EBEFF" stroke="#8EBEFF" />
            </g>
          ) : (
            <path d={SIDE_STATIC_D} fill="#8EBEFF" stroke="#8EBEFF" />
          )}
          <path d={TOP_D} fill="#CFE3FF" stroke="#CFE3FF" />
        </g>
      </g>
    </svg>
  );
}
