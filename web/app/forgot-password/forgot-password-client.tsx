"use client";

import Link from "next/link";
import { type FormEvent, useEffect, useState } from "react";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { getPublicSettings, requestPasswordReset } from "@/lib/api";

export function ForgotPasswordClient() {
  const [email, setEmail] = useState("");
  const [smtpConfigured, setSmtpConfigured] = useState<boolean | null>(null);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    void getPublicSettings()
      .then((s) => setSmtpConfigured(s.smtp_configured))
      .catch(() => setSmtpConfigured(false));
  }, []);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSent(false);
    try {
      await requestPasswordReset(email.trim());
      setSent(true);
    } catch {
      setError("Could not start reset");
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout title="Forgot password?">
      {smtpConfigured === false ? (
        <p className="mb-4 text-sm" style={{ color: "var(--muted-text)" }}>
          Email is not configured on this server. Ask an admin, or run{" "}
          <code className="rounded bg-black/5 px-1">
            sudo dockyard admin reset-password
          </code>{" "}
          on the host.
        </p>
      ) : (
        <p className="mb-4 text-sm" style={{ color: "var(--muted-text)" }}>
          Enter your account email. If it exists, we send a reset link (valid
          for one hour).
        </p>
      )}
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <label className="block text-sm">
          <span
            className="mb-1 block font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Email
          </span>
          <input
            required
            type="email"
            autoComplete="email"
            placeholder="you@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="field-input"
            disabled={smtpConfigured === false}
          />
        </label>
        {sent && smtpConfigured !== false ? (
          <p className="text-sm text-green-700" role="status">
            If an account exists for that address, check your inbox for a reset
            link.
          </p>
        ) : null}
        {error ? (
          <p
            className="text-sm"
            style={{ color: "var(--danger)" }}
            role="alert"
          >
            {error}
          </p>
        ) : null}
        <button
          type="submit"
          disabled={loading || smtpConfigured === false}
          className="min-h-11 rounded-[10px] font-mono text-xs uppercase tracking-wide text-white disabled:opacity-50"
          style={{ background: "var(--ink-navy)" }}
        >
          {loading ? "Please wait…" : "Request reset link"}
        </button>
      </form>
      <Link
        href="/sign-in"
        className="mt-6 inline-block text-sm"
        style={{ color: "var(--link-blue)" }}
      >
        Back to sign in
      </Link>
    </AuthLayout>
  );
}
