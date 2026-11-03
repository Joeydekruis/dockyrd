import Link from "next/link";
import { AuthLayout } from "@/components/dockyard/auth-layout";

export default function ForgotPasswordPage() {
  return (
    <AuthLayout title="Forgot password?">
      <p style={{ color: "var(--muted-text)" }}>
        Password reset via email lands in a later sprint. Ask an admin or run{" "}
        <code className="font-mono text-xs">sudo dockyard admin reset-password</code> on the server.
      </p>
      <Link href="/sign-in" className="mt-6 inline-block text-sm" style={{ color: "var(--link-blue)" }}>
        Back to sign in
      </Link>
    </AuthLayout>
  );
}
