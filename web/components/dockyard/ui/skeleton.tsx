type SkeletonProps = {
  lines?: number;
};

const lineKeys = ["sk-a", "sk-b", "sk-c", "sk-d", "sk-e"] as const;

export function Skeleton({ lines = 3 }: SkeletonProps) {
  const count = Math.min(Math.max(lines, 1), lineKeys.length);
  return (
    <div className="flex flex-col gap-2" aria-hidden>
      {lineKeys.slice(0, count).map((key, i) => (
        <div
          key={key}
          className="h-3 animate-pulse rounded-md"
          style={{
            background: "var(--hover-row)",
            width: i === count - 1 ? "60%" : "100%",
          }}
        />
      ))}
    </div>
  );
}
