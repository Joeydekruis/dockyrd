const settingsCardClass =
  "mb-8 space-y-4 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6";

export function SettingsSection({
  title,
  description,
  children,
  className,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={className ?? settingsCardClass} aria-labelledby={title}>
      <div>
        <h2 className="text-lg font-medium" id={title}>
          {title}
        </h2>
        {description ? (
          <p className="mt-1 text-sm text-[var(--muted-text)]">{description}</p>
        ) : null}
      </div>
      {children}
    </section>
  );
}

export function SettingsErrorBanner({ message }: { message: string | null }) {
  if (!message) return null;
  return (
    <p className="mb-4 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-700">
      {message}
    </p>
  );
}

export function Owner2FARequiredBanner({
  show,
}: {
  show: boolean;
}) {
  if (!show) return null;
  return (
    <p
      className="mb-4 rounded-lg border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm"
      role="status"
    >
      Owners must enable two-factor authentication before using the rest of the
      dashboard. Set it up under Password &amp; 2FA.
    </p>
  );
}
