import { Suspense } from "react";
import { SetupClient } from "./setup-client";

export default function SetupPage() {
  return (
    <Suspense
      fallback={
        <div
          className="flex min-h-dvh items-center justify-center text-sm"
          style={{ color: "var(--muted-text)" }}
        >
          Loading…
        </div>
      }
    >
      <SetupClient />
    </Suspense>
  );
}
