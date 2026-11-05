"use client";

import type { ProjectLinkEdge } from "@/lib/api";

type ProjectLinksMapProps = {
  links: ProjectLinkEdge[];
};

/** Deterministic three-column link map (no graph library). */
export function ProjectLinksMap({ links }: ProjectLinksMapProps) {
  const uniqueApps = links
    .map((l) => ({ id: l.from_service_id, name: l.from_name }))
    .filter((v, i, a) => a.findIndex((x) => x.id === v.id) === i);

  const uniqueData = links
    .map((l) => ({
      id: l.to_service_id,
      name: l.to_name,
      kind: l.to_kind,
    }))
    .filter((v, i, a) => a.findIndex((x) => x.id === v.id) === i);

  const w = 720;
  const h = Math.max(
    280,
    80 + Math.max(uniqueApps.length, uniqueData.length) * 48,
  );

  if (links.length === 0) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
        No connections yet. Link apps to databases or storage from an
        application page.
      </p>
    );
  }

  return (
    <div
      className="overflow-x-auto rounded-[20px] border p-4"
      style={{ borderColor: "var(--border)", background: "var(--surface)" }}
    >
      <svg
        viewBox={`0 0 ${w} ${h}`}
        className="w-full min-w-[640px]"
        role="img"
        aria-label="Project resource link map"
      >
        <defs>
          <pattern
            id="dots"
            width="12"
            height="12"
            patternUnits="userSpaceOnUse"
          >
            <circle cx="1" cy="1" r="1" fill="var(--border)" />
          </pattern>
        </defs>
        <rect width={w} height={h} fill="url(#dots)" />
        <text
          x={80}
          y={32}
          fill="var(--muted-text)"
          className="text-[11px] font-mono uppercase"
        >
          Internet
        </text>
        <text
          x={300}
          y={32}
          fill="var(--muted-text)"
          className="text-[11px] font-mono uppercase"
        >
          Applications
        </text>
        <text
          x={520}
          y={32}
          fill="var(--muted-text)"
          className="text-[11px] font-mono uppercase"
        >
          Data
        </text>
        {uniqueApps.map((node, i) => {
          const y = 60 + i * 48;
          return (
            <g key={node.id}>
              <rect
                x={260}
                y={y}
                width={160}
                height={36}
                rx={8}
                fill="var(--brand-soft)"
                stroke="var(--border)"
              />
              <text
                x={270}
                y={y + 22}
                fontSize={12}
                fontWeight={600}
                fill="var(--ink-navy)"
              >
                {node.name}
              </text>
            </g>
          );
        })}
        {uniqueData.map((node, i) => {
          const y = 60 + i * 48;
          return (
            <g key={node.id}>
              <rect
                x={500}
                y={y}
                width={160}
                height={36}
                rx={8}
                fill="var(--surface)"
                stroke="var(--border)"
              />
              <text
                x={510}
                y={y + 22}
                fontSize={12}
                fontWeight={600}
                fill="var(--ink-navy)"
              >
                {node.name}
              </text>
            </g>
          );
        })}
        {links.map((edge) => {
          const fromIdx = uniqueApps.findIndex(
            (a) => a.id === edge.from_service_id,
          );
          const toIdx = uniqueData.findIndex(
            (d) => d.id === edge.to_service_id,
          );
          if (fromIdx < 0 || toIdx < 0) return null;
          const y1 = 60 + fromIdx * 48 + 18;
          const y2 = 60 + toIdx * 48 + 18;
          return (
            <line
              key={edge.id}
              x1={420}
              y1={y1}
              x2={500}
              y2={y2}
              stroke="var(--muted-text)"
              strokeWidth={1.5}
              strokeDasharray="4 4"
            />
          );
        })}
      </svg>
      <ul className="sr-only">
        {links.map((l) => (
          <li key={l.id}>
            {l.from_name} linked to {l.to_name}
          </li>
        ))}
      </ul>
    </div>
  );
}
