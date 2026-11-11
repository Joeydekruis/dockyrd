import { ArrowUpRight, BookOpen } from "lucide-react";
import {
  PUBLIC_DOCS_INTRO_URL,
  PUBLIC_GITHUB_REPO_URL,
} from "@/lib/public-links";

/** Permanent full-width announcement strip (not dismissible). */
export function AnnouncementBanner() {
  return (
    <section
      aria-label="Announcement"
      className="relative flex h-full min-h-11 w-full flex-wrap items-center justify-between gap-2 border-b px-6 py-2 text-sm sm:gap-4"
      style={{
        background: "var(--banner-blue)",
        borderColor: "var(--border)",
        color: "var(--ink-navy)",
      }}
    >
      <div className="flex min-w-0 flex-wrap items-center gap-2 sm:gap-3">
        <span
          className="inline-flex items-center rounded-md px-2 py-0.5 font-mono text-[11px] uppercase tracking-widest"
          style={{ background: "var(--ink-navy)", color: "#ffffff" }}
        >
          Open source
        </span>
        <p className="text-[13px] font-medium leading-snug">
          Self-hosted deploys on your own servers.
        </p>
      </div>
      <div className="flex flex-wrap items-center gap-1">
        <a
          href={PUBLIC_DOCS_INTRO_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex min-h-9 items-center gap-2 rounded-[10px] px-3 py-1.5 text-[13px] font-medium no-underline transition-colors hover:bg-white/50"
        >
          <BookOpen strokeWidth={1.8} className="size-4" aria-hidden />
          Documentation
          <ArrowUpRight strokeWidth={2} className="size-3.5" aria-hidden />
        </a>
        <a
          href={PUBLIC_GITHUB_REPO_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex min-h-9 items-center gap-2 rounded-[10px] px-3 py-1.5 text-[13px] font-medium no-underline transition-colors hover:bg-white/50"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden
          >
            <circle cx="6" cy="5" r="2.2" />
            <circle cx="6" cy="19" r="2.2" />
            <circle cx="18" cy="9" r="2.2" />
            <path d="M6 7.2v9.6M18 11.2v5.8M6 7.2L18 9" />
          </svg>
          GitHub
          <ArrowUpRight strokeWidth={2} className="size-3.5" aria-hidden />
        </a>
      </div>
    </section>
  );
}
