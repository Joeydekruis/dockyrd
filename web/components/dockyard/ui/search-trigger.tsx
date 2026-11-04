type SearchTriggerProps = {
  onClick: () => void;
};

export function SearchTrigger({ onClick }: SearchTriggerProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="dockyard-search-trigger"
      aria-label="Search pages, resources and actions"
    >
      <span>Search…</span>
      <kbd className="font-mono text-[11px]">⌘K</kbd>
    </button>
  );
}
