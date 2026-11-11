export const CONNECTION_VALUE_MASK = "••••••••";

export function isSecretConnectionVariableKey(key: string): boolean {
  const upper = key.toUpperCase();
  return (
    upper.includes("PASSWORD") || key === "DATABASE_URL" || key === "REDIS_URL"
  );
}
