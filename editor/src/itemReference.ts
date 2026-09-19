import { isObject } from "./formatUtils";

// Definition existence and gameplay compatibility are validated by the Go loader.
export function isItemDefinitionId(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

export type ItemReference = { definitionId: string; count: number };
export function normalizeItemReference(value: unknown): ItemReference {
  if (!isObject(value)) throw new Error("item reference must be an object");
  if ("name" in value || "type" in value) throw new Error("item references must use definitionId instead of name/type");
  const allowed = ["definitionId", "count"];
  if (Object.keys(value).some(key => !allowed.includes(key))) throw new Error("unknown item reference field");
  const definitionId = value.definitionId;
  if (!isItemDefinitionId(definitionId)) throw new Error("definitionId must be a non-empty string");
  if (!Number.isInteger(value.count) || Number(value.count) < 1 || Number(value.count) > 2147483647) {
    throw new Error("item count must be an integer from 1 to 2147483647");
  }
  return { definitionId, count: Number(value.count) };
}

export function validateItemReference(value: unknown, label: string, errors: string[]) {
  try {
    normalizeItemReference(value);
  } catch (error) {
    errors.push(`${label}: ${error instanceof Error ? error.message : String(error)}`);
  }
}
