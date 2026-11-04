import { Loader2 } from "lucide-react";

type SpinnerProps = {
  label?: string;
  size?: "sm" | "md";
};

const sizes = { sm: "size-4", md: "size-6" };

export function Spinner({ label = "Loading", size = "md" }: SpinnerProps) {
  return (
    <span className="inline-flex items-center gap-2" role="status">
      <Loader2
        strokeWidth={2}
        className={`animate-spin ${sizes[size]}`}
        style={{ color: "var(--accent-blue)" }}
        aria-hidden
      />
      <span className="sr-only">{label}</span>
    </span>
  );
}
