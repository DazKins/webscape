import * as THREE from "three";

type Resource = THREE.BufferGeometry | THREE.Material;
const references = new WeakMap<Resource, number>();

// Cache entries and live instances each own one reference. Eviction must never
// invalidate another instance (including equipment attached to a different rig).
export function retainResource(resource: Resource) {
  references.set(resource, (references.get(resource) ?? 0) + 1);
}

export function releaseResource(resource: Resource) {
  const remaining = (references.get(resource) ?? 1) - 1;
  if (remaining > 0) {
    references.set(resource, remaining);
    return;
  }
  references.delete(resource);
  if (resource instanceof THREE.Material) {
    for (const value of Object.values(resource)) {
      if (value instanceof THREE.Texture) value.dispose();
    }
  }
  resource.dispose();
}

export class AssetCache<T> {
  private entries = new Map<string, T>();
  constructor(private readonly capacity: number, private readonly dispose: (value: T) => void) {}

  get(key: string): T | undefined {
    const value = this.entries.get(key);
    if (value !== undefined) {
      this.entries.delete(key);
      this.entries.set(key, value);
    }
    return value;
  }

  set(key: string, value: T) {
    const existing = this.entries.get(key);
    if (existing === value) return;
    if (existing !== undefined) this.dispose(existing);
    this.entries.delete(key);
    this.entries.set(key, value);
    while (this.entries.size > this.capacity) {
      const oldest = this.entries.keys().next().value!;
      this.dispose(this.entries.get(oldest)!);
      this.entries.delete(oldest);
    }
  }

  clear() {
    for (const value of this.entries.values()) this.dispose(value);
    this.entries.clear();
  }

  values() { return [...this.entries.entries()]; }
}

export const geometryAssets = new AssetCache<THREE.BufferGeometry>(1024, releaseResource);
const materialAssets = new AssetCache<THREE.Material>(256, releaseResource);

export function sharedGeometry(key: string, build: () => THREE.BufferGeometry) {
  const existing = geometryAssets.get(key);
  if (existing) return existing;
  const geometry = build();
  geometry.userData.assetKey = key;
  retainResource(geometry);
  geometryAssets.set(key, geometry);
  return geometry;
}

export function sharedMaterial<T extends THREE.Material>(key: string, build: () => T): T {
  const existing = materialAssets.get(key);
  if (existing) return existing as T;
  const material = build();
  retainResource(material);
  materialAssets.set(key, material);
  return material;
}

export function clearSharedAssets() {
  geometryAssets.clear();
  materialAssets.clear();
}
