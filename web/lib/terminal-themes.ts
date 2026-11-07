export type TerminalThemeId =
  | "dockyrd-dark"
  | "dockyrd-light"
  | "hacker-green"
  | "normal-light";

export type TerminalTheme = {
  id: TerminalThemeId;
  name: string;
  mode: "Dark" | "Light";
  surface: string;
  bar: string;
  border: string;
  preview: {
    bg: string;
    border: string;
    prompt: string;
    cmd: string;
    text: string;
    ok: string;
    warn: string;
  };
  xterm: {
    background: string;
    foreground: string;
    cursor: string;
    selectionBackground?: string;
  };
  /** Toolbar, tabs, and chrome controls (not dashboard --text). */
  chrome: {
    fg: string;
    muted: string;
    border: string;
  };
};

export const TERMINAL_THEMES: TerminalTheme[] = [
  {
    id: "dockyrd-dark",
    name: "Dockyrd dark",
    mode: "Dark",
    surface: "#0A1B33",
    bar: "#071326",
    border: "#16315A",
    preview: {
      bg: "#0A1B33",
      border: "#16315A",
      prompt: "#CFE3FF",
      cmd: "#E3F0FF",
      text: "#8EBEFF",
      ok: "#6EE7A8",
      warn: "#F5B85C",
    },
    xterm: {
      background: "#0A1B33",
      foreground: "#E3F0FF",
      cursor: "#8EBEFF",
      selectionBackground: "#25497E",
    },
    chrome: { fg: "#E3F0FF", muted: "#8AA9D3", border: "#25497E" },
  },
  {
    id: "dockyrd-light",
    name: "Dockyrd light",
    mode: "Light",
    surface: "#F4F9FF",
    bar: "#E3F0FF",
    border: "#CFE3FF",
    preview: {
      bg: "#F4F9FF",
      border: "#CFE3FF",
      prompt: "#0B3D91",
      cmd: "#0B1F3A",
      text: "#0B57C9",
      ok: "#1F6B3A",
      warn: "#8A5200",
    },
    xterm: {
      background: "#F4F9FF",
      foreground: "#0B1F3A",
      cursor: "#0B57C9",
      selectionBackground: "#B7D3F7",
    },
    chrome: { fg: "#0B1F3A", muted: "#5A6577", border: "#B7D3F7" },
  },
  {
    id: "hacker-green",
    name: "Hacker green",
    mode: "Dark",
    surface: "#050A06",
    bar: "#020503",
    border: "#0F2A14",
    preview: {
      bg: "#050A06",
      border: "#0F2A14",
      prompt: "#8DFFAA",
      cmd: "#C4FFD4",
      text: "#3DEB6F",
      ok: "#A6FF5E",
      warn: "#FFD84A",
    },
    xterm: {
      background: "#050A06",
      foreground: "#3DEB6F",
      cursor: "#3DEB6F",
      selectionBackground: "#1B4A24",
    },
    chrome: { fg: "#8DFFAA", muted: "#4FB46A", border: "#1B4A24" },
  },
  {
    id: "normal-light",
    name: "Normal light",
    mode: "Light",
    surface: "#FFFFFF",
    bar: "#F1F1F1",
    border: "#DDDDDD",
    preview: {
      bg: "#FFFFFF",
      border: "#DDDDDD",
      prompt: "#116329",
      cmd: "#1A1A1A",
      text: "#24292F",
      ok: "#116329",
      warn: "#8A5200",
    },
    xterm: {
      background: "#FFFFFF",
      foreground: "#24292F",
      cursor: "#24292F",
      selectionBackground: "#CFCFCF",
    },
    chrome: { fg: "#1A1A1A", muted: "#555555", border: "#CFCFCF" },
  },
];

const STORAGE_KEY = "dy_terminal_theme";

export function loadTerminalThemeId(): TerminalThemeId {
  if (typeof window === "undefined") return "dockyrd-dark";
  const raw = window.localStorage.getItem(STORAGE_KEY);
  const hit = TERMINAL_THEMES.find((t) => t.id === raw);
  return hit?.id ?? "dockyrd-dark";
}

export function saveTerminalThemeId(id: TerminalThemeId) {
  window.localStorage.setItem(STORAGE_KEY, id);
}

export function themeById(id: TerminalThemeId): TerminalTheme {
  return TERMINAL_THEMES.find((t) => t.id === id) ?? TERMINAL_THEMES[0];
}

/** Strip ANSI escape sequences for plain-text copy (e.g. paste into an AI chat). */
export function stripAnsi(input: string): string {
  const esc = String.fromCharCode(27);
  const bell = String.fromCharCode(7);
  return input
    .replace(new RegExp(`${esc}\\[[0-9;?]*[ -/]*[@-~]`, "g"), "")
    .replace(new RegExp(`${esc}\\][^${bell}]*${bell}`, "g"), "")
    .replace(new RegExp(`${esc}[()][AB012]`, "g"), "");
}
