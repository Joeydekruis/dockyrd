import { X } from "lucide-react";

type ModalProps = {
  open: boolean;
  title: string;
  description?: string;
  children?: React.ReactNode;
  footer?: React.ReactNode;
  onClose?: () => void;
  size?: "md" | "lg";
};

export function Modal({
  open,
  title,
  description,
  children,
  footer,
  onClose,
  size = "md",
}: ModalProps) {
  if (!open) return null;

  const maxW = size === "lg" ? "max-w-2xl" : "max-w-lg";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        className="absolute inset-0 border-0"
        style={{ background: "rgba(11, 31, 58, 0.45)" }}
        aria-label="Close dialog"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="dockyrd-modal-title"
        className={`relative z-10 max-h-[85vh] w-full ${maxW} overflow-y-auto rounded-[20px] border p-6 shadow-lg`}
        style={{ background: "var(--surface)", borderColor: "var(--border)" }}
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 id="dockyrd-modal-title" className="text-lg font-semibold">
              {title}
            </h2>
            {description ? (
              <p
                className="mt-2 text-sm"
                style={{ color: "var(--muted-text)" }}
              >
                {description}
              </p>
            ) : null}
          </div>
          {onClose ? (
            <button
              type="button"
              className="btn-ghost rounded-lg p-2"
              aria-label="Close"
              onClick={onClose}
            >
              <X className="size-5" strokeWidth={1.8} />
            </button>
          ) : null}
        </div>
        {children ? <div className="mt-4">{children}</div> : null}
        {footer ? (
          <div className="mt-6 flex flex-wrap justify-end gap-2">{footer}</div>
        ) : null}
      </div>
    </div>
  );
}
