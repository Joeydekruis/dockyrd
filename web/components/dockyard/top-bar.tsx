import { AnnouncementBanner } from "@/components/dockyard/ui/announcement-banner";

export const BANNER_HEIGHT_PX = 44;

type TopBarProps = {
  className?: string;
};

/** Fixed announcement strip spanning the full app width (above sidebar and header). */
export function TopBar({ className = "" }: TopBarProps) {
  return (
    <div
      className={className}
      style={{
        height: BANNER_HEIGHT_PX,
        minHeight: BANNER_HEIGHT_PX,
      }}
    >
      <AnnouncementBanner />
    </div>
  );
}
