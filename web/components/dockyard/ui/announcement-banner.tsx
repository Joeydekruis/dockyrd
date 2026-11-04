import { ArrowRight } from "lucide-react";
import { Chip } from "@/components/dockyard/ui/chip";
import { PUBLIC_SITE_ORIGIN } from "@/lib/public-links";

export function AnnouncementBanner() {
  return (
    <section
      aria-label="Announcement"
      className="flex h-full min-h-11 w-full items-center gap-4 border-b px-6 py-2 text-sm"
      style={{
        background: "var(--banner-blue)",
        borderColor: "var(--border)",
        color: "var(--ink-navy)",
      }}
    >
      <div className="flex w-[140px] shrink-0 justify-start">
        <Chip>Open source</Chip>
      </div>
      <p className="min-w-0 flex-1 text-center text-sm leading-snug">
        One lightweight binary on your own server. No lock-in, no mess.
      </p>
      <div className="flex w-[140px] shrink-0 justify-end">
        <a
          href={PUBLIC_SITE_ORIGIN}
          target="_blank"
          rel="noopener noreferrer"
          aria-label="Learn more about Dockyrd"
          className="inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-[13px] font-medium no-underline"
          style={{
            background: "rgba(255, 255, 255, 0.35)",
            color: "var(--ink-navy)",
          }}
        >
          Learn more
          <ArrowRight strokeWidth={2} className="size-3.5" aria-hidden />
        </a>
      </div>
    </section>
  );
}
