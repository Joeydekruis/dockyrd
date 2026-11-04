type ChartPanelProps = {
  title: string;
  subtitle?: string;
  height?: number;
};

export function ChartPanel({ title, subtitle, height = 160 }: ChartPanelProps) {
  const w = 400;
  const h = height;
  const points =
    "0,120 40,100 80,110 120,70 160,85 200,45 240,60 280,30 320,50 360,20 400,35";
  return (
    <article
      className="rounded-[20px] border p-5"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <p className="font-semibold">{title}</p>
      {subtitle ? (
        <p className="mt-1 text-[13px]" style={{ color: "var(--muted-text)" }}>
          {subtitle}
        </p>
      ) : null}
      <svg
        viewBox={`0 0 ${w} ${h}`}
        className="mt-4 w-full"
        role="img"
        aria-label={`${title} chart placeholder`}
        preserveAspectRatio="none"
      >
        <defs>
          <linearGradient id="chart-fill" x1="0" y1="0" x2="0" y2="1">
            <stop
              offset="0%"
              stopColor="var(--accent-blue)"
              stopOpacity="0.25"
            />
            <stop
              offset="100%"
              stopColor="var(--accent-blue)"
              stopOpacity="0"
            />
          </linearGradient>
        </defs>
        <polygon
          fill="url(#chart-fill)"
          points={`${points} ${w},${h} 0,${h}`}
        />
        <polyline
          fill="none"
          stroke="var(--accent-blue)"
          strokeWidth="2"
          points={points}
        />
      </svg>
    </article>
  );
}
