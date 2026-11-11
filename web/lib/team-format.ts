export function initialsFromName(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function formatRelativeTime(unixSec: number, nowSec = Date.now() / 1000): string {
  const diff = Math.max(0, nowSec - unixSec);
  if (diff < 120) return "Active now";
  if (diff < 3600) {
    const m = Math.round(diff / 60);
    return `Active ${m} min ago`;
  }
  if (diff < 86400) {
    const h = Math.round(diff / 3600);
    return `Active ${h} h ago`;
  }
  const d = Math.round(diff / 86400);
  if (d === 1) return "Active yesterday";
  return `Active ${d} days ago`;
}

export function formatInviteExpiry(expiresAt: number, nowSec = Date.now() / 1000): {
  expired: boolean;
  status: string;
  detail: string;
} {
  if (expiresAt <= nowSec) {
    const days = Math.max(1, Math.round((nowSec - expiresAt) / 86400));
    return {
      expired: true,
      status: "Expired",
      detail: `Expired ${days} day${days === 1 ? "" : "s"} ago`,
    };
  }
  const left = expiresAt - nowSec;
  const days = Math.ceil(left / 86400);
  return {
    expired: false,
    status: "Pending",
    detail: `Expires in ${days} day${days === 1 ? "" : "s"}`,
  };
}
