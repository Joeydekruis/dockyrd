type ModalProps = {
  open: boolean;
  title: string;
  description?: string;
  children?: React.ReactNode;
  footer?: React.ReactNode;
  onClose?: () => void;
};

export function Modal({
  open,
  title,
  description,
  children,
  footer,
  onClose,
}: ModalProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        className="absolute inset-0 border-0 bg-black/40"
        aria-label="Close dialog"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="dockyard-modal-title"
        className="relative z-10 max-h-[85vh] w-full max-w-lg overflow-y-auto rounded-[20px] border p-6 shadow-lg"
        style={{ background: "var(--surface)", borderColor: "var(--border)" }}
      >
        <h2 id="dockyard-modal-title" className="text-lg font-semibold">
          {title}
        </h2>
        {description ? (
          <p className="mt-2 text-sm" style={{ color: "var(--muted-text)" }}>
            {description}
          </p>
        ) : null}
        {children ? <div className="mt-4">{children}</div> : null}
        {footer ? (
          <div className="mt-6 flex flex-wrap gap-2">{footer}</div>
        ) : null}
      </div>
    </div>
  );
}
