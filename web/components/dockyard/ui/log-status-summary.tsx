import { Chip } from "@/components/dockyard/ui/chip";

type LogStatusSummaryProps = {
  counts: { "2xx"?: number; "3xx"?: number; "4xx"?: number; "5xx"?: number };
};

export function LogStatusSummary({ counts }: LogStatusSummaryProps) {
  const entries = (
    [
      ["2xx", counts["2xx"]],
      ["3xx", counts["3xx"]],
      ["4xx", counts["4xx"]],
      ["5xx", counts["5xx"]],
    ] as const
  ).filter(([, n]) => n != null && n > 0);
  return (
    <section className="flex flex-wrap gap-2" aria-label="Status code summary">
      {entries.map(([label, n]) => (
        <span key={label} className="inline-flex items-center gap-1.5">
          <Chip>{label}</Chip>
          <span className="font-mono text-[13px] tabular-nums">{n}</span>
        </span>
      ))}
    </section>
  );
}
