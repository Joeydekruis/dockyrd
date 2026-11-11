"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { type FormEvent, useEffect, useState } from "react";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { Button } from "@/components/dockyard/ui/button";
import { acceptInvite, previewInvite } from "@/lib/api";

export function AcceptInviteClient() {
  const router = useRouter();
  const params = useSearchParams();
  const token = params.get("token") ?? "";
  const [email, setEmail] = useState("");
  const [emailRequired, setEmailRequired] = useState(false);
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
        setEmailRequired(!p.email);
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
      await acceptInvite({
        token,
        name,
        password,
        ...(emailRequired || email ? { email: email.trim() } : {}),
      });
      router.replace("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not accept invite");
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout title="Accept invite">
      {email && !emailRequired ? (
        <p className="mb-4 text-sm" style={{ color: "var(--muted-text)" }}>
          Joining as <strong>{email}</strong>
          {role ? ` (${role})` : ""}
        </p>
      ) : null}
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        {emailRequired ? (
          <label className="block text-sm">
            <span
              className="mb-1 block font-mono text-[11px] uppercase tracking-widest"
              style={{ color: "var(--muted-text)" }}
            >
              Work email
            </span>
            <input
              required
              type="email"
              placeholder="you@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="field-input"
              autoComplete="email"
            />
          </label>
        ) : null}
        <label className="block text-sm">
          <span
            className="mb-1 block font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Your name
          </span>
          <input
            required
            placeholder="Your full name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="field-input"
          />
        </label>
        <label className="block text-sm">
          <span
            className="mb-1 block font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Password (12+ characters)
          </span>
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
          {loading ? "Creating account…" : "Accept invite"}
        </Button>
      </form>
      <p className="mt-6 text-sm">
        <Link href="/sign-in" style={{ color: "var(--link-blue)" }}>
          Back to sign in
        </Link>
      </p>
    </AuthLayout>
  );
}
