"use client";

import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  SnackbarViewport,
  type SnackbarTone,
} from "@/components/dockyard/ui/snackbar";
import { writeClipboardText } from "@/lib/clipboard";

type SnackbarState = {
  message: string;
  tone: SnackbarTone;
} | null;

type SnackbarContextValue = {
  showSnackbar: (message: string, tone?: SnackbarTone) => void;
  dismissSnackbar: () => void;
  copyToClipboard: (value: string, message?: string) => Promise<void>;
};

const SnackbarContext = createContext<SnackbarContextValue | null>(null);

const noopContext: SnackbarContextValue = {
  showSnackbar: () => {},
  dismissSnackbar: () => {},
  copyToClipboard: async (value) => {
    await writeClipboardText(value);
  },
};

export function SnackbarProvider({ children }: { children: ReactNode }) {
  const [snackbar, setSnackbar] = useState<SnackbarState>(null);

  const dismissSnackbar = useCallback(() => setSnackbar(null), []);

  useEffect(() => {
    if (!snackbar) return;
    const id = window.setTimeout(dismissSnackbar, 4000);
    return () => window.clearTimeout(id);
  }, [snackbar, dismissSnackbar]);

  const showSnackbar = useCallback(
    (message: string, tone: SnackbarTone = "info") => {
      setSnackbar({ message, tone });
    },
    [],
  );

  const copyToClipboard = useCallback(
    async (value: string, message = "Copied to clipboard") => {
      try {
        await writeClipboardText(value);
        showSnackbar(message, "success");
      } catch {
        showSnackbar("Could not copy to clipboard", "error");
      }
    },
    [showSnackbar],
  );

  const value = useMemo(
    () => ({ showSnackbar, dismissSnackbar, copyToClipboard }),
    [showSnackbar, dismissSnackbar, copyToClipboard],
  );

  return (
    <SnackbarContext.Provider value={value}>
      {children}
      <SnackbarViewport
        open={snackbar !== null}
        message={snackbar?.message ?? ""}
        tone={snackbar?.tone ?? "info"}
        onDismiss={dismissSnackbar}
        durationMs={4000}
      />
    </SnackbarContext.Provider>
  );
}

export function useSnackbar(): SnackbarContextValue {
  return useContext(SnackbarContext) ?? noopContext;
}

export function useCopyToClipboard() {
  const { copyToClipboard } = useSnackbar();
  return copyToClipboard;
}
