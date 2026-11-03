"use client";

import Link from "next/link";
import { FormEvent, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { acceptInvite, previewInvite } from "@/lib/api";

export function AcceptInviteClient() {
  const router = useRouter();
  const params = useSearchParams();
  const token = params.get("token") ?? "";
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!token) {
      setError("Missing invite token.");
      return;
    }
    previewInvite(token)
      .then((p) => {
        setEmail(p.email);
        setRole(p.role);
      })
      .catch(() => setError("This invite is invalid or expired."));
  }, [token]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (!token) return;
    setLoading(true);
    setError(null);
    try {
      await acceptInvite({ token, name, password });
      router.replace("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not accept invite");
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout title="Accept invite">
      {email ? (
        <p className="mb-4 text-sm" style={{ color: "var(--muted-text)" }}>
          Joining as <strong>{email}</strong>
          {role ? ` (${role})` : ""}
        </p>
      ) : null}
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <label className="block text-sm">
          <span className="mb-1 block font-mono text-[11px] uppercase tracking-widest" style={{ color: "var(--muted-text)" }}>
            Your name
          </span>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="field-input"
          />
        </label>
        <label className="block text-sm">
          <span className="mb-1 block font-mono text-[11px] uppercase tracking-widest" style={{ color: "var(--muted-text)" }}>
            Password (12+ characters)
          </span>
          <input
            required
            type="password"
            minLength={12}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="field-input"
          />
        </label>
        {error ? (
          <p className="text-sm" style={{ color: "var(--danger)" }} role="alert">
            {error}
          </p>
        ) : null}
        <button
          type="submit"
          disabled={loading || !token}
          className="min-h-11 rounded-[10px] font-mono text-xs uppercase tracking-wide text-white"
          style={{ background: "var(--ink-navy)" }}
        >
          {loading ? "Creating account…" : "Accept invite"}
        </button>
      </form>
      <p className="mt-6 text-sm">
        <Link href="/sign-in" style={{ color: "var(--link-blue)" }}>
          Back to sign in
        </Link>
      </p>
    </AuthLayout>
  );
}
