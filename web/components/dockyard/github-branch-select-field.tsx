"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { SearchSelectField } from "@/components/dockyard/ui/search-select-field";
import { listGitHubBranches, parseGitHubOwnerRepo } from "@/lib/api";

type GitHubBranchSelectFieldProps = {
  label?: string;
  repo: string;
  value: string;
  onChange: (branch: string) => void;
  /** Used when the current branch is missing from the fetched list (e.g. repo default). */
  preferredBranch?: string;
  hint?: string;
  disabled?: boolean;
};

export function GitHubBranchSelectField({
  label = "Branch",
  repo,
  value,
  onChange,
  preferredBranch,
  hint = "Pushes to this branch deploy.",
  disabled = false,
}: GitHubBranchSelectFieldProps) {
  const [branchNames, setBranchNames] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const valueRef = useRef(value);
  valueRef.current = value;

  const repoKey = repo.trim();
  const parsed = useMemo(() => parseGitHubOwnerRepo(repoKey), [repoKey]);

  useEffect(() => {
    if (!parsed) {
      setBranchNames([]);
      setLoadError(null);
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    setLoadError(null);

    void listGitHubBranches(parsed.owner, parsed.repo)
      .then(({ branches }) => {
        if (cancelled) return;
        const names = branches.map((b) => b.name);
        setBranchNames(names);
        if (names.length === 0) return;

        const current = valueRef.current.trim();
        if (current && names.includes(current)) return;

        const preferred = preferredBranch?.trim();
        if (preferred && names.includes(preferred)) {
          onChange(preferred);
          return;
        }
        onChange(names[0]);
      })
      .catch(() => {
        if (!cancelled) {
          setLoadError("Could not load branches from GitHub.");
          setBranchNames([]);
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [parsed, preferredBranch, onChange]);

  const noRepo = !parsed;
  const fieldDisabled = disabled || noRepo || loading;

  const placeholder = noRepo
    ? "Choose a repository first"
    : loading
      ? "Loading branches…"
      : "Choose a branch";

  const options = branchNames.map((name) => ({ value: name, label: name }));

  return (
    <SearchSelectField
      label={label}
      placeholder={placeholder}
      searchPlaceholder="Search branches…"
      emptyMessage="No branches found."
      value={fieldDisabled && !value ? "" : value}
      onChange={onChange}
      hint={loadError ?? hint}
      error={loadError ?? undefined}
      disabled={fieldDisabled}
      options={options}
    />
  );
}
