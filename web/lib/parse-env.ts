export type ParsedEnvLine = {
  key: string;
  value: string;
  isSecret: boolean;
};

/** Parse a pasted .env blob into key/value pairs (one entry per non-comment line). */
export function parseEnvPaste(text: string): ParsedEnvLine[] {
  const out: ParsedEnvLine[] = [];
  const seen = new Set<string>();
  for (const rawLine of text.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) {
      continue;
    }
    let body = line;
    if (body.startsWith("export ")) {
      body = body.slice(7).trim();
    }
    const eq = body.indexOf("=");
    if (eq <= 0) {
      continue;
    }
    const key = body.slice(0, eq).trim();
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(key)) {
      continue;
    }
    let value = body.slice(eq + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    const upper = key.toUpperCase();
    if (seen.has(upper)) {
      continue;
    }
    seen.add(upper);
    const isSecret =
      /secret|token|password|api_key|apikey|private/i.test(key) &&
      !/public/i.test(key);
    out.push({ key: upper, value, isSecret });
  }
  return out;
}
