type HeroPromoCardProps = {
  badge: string;
  title: string;
  description: string;
};

/** Gradient promo block (object storage empty state, design artboards). */
export function HeroPromoCard({ badge, title, description }: HeroPromoCardProps) {
  return (
    <section
      className="flex min-h-[200px] flex-col justify-end rounded-[20px] p-8 text-white"
      style={{ background: "var(--hero-gradient)" }}
      aria-label={title}
    >
      <span className="auth-brand-badge mb-4">{badge}</span>
      <h2 className="m-0 max-w-lg text-[36px] font-semibold leading-[1.08] tracking-[-0.04em]">
        {title}
      </h2>
      <p className="mt-3 max-w-lg text-sm leading-relaxed text-white/90">
        {description}
      </p>
    </section>
  );
}
