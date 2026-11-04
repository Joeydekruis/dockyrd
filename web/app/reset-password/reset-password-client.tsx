"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { type FormEvent, useState } from "react";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { Button } from "@/components/dockyard/ui/button";
import { resetPassword } from "@/lib/api";

export function ResetPasswordClient() {
  const router = useRouter();
  const params = useSearchParams();
  const token = params.get("token") ?? "";
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (!token) return;
    setLoading(true);
    setError(null);
    try {
      await resetPassword(token, password);
      router.replace("/sign-in");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Reset failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout title="Reset password">
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <label className="block text-sm">
          <span
            className="mb-1 block font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            New password (12+ characters)
          </span>
          <input
            required
            type="password"
            minLength={12}
            autoComplete="new-password"
            placeholder="At least 12 characters"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="field-input"
          />
        </label>
        {error ? (
          <p
            className="text-sm"
            style={{ color: "var(--danger)" }}
            role="alert"
          >
            {error}
          </p>
        ) : null}
        <Button type="submit" disabled={loading || !token} className="w-full">
          {loading ? "Saving…" : "Update password"}
        </Button>
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
