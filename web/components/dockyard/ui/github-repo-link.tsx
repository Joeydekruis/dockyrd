import { GitHubMarkIcon } from "@/components/dockyard/ui/github-mark-icon";

type GitHubRepoLinkProps = {
  fullName: string;
  href: string;
};

export function GitHubRepoLink({ fullName, href }: GitHubRepoLinkProps) {
  const display = fullName.trim().replace(/^github\.com\//i, "");
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="dockyrd-surface-interactive dockyrd-focus-ring inline-flex max-w-full items-center gap-2 rounded-lg border px-3 py-1.5 text-[13px] font-medium"
      style={{
        borderColor: "var(--border)",
        background: "var(--surface)",
        color: "var(--ink-navy)",
      }}
    >
      <GitHubMarkIcon className="size-4 shrink-0" />
      <span className="truncate font-mono">{display}</span>
    </a>
  );
}
