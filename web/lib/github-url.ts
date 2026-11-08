/** `owner/repo` → https://github.com/owner/repo */
export function githubRepoWebUrl(fullName: string): string | undefined {
  const repo = fullName.trim().replace(/^github\.com\//i, "");
  const parts = repo.split("/").filter(Boolean);
  if (parts.length !== 2) return undefined;
  return `https://github.com/${parts[0]}/${parts[1]}`;
}

export function githubCommitWebUrl(
  fullName: string,
  sha: string,
): string | undefined {
  const base = githubRepoWebUrl(fullName);
  const s = sha.trim();
  if (!base || !s) return undefined;
  return `${base}/commit/${s}`;
}
