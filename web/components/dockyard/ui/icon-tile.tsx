import type { LucideIcon } from "lucide-react";

type IconTileProps = {
  icon: LucideIcon;
  label?: string;
  size?: "sm" | "md";
};

const box = { sm: "size-9 rounded-lg", md: "size-10 rounded-xl" };
const icon = { sm: "size-4", md: "size-5" };

export function IconTile({ icon: Icon, label, size = "md" }: IconTileProps) {
  return (
    <div
      className={`flex shrink-0 items-center justify-center ${box[size]}`}
      style={{ background: "var(--brand-soft)", color: "var(--link-blue)" }}
      title={label}
    >
      <Icon strokeWidth={1.8} className={icon[size]} aria-hidden />
    </div>
  );
}
