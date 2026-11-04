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
import { getLocalePrefs, patchLocalePrefs } from "@/lib/api";
import {
  defaultUserLocalePrefs,
  formatClockTime,
  formatDateTime,
  formatTimeAxisLabel,
  metricTimeLabels,
  type UserLocalePrefs,
} from "@/lib/user-locale";

type UserLocaleContextValue = {
  prefs: UserLocalePrefs;
  loading: boolean;
  refresh: () => Promise<void>;
  save: (next: UserLocalePrefs) => Promise<UserLocalePrefs>;
  formatClockTime: (tsUnixSec: number) => string;
  formatDateTime: (tsUnixSec: number) => string;
  formatTimeAxisLabel: (
    tsUnixSec: number,
    index: number,
    total: number,
  ) => string;
  metricTimeLabels: (points: { ts: number }[]) => string[];
};

const UserLocaleContext = createContext<UserLocaleContextValue | null>(null);

export function UserLocaleProvider({ children }: { children: ReactNode }) {
  const [prefs, setPrefs] = useState<UserLocalePrefs>(defaultUserLocalePrefs);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getLocalePrefs();
      setPrefs(data);
    } catch {
      setPrefs(defaultUserLocalePrefs);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const save = useCallback(async (next: UserLocalePrefs) => {
    const updated = await patchLocalePrefs(next);
    setPrefs(updated);
    return updated;
  }, []);

  const value = useMemo<UserLocaleContextValue>(() => {
    return {
      prefs,
      loading,
      refresh,
      save,
      formatClockTime: (ts) => formatClockTime(ts, prefs),
      formatDateTime: (ts) => formatDateTime(ts, prefs),
      formatTimeAxisLabel: (ts, index, total) =>
        formatTimeAxisLabel(ts, index, total, prefs),
      metricTimeLabels: (points) => metricTimeLabels(points, prefs),
    };
  }, [prefs, loading, refresh, save]);

  return (
    <UserLocaleContext.Provider value={value}>
      {children}
    </UserLocaleContext.Provider>
  );
}

export function useUserLocale(): UserLocaleContextValue {
  const ctx = useContext(UserLocaleContext);
  if (!ctx) {
    return {
      prefs: defaultUserLocalePrefs,
      loading: false,
      refresh: async () => {},
      save: async (next) => next,
      formatClockTime: (ts) => formatClockTime(ts, defaultUserLocalePrefs),
      formatDateTime: (ts) => formatDateTime(ts, defaultUserLocalePrefs),
      formatTimeAxisLabel: (ts, index, total) =>
        formatTimeAxisLabel(ts, index, total, defaultUserLocalePrefs),
      metricTimeLabels: (points) =>
        metricTimeLabels(points, defaultUserLocalePrefs),
    };
  }
  return ctx;
}
