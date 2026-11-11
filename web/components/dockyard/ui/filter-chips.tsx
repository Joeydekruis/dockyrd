import { Chip } from "@/components/dockyard/ui/chip";

type FilterChipsProps = {
  options: { value: string; label: string }[];
  active: string;
  onChange: (value: string) => void;
};

export function FilterChips({ options, active, onChange }: FilterChipsProps) {
  return (
    <div className="flex flex-wrap gap-2" role="tablist" aria-label="Filters">
      {options.map((opt) => {
        const isActive = opt.value === active;
        return (
          <button
            key={opt.value}
            type="button"
            role="tab"
            aria-selected={isActive}
            onClick={() => onChange(opt.value)}
            className="dockyrd-pill-tab rounded-md border-0 p-0 transition-opacity"
            style={{ opacity: isActive ? 1 : 0.65 }}
          >
            <Chip>{opt.label}</Chip>
          </button>
        );
      })}
    </div>
  );
}
