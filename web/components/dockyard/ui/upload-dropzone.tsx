import { CheckCircle2, CloudUpload, Loader2, XCircle } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "@/components/dockyard/ui/button";

export type UploadDropzoneState = "idle" | "uploading" | "done" | "error";

type UploadDropzoneProps = {
  state?: UploadDropzoneState;
  title?: string;
  hint?: string;
  errorMessage?: string;
  onBrowse?: () => void;
  children?: ReactNode;
};

export function UploadDropzone({
  state = "idle",
  title = "Drop files here or browse",
  hint = "PNG, JPG up to 10 MB",
  errorMessage,
  onBrowse,
  children,
}: UploadDropzoneProps) {
  const Icon =
    state === "uploading"
      ? Loader2
      : state === "done"
        ? CheckCircle2
        : state === "error"
          ? XCircle
          : CloudUpload;
  const iconColor =
    state === "done"
      ? "#1a7f4b"
      : state === "error"
        ? "var(--danger)"
        : "var(--link-blue)";
  return (
    <div
      className="flex flex-col items-center justify-center rounded-[20px] border border-dashed px-6 py-10 text-center"
      style={{
        borderColor:
          state === "error" ? "var(--danger)" : "var(--input-border)",
        background: "var(--surface)",
      }}
    >
      <Icon
        strokeWidth={1.8}
        className={`mb-3 size-8 ${state === "uploading" ? "animate-spin" : ""}`}
        style={{ color: iconColor }}
        aria-hidden
      />
      <p className="text-sm font-medium">{title}</p>
      {state === "error" && errorMessage ? (
        <p className="mt-1 text-[13px]" style={{ color: "var(--danger)" }}>
          {errorMessage}
        </p>
      ) : (
        <p className="mt-1 text-[13px]" style={{ color: "var(--muted-text)" }}>
          {hint}
        </p>
      )}
      {children ??
        (state === "idle" && onBrowse ? (
          <Button variant="outline" className="mt-4" onClick={onBrowse}>
            Browse files
          </Button>
        ) : null)}
    </div>
  );
}
