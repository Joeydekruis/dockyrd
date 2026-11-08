function initialsFromName(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  const first = parts[0];
  if (parts.length === 1) return (first ?? "?").slice(0, 2).toUpperCase();
  const last = parts[parts.length - 1];
  return `${first?.[0] ?? ""}${last?.[0] ?? ""}`.toUpperCase();
}

export function DeploymentAuthorAvatar({
  name,
  title,
}: {
  name: string;
  title?: string;
}) {
  const label = initialsFromName(name);
  return (
    <span
      className="inline-flex size-6 shrink-0 items-center justify-center rounded-full border text-[10px] font-semibold uppercase"
      style={{
        borderColor: "var(--border)",
        background: "var(--hover-row)",
        color: "var(--ink-navy)",
      }}
      title={title ?? name}
      aria-hidden
    >
      {label}
    </span>
  );
}
