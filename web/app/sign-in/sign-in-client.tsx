"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { type FormEvent, useEffect, useRef, useState } from "react";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { getMe, getSetupStatus, login, submitMfa } from "@/lib/api";

export function SignInClient() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [mfaCode, setMfaCode] = useState("");
  const [mfaRecovery, setMfaRecovery] = useState("");
  const [useRecovery, setUseRecovery] = useState(false);
  const [mfaStep, setMfaStep] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const mfaInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    getSetupStatus()
      .then((s) => {
        if (!s.complete) router.replace("/setup");
      })
      .catch(() => {});
    getMe()
      .then(() => router.replace("/"))
      .catch(() => {});
  }, [router]);

  useEffect(() => {
    if (mfaStep) {
      mfaInputRef.current?.focus();
    }
  }, [mfaStep]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      if (mfaStep) {
        if (useRecovery) {
          await submitMfa("", mfaRecovery.trim());
        } else {
          await submitMfa(mfaCode.trim());
        }
        router.replace("/");
        return;
      }
      const res = await login(email, password);
      if (res.mfa_required) {
        setMfaStep(true);
        return;
      }
      router.replace("/");
    } catch {
      setError(
        mfaStep ? "Invalid authentication code" : "Incorrect email or password",
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout title={mfaStep ? "Two-factor authentication" : "Sign in"}>
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        {mfaStep ? (
          <>
            {useRecovery ? (
              <Field label="Recovery code">
                <input
                  required
                  autoComplete="off"
                  value={mfaRecovery}
                  onChange={(e) => setMfaRecovery(e.target.value)}
                  className="field-input font-mono"
                />
              </Field>
            ) : (
              <Field label="Authenticator code">
                <input
                  ref={mfaInputRef}
                  required
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  value={mfaCode}
                  onChange={(e) => setMfaCode(e.target.value)}
                  className="field-input"
                />
              </Field>
            )}
            <button
              type="button"
              className="text-left text-sm"
              style={{ color: "var(--link-blue)" }}
              onClick={() => {
                setUseRecovery((v) => !v);
                setError(null);
              }}
            >
              {useRecovery
                ? "Use authenticator app instead"
                : "Use a recovery code"}
            </button>
          </>
        ) : (
          <>
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
            <Field label="Password">
              <div className="relative">
                <input
                  required
                  type={showPassword ? "text" : "password"}
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="field-input pr-24"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="absolute top-1/2 right-2 -translate-y-1/2 rounded px-2 py-1 text-xs"
                  style={{ color: "var(--link-blue)" }}
                >
                  {showPassword ? "Hide" : "Show"}
                </button>
              </div>
            </Field>
            <Link
              href="/forgot-password"
              className="text-sm"
              style={{ color: "var(--link-blue)" }}
            >
              Forgot password?
            </Link>
          </>
        )}
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
          disabled={loading}
          className="min-h-11 rounded-[10px] font-mono text-xs uppercase tracking-wide text-white"
          style={{ background: "var(--ink-navy)" }}
        >
          {loading ? "Please wait…" : mfaStep ? "Verify" : "Sign in"}
        </button>
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
