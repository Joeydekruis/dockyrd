"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { type FormEvent, useEffect, useRef, useState } from "react";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { AuthField } from "@/components/dockyard/ui/auth-field";
import { Button } from "@/components/dockyard/ui/button";
import { getSetupStatus, login, probeAuthState, submitMfa } from "@/lib/api";

export function SignInClient() {
  const router = useRouter();
  const loginFormRef = useRef<HTMLFormElement>(null);
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
    probeAuthState().then((state) => {
      if (state.status === "authenticated") {
        router.replace("/");
        return;
      }
      if (state.status === "mfa_pending") {
        setMfaStep(true);
      }
    });
  }, [router]);

  useEffect(() => {
    if (mfaStep) {
      mfaInputRef.current?.focus();
    }
  }, [mfaStep]);

  function readLoginFields() {
    const form = loginFormRef.current;
    if (!form) {
      return { email: email.trim(), password };
    }
    const data = new FormData(form);
    return {
      email: String(data.get("email") ?? email).trim(),
      password: String(data.get("password") ?? password),
    };
  }

  async function onLoginSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    const { email: loginEmail, password: loginPassword } = readLoginFields();
    setEmail(loginEmail);
    setPassword(loginPassword);
    try {
      const res = await login(loginEmail, loginPassword);
      if (res.mfa_required) {
        setMfaStep(true);
        return;
      }
      router.replace("/");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Incorrect email or password",
      );
    } finally {
      setLoading(false);
    }
  }

  async function onMfaSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      if (useRecovery) {
        await submitMfa("", mfaRecovery.trim());
      } else {
        await submitMfa(mfaCode.trim());
      }
      router.replace("/");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Invalid authentication code",
      );
    } finally {
      setLoading(false);
    }
  }

  if (mfaStep) {
    return (
      <AuthLayout
        title="Two-factor authentication"
        subtitle="Enter the code from your authenticator app, or use a recovery code."
      >
        <form
          id="mfa-form"
          method="post"
          action="/sign-in"
          autoComplete="on"
          onSubmit={onMfaSubmit}
          className="flex flex-col gap-4"
        >
          <input
            type="text"
            name="username"
            autoComplete="username"
            value={email}
            readOnly
            tabIndex={-1}
            aria-hidden
            className="pointer-events-none absolute h-0 w-0 opacity-0"
          />
          {useRecovery ? (
            <AuthField
              label="Recovery code"
              id="recovery-code"
              name="recovery_code"
              autoComplete="off"
              placeholder="XXXX-XXXX-XXXX"
              value={mfaRecovery}
              onChange={(e) => setMfaRecovery(e.target.value)}
              className="font-mono uppercase"
            />
          ) : (
            <AuthField
              ref={mfaInputRef}
              label="Authenticator code"
              id="otp"
              name="otp"
              inputMode="numeric"
              autoComplete="one-time-code"
              placeholder="000000"
              maxLength={8}
              value={mfaCode}
              onChange={(e) => setMfaCode(e.target.value.replace(/\s/g, ""))}
              className="font-mono tracking-widest"
            />
          )}
          <Button
            type="button"
            variant="link"
            className="h-auto min-h-0 justify-start px-0 text-left"
            onClick={() => {
              setUseRecovery((v) => !v);
              setError(null);
              setMfaCode("");
              setMfaRecovery("");
            }}
          >
            {useRecovery
              ? "Use authenticator app instead"
              : "Use a recovery code"}
          </Button>
          {error ? (
            <p
              className="text-sm"
              style={{ color: "var(--danger)" }}
              role="alert"
            >
              {error}
            </p>
          ) : null}
          <Button type="submit" disabled={loading} className="w-full">
            {loading ? "Please wait…" : "Verify"}
          </Button>
        </form>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout
      title="Sign in"
      subtitle="Use your Dockyrd account email and password."
    >
      <form
        ref={loginFormRef}
        id="sign-in-form"
        method="post"
        action="/sign-in"
        autoComplete="on"
        onSubmit={onLoginSubmit}
        className="flex flex-col gap-4"
      >
        <AuthField
          label="Email"
          id="email"
          name="email"
          type="email"
          required
          autoComplete="username"
          placeholder="you@company.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          onInput={(e) => setEmail(e.currentTarget.value)}
        />
        <div>
          <span
            className="mb-1 block font-mono text-[11px] uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Password
          </span>
          <div className="relative">
            <input
              id="password"
              name="password"
              required
              type={showPassword ? "text" : "password"}
              autoComplete="current-password"
              placeholder="Your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              onInput={(e) => setPassword(e.currentTarget.value)}
              className="field-input pr-24"
            />
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              className="absolute top-1/2 right-2 min-h-11 -translate-y-1/2 rounded px-2 text-xs"
              style={{ color: "var(--link-blue)" }}
              aria-pressed={showPassword}
            >
              {showPassword ? "Hide" : "Show"}
            </button>
          </div>
        </div>
        <Link
          href="/forgot-password"
          className="text-sm"
          style={{ color: "var(--link-blue)" }}
        >
          Forgot password?
        </Link>
        {error ? (
          <p
            className="text-sm"
            style={{ color: "var(--danger)" }}
            role="alert"
          >
            {error}
          </p>
        ) : null}
        <Button type="submit" disabled={loading} className="w-full">
          {loading ? "Please wait…" : "Sign in"}
        </Button>
        <p
          className="text-center font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          or
        </p>
        <Button
          type="button"
          variant="outline"
          disabled
          title="GitHub sign-in is not configured on this server yet."
          className="w-full opacity-60"
        >
          Continue with GitHub
        </Button>
      </form>
    </AuthLayout>
  );
}
