type AvatarProps = {
  initials: string;
  size?: "sm" | "md" | "lg";
};

const sizeClass = {
  sm: "size-8 text-xs",
  md: "size-10 text-sm",
  lg: "size-12 text-base",
};

export function Avatar({ initials, size = "md" }: AvatarProps) {
  return (
    <span
      className={`inline-flex shrink-0 items-center justify-center rounded-full font-semibold uppercase ${sizeClass[size]}`}
      style={{ background: "var(--brand-soft)", color: "var(--link-blue)" }}
      aria-hidden
    >
      {initials.slice(0, 2)}
    </span>
  );
}
