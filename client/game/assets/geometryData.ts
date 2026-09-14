import * as THREE from "three";

type AttributeData = { array: Float32Array; itemSize: number; normalized: boolean };
export type GeometryData = {
  attributes: Record<string, AttributeData>;
  index: Uint16Array | Uint32Array | null;
  groups: { start: number; count: number; materialIndex?: number }[];
  bounds: { min: number[]; max: number[]; center: number[]; radius: number };
};

export function packGeometry(geometry: THREE.BufferGeometry): GeometryData {
  geometry.computeBoundingBox();
  geometry.computeBoundingSphere();
  return {
    attributes: Object.fromEntries(Object.entries(geometry.attributes).map(([name, attribute]) => [name, {
      // Copy before transfer: cached worker geometry must retain its own buffers.
      array: new Float32Array(attribute.array), itemSize: attribute.itemSize, normalized: attribute.normalized,
    }])),
    index: geometry.index ? new (geometry.index.array instanceof Uint32Array ? Uint32Array : Uint16Array)(geometry.index.array) : null,
    groups: geometry.groups.map(group => ({ ...group })),
    bounds: {
      min: geometry.boundingBox!.min.toArray(), max: geometry.boundingBox!.max.toArray(),
      center: geometry.boundingSphere!.center.toArray(), radius: geometry.boundingSphere!.radius,
    },
  };
}

export function unpackGeometry(data: GeometryData) {
  const geometry = new THREE.BufferGeometry();
  for (const [name, attribute] of Object.entries(data.attributes)) {
    geometry.setAttribute(name, new THREE.BufferAttribute(attribute.array, attribute.itemSize, attribute.normalized));
  }
  if (data.index) geometry.setIndex(new THREE.BufferAttribute(data.index, 1));
  geometry.groups = data.groups;
  geometry.boundingBox = new THREE.Box3(new THREE.Vector3().fromArray(data.bounds.min), new THREE.Vector3().fromArray(data.bounds.max));
  geometry.boundingSphere = new THREE.Sphere(new THREE.Vector3().fromArray(data.bounds.center), data.bounds.radius);
  return geometry;
}

export function geometryTransfers(data: GeometryData): ArrayBuffer[] {
  return [...Object.values(data.attributes).map(attribute => attribute.array.buffer as ArrayBuffer),
    ...(data.index ? [data.index.buffer as ArrayBuffer] : [])];
}
