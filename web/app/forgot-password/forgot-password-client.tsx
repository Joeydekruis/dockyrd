"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { requestPasswordReset } from "@/lib/api";

export function ForgotPasswordClient() {
  const [email, setEmail] = useState("");
  const [resetPath, setResetPath] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResetPath(null);
    try {
      const res = await requestPasswordReset(email.trim());
      if (res.reset_path) {
        setResetPath(res.reset_path);
      }
    } catch {
      setError("Could not start reset");
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout title="Forgot password?">
      <p className="mb-4 text-sm" style={{ color: "var(--muted-text)" }}>
        Email delivery is not configured yet. When your account exists, the reset link is shown
        here (self-hosted).
      </p>
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <label className="block text-sm">
          <span className="mb-1 block font-mono text-[11px] uppercase tracking-widest" style={{ color: "var(--muted-text)" }}>
            Email
          </span>
          <input
            required
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="field-input"
          />
        </label>
        {resetPath ? (
          <p className="text-sm text-green-700" role="status">
            Use this link within 1 hour:{" "}
            <Link href={resetPath} className="underline">
              Reset password
            </Link>
          </p>
        ) : null}
        {error ? (
          <p className="text-sm" style={{ color: "var(--danger)" }} role="alert">
            {error}
          </p>
        ) : null}
        <button
          type="submit"
          disabled={loading}
          className="min-h-11 rounded-[10px] font-mono text-xs uppercase tracking-wide text-white"
          style={{ background: "var(--ink-navy)" }}
        >
          {loading ? "Please wait…" : "Request reset link"}
        </button>
      </form>
      <Link href="/sign-in" className="mt-6 inline-block text-sm" style={{ color: "var(--link-blue)" }}>
        Back to sign in
      </Link>
    </AuthLayout>
  );
}
