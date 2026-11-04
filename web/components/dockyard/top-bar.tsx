import { AnnouncementBanner } from "@/components/dockyard/ui/announcement-banner";

const BANNER_HEIGHT_PX = 44;

type TopBarProps = {
  className?: string;
};

/** Fixed announcement strip in the main column (not over the sidebar). */
export function TopBar({ className = "" }: TopBarProps) {
  return (
    <div
      className={className}
      style={{ height: BANNER_HEIGHT_PX, minHeight: BANNER_HEIGHT_PX }}
    >
      <AnnouncementBanner />
    </div>
  );
}

export { BANNER_HEIGHT_PX };
