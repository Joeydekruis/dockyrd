"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { completeSetup, getSetupStatus } from "@/lib/api";

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
      router.replace("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Setup failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout title="Create owner account">
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <Field label="Setup token">
          <input required value={token} onChange={(e) => setToken(e.target.value)} className="field-input" />
        </Field>
        <Field label="Name">
          <input required value={name} onChange={(e) => setName(e.target.value)} className="field-input" />
        </Field>
        <Field label="Email">
          <input
            required
            type="email"
            autoComplete="email"
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
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="field-input"
          />
        </Field>
        {error ? (
          <p className="text-sm" style={{ color: "var(--danger)" }} role="alert">
            {error}
          </p>
        ) : null}
        <button
          type="submit"
          disabled={loading}
          className="mt-2 min-h-11 rounded-[10px] font-mono text-xs uppercase tracking-wide text-white"
          style={{ background: "var(--ink-navy)" }}
        >
          {loading ? "Saving…" : "Finish setup"}
        </button>
      </form>
    </AuthLayout>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block text-sm">
      <span
        className="mb-1 block font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {label}
      </span>
      {children}
    </label>
  );
}
