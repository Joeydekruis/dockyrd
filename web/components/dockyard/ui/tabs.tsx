"use client";

type TabItem = {
  id: string;
  label: string;
};

type TabsProps = {
  tabs: TabItem[];
  activeId: string;
  onChange: (id: string) => void;
};

export function Tabs({ tabs, activeId, onChange }: TabsProps) {
  return (
    <div
      role="tablist"
      aria-label="Sections"
      className="flex flex-wrap gap-6 border-b"
      style={{ borderColor: "var(--border)" }}
    >
      {tabs.map((tab) => {
        const active = tab.id === activeId;
        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={active}
            className="dockyrd-page-tab -mb-px border-b-2 pb-3 text-sm font-medium"
            style={{
              borderColor: active ? "var(--ink-navy)" : "transparent",
              color: active ? "var(--ink-navy)" : "var(--muted-text)",
            }}
            onClick={() => onChange(tab.id)}
          >
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
