"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  getInstanceSettings,
  type InstanceSettings,
  patchInstanceSettings,
  verifyInstanceWildcard,
} from "@/lib/api";
import { copyText } from "@/lib/domain-dns";

type AppsOnYourDomainSectionProps = {
  instance: InstanceSettings;
  onInstanceChange: (next: InstanceSettings) => void;
  onMessage: (message: string, tone: "success" | "warn" | "info") => void;
};

function statusPill(instance: InstanceSettings): {
  label: string;
  variant: "ok" | "warn" | "idle" | "err";
} {
  if (instance.wildcard_verified) {
    return { label: "Verified", variant: "ok" };
  }
  const dns = (instance.wildcard_dns_status ?? "").toLowerCase();
  if (!instance.wildcard_domain?.trim()) {
    return { label: "Not set up", variant: "idle" };
  }
  if (dns === "wrong_target") {
    return { label: "Check failed", variant: "err" };
  }
  if (dns === "pending") {
    return { label: "Unverified", variant: "warn" };
  }
  return { label: "Unverified", variant: "warn" };
}

export function AppsOnYourDomainSection({
  instance,
  onInstanceChange,
  onMessage,
}: AppsOnYourDomainSectionProps) {
  const [wildcard, setWildcard] = useState(instance.wildcard_domain ?? "");
  const [verifying, setVerifying] = useState(false);

  useEffect(() => {
    setWildcard(instance.wildcard_domain ?? "");
  }, [instance.wildcard_domain]);

  const pill = statusPill(instance);
  const probe = instance.wildcard_probe_hostname ?? "";
  const ipv4 = instance.public_ipv4 ?? "";
  const suffix = wildcard.replace(/^\*\./, "");

  const saveWildcard = useCallback(async () => {
    try {
      const updated = await patchInstanceSettings({
        wildcard_domain: wildcard,
      });
      onInstanceChange(updated);
      onMessage("Apps domain saved.", "success");
    } catch (e) {
      onMessage(e instanceof Error ? e.message : "Could not save", "warn");
    }
  }, [wildcard, onInstanceChange, onMessage]);

  const verify = useCallback(async () => {
    setVerifying(true);
    try {
      await saveWildcard();
      const res = await verifyInstanceWildcard();
      const refreshed = await getInstanceSettings();
      onInstanceChange(refreshed);
      if (res.wildcard_verified) {
        onMessage("Apps domain verified.", "success");
      } else {
        onMessage(
          "DNS not pointing here yet. You can deploy; default URLs use sslip.io until verified.",
          "info",
        );
      }
    } catch (e) {
      onMessage(e instanceof Error ? e.message : "Verify failed", "warn");
    } finally {
      setVerifying(false);
    }
  }, [saveWildcard, onInstanceChange, onMessage]);

  return (
    <section
      aria-label="Apps on your domain"
      className="mt-6 space-y-4 border-t border-[var(--border)] pt-6"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 className="text-base font-medium">Apps on your domain</h3>
          <p className="mt-1 text-sm text-[var(--muted-foreground)]">
            Point a wildcard DNS record at this server so new apps get default
            URLs on your domain. Until verified, dockyrd uses sslip.io.
          </p>
        </div>
        <StatusBadge label={pill.label} variant={pill.variant} />
      </div>

      {instance.suggested_apps_domain && !wildcard.trim() ? (
        <Button
          type="button"
          variant="outline"
          className="text-[13px]"
          onClick={() => setWildcard(instance.suggested_apps_domain ?? "")}
        >
          Use {instance.suggested_apps_domain}
        </Button>
      ) : null}

      <label className="block text-sm">
        <span className="text-[var(--muted-foreground)]">Apps domain</span>
        <input
          className="field-input mt-1 w-full font-mono text-sm"
          value={wildcard}
          onChange={(e) => setWildcard(e.target.value)}
          placeholder="*.apps.example.com"
        />
      </label>

      {wildcard.trim() ? (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--subtle)] p-4 text-sm">
          <p className="mb-2 font-medium">DNS record</p>
          <div className="grid gap-2 font-mono text-[12px] sm:grid-cols-3">
            <span>Type: A</span>
            <span>Name: *.{suffix || "apps.example.com"}</span>
            <span className="break-all">Value: {ipv4 || "—"}</span>
          </div>
          {probe ? (
            <p className="mt-3 text-[var(--muted-foreground)]">
              Probe hostname:{" "}
              <button
                type="button"
                className="font-mono underline"
                onClick={() => void copyText(probe)}
              >
                {probe}
              </button>
            </p>
          ) : null}
          {!instance.wildcard_verified ? (
            <p className="mt-2 text-[var(--muted-foreground)]">
              You can deploy now. Default URLs use sslip.io until this is
              verified.
            </p>
          ) : null}
        </div>
      ) : null}

      <div className="flex flex-wrap gap-2">
        {!instance.wildcard_verified && wildcard.trim() ? (
          <Button
            type="button"
            disabled={verifying}
            onClick={() => void verify()}
          >
            {verifying ? "Checking DNS…" : "Verify"}
          </Button>
        ) : (
          <Button
            type="button"
            variant="outline"
            onClick={() => void saveWildcard()}
          >
            Save apps domain
          </Button>
        )}
        {instance.wildcard_verified ? (
          <Button
            type="button"
            variant="outline"
            disabled={verifying}
            onClick={() => void verify()}
          >
            Re-check
          </Button>
        ) : null}
        <Link
          href="https://dockyrd.io/docs/domains.html"
          className="self-center text-sm underline"
          target="_blank"
          rel="noopener noreferrer"
        >
          Show full DNS help
        </Link>
      </div>
    </section>
  );
}
