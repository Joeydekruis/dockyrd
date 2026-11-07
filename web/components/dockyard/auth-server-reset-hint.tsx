type AuthServerResetHintProps = {
  email: string;
};

/** Forgot-password fallback when SMTP is not configured (design card). */
export function AuthServerResetHint({ email }: AuthServerResetHintProps) {
  const address = email.trim() || "you@company.nl";
  const command = `dockyrd admin reset-password ${address}`;

  return (
    <div className="auth-cli-card">
      <span
        className="font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        No email server set up?
      </span>
      <p className="m-0 text-sm" style={{ color: "var(--muted-text)" }}>
        Run this on your server instead:
      </p>
      <code className="auth-cli-code">{command}</code>
    </div>
  );
}
