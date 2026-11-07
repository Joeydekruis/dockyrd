import { Button } from "@/components/dockyrd/ui/button";

type SessionRowProps = {
  device: string;
  location?: string;
  lastActive: string;
  current?: boolean;
  onRevoke?: () => void;
};

export function SessionRow({
  device,
  location,
  lastActive,
  current,
  onRevoke,
}: SessionRowProps) {
  return (
    <div
      className="flex flex-wrap items-center gap-3 border-t py-4 first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <div className="min-w-0 flex-1">
        <p className="font-medium">
          {device}
          {current ? (
            <span
              className="ml-2 font-mono text-[11px] uppercase tracking-widest"
              style={{ color: "var(--link-blue)" }}
            >
              This device
            </span>
          ) : null}
        </p>
        <p
          className="mt-0.5 text-[13px]"
          style={{ color: "var(--muted-text)" }}
        >
          {[location, lastActive].filter(Boolean).join(" · ")}
        </p>
      </div>
      {!current && onRevoke ? (
        <Button variant="outline" onClick={onRevoke}>
          Revoke
        </Button>
      ) : null}
    </div>
  );
}
