type DockyrdMarkProps = {
  className?: string;
};

/** Colored cube + white wordmark for auth gradient panel. */
export function DockyrdMark({ className = "" }: DockyrdMarkProps) {
  const maskId = "dockyrd-mark-icon";
  return (
    <span className={`inline-flex items-center gap-3 ${className}`.trim()}>
      <svg
        viewBox="0 0 64 64"
        width={40}
        height={40}
        aria-hidden
        className="shrink-0"
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
              d="M8 22 L32 36 L56 22 M32 36 V52"
              fill="none"
              stroke="#000"
              strokeWidth="4.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </mask>
        </defs>
        <g mask={`url(#${maskId})`} strokeWidth={3} strokeLinejoin="round">
          <path
            d="M32 8 L56 22 V46 q-6 -7 -12 0 t-12 0 t-12 0 t-12 0 V22 Z"
            fill="#8EBEFF"
            stroke="#8EBEFF"
          />
          <path
            d="M32 8 L56 22 L32 36 L8 22 Z"
            fill="#CFE3FF"
            stroke="#CFE3FF"
          />
        </g>
      </svg>
      <span className="text-[28px] font-semibold tracking-[-0.04em] text-white lowercase">
        dockyrd
      </span>
    </span>
  );
}
