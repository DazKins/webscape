export type ValidationResult = {
  valid: boolean;
  errors: string[];
};

export const ID_PATTERN = /^[a-z0-9][a-z0-9_-]*$/;

export function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function serializeJson(value: unknown): string {
  return `${JSON.stringify(value, null, 2)}\n`;
}

export function titleFromId(id: string): string {
  return id
    .split(/[_-]+/)
    .filter(Boolean)
    .map((part) => `${part.slice(0, 1).toUpperCase()}${part.slice(1)}`)
    .join(" ");
}
