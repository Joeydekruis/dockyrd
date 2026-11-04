type DangerZoneProps = {
  title: string;
  description: string;
  action: React.ReactNode;
};

export function DangerZone({ title, description, action }: DangerZoneProps) {
  return (
    <section
      className="rounded-[20px] border p-6"
      style={{ borderColor: "#f5c2c0", background: "#fff8f7" }}
    >
      <h2 className="text-lg font-semibold" style={{ color: "var(--danger)" }}>
        {title}
      </h2>
      <p className="mt-2 text-sm" style={{ color: "var(--secondary-text)" }}>
        {description}
      </p>
      <div className="mt-4">{action}</div>
    </section>
  );
}
