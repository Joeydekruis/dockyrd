"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Button } from "@/components/dockyrd/ui/button";
import { getInstanceSettings, getMe } from "@/lib/api";

const DISMISS_KEY = "dockyrd_apps_domain_nudge_dismissed";

type AppsDomainNudgeProps = {
  scope?: "service";
};

export function AppsDomainNudge({ scope }: AppsDomainNudgeProps) {
  const [show, setShow] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const me = await getMe();
        if (me.role !== "owner" && me.role !== "admin") return;
        if (
          typeof window !== "undefined" &&
          localStorage.getItem(DISMISS_KEY)
        ) {
          return;
        }
        const inst = await getInstanceSettings();
        if (cancelled) return;
        if (inst.wildcard_verified) return;
        if (!inst.wildcard_domain?.trim()) {
          setMessage(
            "Set an apps domain to get default URLs on your own hostname.",
          );
        } else {
          setMessage(
            "Verify your apps domain DNS to use it for new default URLs.",
          );
        }
        setShow(true);
      } catch {
        /* not admin or offline */
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  if (!show) return null;

  return (
    <div
      className="mb-4 flex flex-col gap-3 rounded-[16px] border border-[var(--border)] bg-[var(--subtle)] px-4 py-3 sm:flex-row sm:items-center sm:justify-between"
      role="status"
    >
      <p className="text-sm">
        {scope === "service" ? `${message} ` : message}
        {scope === "service" ? "Custom domains still work anytime." : null}
      </p>
      <div className="flex flex-wrap gap-2">
        <Link
          href="/settings/server/general"
          className="btn btn-outline min-h-9 px-3 py-1.5 text-[13px] inline-flex items-center"
        >
          Server settings
        </Link>
        <Button
          type="button"
          variant="ghost"
          className="min-h-9 text-[13px]"
          onClick={() => {
            localStorage.setItem(DISMISS_KEY, "1");
            setShow(false);
          }}
        >
          Dismiss
        </Button>
      </div>
    </div>
  );
}
