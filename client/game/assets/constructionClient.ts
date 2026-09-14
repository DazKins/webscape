import { AssetCache } from "../models/assetCache";
import type { ModelName } from "../models/registry";
import type { ModelOptions } from "../models/types";
import { construct, installModelGeometry, type ConstructionRequest, type ConstructionResult } from "./construction";
import { unpackGeometry } from "./geometryData";

class ConstructionClient {
  private worker: Worker | undefined;
  private failed = false;
  private nextId = 0;
  private pending = new Map<number, { request: ConstructionRequest; resolve: (result: ConstructionResult) => void; reject: (error: unknown) => void }>();

  run(request: ConstructionRequest): Promise<ConstructionResult> {
    if (!this.worker && !this.failed) {
      try {
        this.worker = new Worker(new URL("./construction.worker.ts", import.meta.url), { type: "module" });
        this.worker.onmessage = (event: MessageEvent<{ id: number; result: ConstructionResult; error?: string }>) => {
          const pending = this.pending.get(event.data.id);
          if (!pending) return;
          this.pending.delete(event.data.id);
          if (event.data.error) pending.reject(new Error(event.data.error));
          else pending.resolve(event.data.result);
        };
        this.worker.onerror = () => this.fallback();
        this.worker.onmessageerror = () => this.fallback();
      } catch { this.failed = true; }
    }
    if (this.failed) return this.runLocally(request);
    return new Promise((resolve, reject) => {
      const id = this.nextId++;
      this.pending.set(id, { request, resolve, reject });
      try { this.worker!.postMessage({ id, request }); } catch { this.fallback(); }
    });
  }

  private fallback() {
    this.worker?.terminate();
    this.worker = undefined;
    this.failed = true;
    for (const { request, resolve, reject } of this.pending.values()) {
      this.runLocally(request).then(resolve, reject);
    }
    this.pending.clear();
  }

  private runLocally(request: ConstructionRequest): Promise<ConstructionResult> {
    return new Promise((resolve, reject) => setTimeout(() => {
      try { resolve(construct(request)); } catch (error) { reject(error); }
    }, 0));
  }
}

export const constructionClient = new ConstructionClient();
let warming: Promise<void> | undefined;
export function prepareModelAssets() {
  return warming ??= constructionClient.run({ kind: "warmModels" }).then(result => {
    if (result.kind === "warmModels") installModelGeometry(result.geometries, unpackGeometry);
  });
}

const preparedVariants = new AssetCache<boolean>(64, () => {});
const preparingVariants = new Set<string>();

// Used for dimensions supplied by content, which cannot all be prepared at startup.
export function modelVariantReady(name: ModelName, options: ModelOptions): boolean {
  const key = JSON.stringify([name, options]);
  if (preparedVariants.get(key)) return true;
  if (!preparingVariants.has(key)) {
    preparingVariants.add(key);
    constructionClient.run({ kind: "model", name, options }).then(result => {
      if (result.kind === "warmModels") installModelGeometry(result.geometries, unpackGeometry);
    }).catch(error => console.error("Model preparation failed; using local construction", error)).finally(() => {
      preparingVariants.delete(key);
      preparedVariants.set(key, true);
    });
  }
  return false;
}
