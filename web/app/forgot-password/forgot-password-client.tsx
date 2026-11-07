"use client";

import { type FormEvent, useState } from "react";
import { AuthLayout } from "@/components/dockyard/auth-layout";
import { AuthServerResetHint } from "@/components/dockyard/auth-server-reset-hint";
import { AuthField } from "@/components/dockyard/ui/auth-field";
import { Button } from "@/components/dockyard/ui/button";
import { requestPasswordReset } from "@/lib/api";

export function ForgotPasswordClient() {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSent(false);
    const trimmed = email.trim();
    try {
      const res = await requestPasswordReset(trimmed);
      if (res.email_sent) {
        setSent(true);
        return;
      }
      setError(
        "Email is not configured on this server. Use the command below on the host.",
      );
    } catch {
      setError("Could not start reset");
    } finally {
      setLoading(false);
    }
  }

  if (sent) {
    return (
      <AuthLayout
        backHref="/sign-in"
        title="Check your inbox"
        subtitle={`If an account exists for ${email.trim()}, we sent a link to choose a new password. It expires in one hour.`}
      >
        <Button
          type="button"
          variant="outline"
          className="w-full"
          onClick={() => setSent(false)}
        >
          Send again
        </Button>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout
      backHref="/sign-in"
      title="Forgot password?"
      subtitle="Enter your email and we'll send you a link to choose a new one."
    >
      <form onSubmit={onSubmit} className="flex flex-col gap-4">
        <AuthField
          label="Email"
          type="email"
          name="email"
          required
          autoComplete="email"
          placeholder="you@company.nl"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
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
          {loading ? "Please wait…" : "Send reset link"}
        </Button>
      </form>
      <div className="mt-6">
        <AuthServerResetHint email={email} />
      </div>
    </AuthLayout>
  );
}
