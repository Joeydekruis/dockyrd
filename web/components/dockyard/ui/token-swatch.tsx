type TokenSwatchProps = {
  name: string;
  hex: string;
};

export function TokenSwatch({ name, hex }: TokenSwatchProps) {
  return (
    <div className="flex items-center gap-3">
      <span
        className="size-10 shrink-0 rounded-lg border"
        style={{ background: hex, borderColor: "var(--border)" }}
        aria-hidden
      />
      <div>
        <p className="font-mono text-[11px] uppercase tracking-widest">
          {name}
        </p>
        <p
          className="font-mono text-[13px] tabular-nums"
          style={{ color: "var(--muted-text)" }}
        >
          {hex}
        </p>
      </div>
    </div>
  );
}
