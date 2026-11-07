"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { type FormEvent, useEffect, useState } from "react";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { Button } from "@/components/dockyard/ui/button";
import {
  completeSetup,
  getSetupStatus,
  patchInstanceSettings,
  verifyInstanceWildcard,
} from "@/lib/api";

export function SetupClient() {
  const router = useRouter();
  const params = useSearchParams();
  const tokenFromUrl = params.get("token") ?? "";
  const [token, setToken] = useState(tokenFromUrl);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [phase, setPhase] = useState<"account" | "apps-domain">("account");
  const [dashboardUrl, setDashboardUrl] = useState("");
  const [appsDomain, setAppsDomain] = useState("");

  useEffect(() => {
    getSetupStatus()
      .then((s) => {
        if (s.complete) router.replace("/sign-in");
      })
      .catch(() => {});
  }, [router]);

  useEffect(() => {
    if (tokenFromUrl) setToken(tokenFromUrl);
  }, [tokenFromUrl]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await completeSetup({ token, name, email, password });
      setPhase("apps-domain");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Setup failed");
    } finally {
      setLoading(false);
    }
  }

  async function onAppsDomainSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      if (dashboardUrl.trim() || appsDomain.trim()) {
        await patchInstanceSettings({
          public_url: dashboardUrl.trim() || undefined,
          acme_email: dashboardUrl.trim() ? email.trim() : undefined,
          wildcard_domain: appsDomain.trim() || undefined,
        });
        if (appsDomain.trim()) {
          await verifyInstanceWildcard().catch(() => {});
        }
      }
      router.replace("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save");
    } finally {
      setLoading(false);
    }
  }

  if (phase === "apps-domain") {
    return (
      <AuthLayout title="Default URLs for apps">
        <p className="mb-4 text-sm text-[var(--muted-foreground)]">
          Optional: use your own domain for app URLs, or skip and use sslip.io
          until you configure this under Server settings.
        </p>
        <form onSubmit={onAppsDomainSubmit} className="flex flex-col gap-4">
          <Field label="Dashboard URL (optional)">
            <input
              value={dashboardUrl}
              onChange={(e) => setDashboardUrl(e.target.value)}
              placeholder="https://panel.example.com"
              className="field-input"
            />
          </Field>
          <Field label="Apps domain (optional)">
            <input
              value={appsDomain}
              onChange={(e) => setAppsDomain(e.target.value)}
              placeholder="*.apps.example.com"
              className="field-input font-mono text-sm"
            />
          </Field>
          {error ? (
            <p
              className="text-sm"
              style={{ color: "var(--danger)" }}
              role="alert"
            >
              {error}
            </p>
          ) : null}
          <div className="mt-2 flex flex-col gap-2">
            <Button type="submit" disabled={loading} className="w-full">
              {loading ? "Saving…" : "Use my domain"}
            </Button>
            <Button
              type="button"
              variant="outline"
              className="w-full"
              disabled={loading}
              onClick={() => router.replace("/")}
            >
              Skip for now
            </Button>
          </div>
        </form>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout title="Create owner account">
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <Field label="Setup token">
          <input
            required
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="Paste token from installer"
            className="field-input"
          />
        </Field>
        <Field label="Name">
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Your name"
            className="field-input"
          />
        </Field>
        <Field label="Email">
          <input
            required
            type="email"
            autoComplete="email"
            placeholder="you@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="field-input"
          />
        </Field>
        <Field label="Password (min. 12 characters)">
          <input
            required
            type="password"
            minLength={12}
            autoComplete="new-password"
            placeholder="Choose a strong password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="field-input"
          />
        </Field>
        {error ? (
          <p
            className="text-sm"
            style={{ color: "var(--danger)" }}
            role="alert"
          >
            {error}
          </p>
        ) : null}
        <Button type="submit" disabled={loading} className="mt-2 w-full">
          {loading ? "Saving…" : "Finish setup"}
        </Button>
      </form>
    </AuthLayout>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <fieldset className="m-0 block border-0 p-0 text-sm">
      <legend
        className="mb-1 block w-full font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {label}
      </legend>
      {children}
    </fieldset>
  );
}
