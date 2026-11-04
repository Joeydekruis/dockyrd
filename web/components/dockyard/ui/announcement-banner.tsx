import { X } from "lucide-react";
import { Chip } from "@/components/dockyard/ui/chip";

type AnnouncementBannerProps = {
  onDismiss?: () => void;
};

export function AnnouncementBanner({ onDismiss }: AnnouncementBannerProps) {
  return (
    <section
      aria-label="Announcement"
      className="flex min-h-11 items-center justify-between gap-4 border-b px-8 py-2 text-sm"
      style={{
        background: "var(--banner-blue)",
        borderColor: "var(--border)",
        color: "var(--ink-navy)",
      }}
    >
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <Chip>Open source</Chip>
        <span className="truncate">
          Dockyrd is self-hosted deploy infrastructure — star us on GitHub.
        </span>
      </div>
      {onDismiss ? (
        <button
          type="button"
          onClick={onDismiss}
          className="dockyard-icon-btn shrink-0"
          aria-label="Dismiss announcement"
        >
          <X strokeWidth={1.8} className="size-4" aria-hidden />
        </button>
      ) : null}
    </section>
  );
}
