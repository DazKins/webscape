import * as THREE from "three";

type Appearance = { color: number; variation: number; detail?: "grass" | "stone"; height?: number; count?: number; minCount?: number };
const appearances: Record<string, Appearance> = {
  grass: { color: 0x73964f, variation: 0.045, detail: "grass", height: 0.09, count: 4 },
  grassShort: { color: 0x73964f, variation: 0.045, detail: "grass", height: 0.09, count: 4 },
  grassLong: { color: 0x698c46, variation: 0.045, detail: "grass", height: 0.48, count: 12, minCount: 4 },
  stone: { color: 0x8b9296, variation: 0.035, detail: "stone", height: 0.055, count: 3 },
  dirt: { color: 0x9a6b42, variation: 0.035 },
  road: { color: 0xb8ab88, variation: 0.025 },
  water: { color: 0x4f8fb8, variation: 0 },
};
const fallback: Appearance = { color: 0xe77d11, variation: 0 };
export function terrainAppearance(type: string): Appearance { return Object.prototype.hasOwnProperty.call(appearances, type) ? appearances[type] : fallback; }
export function terrainColor(type: string) { return terrainAppearance(type).color; }

// Independent, stable streams: geometry rebuilds and loading order cannot reshuffle a tile.
export function tileRandom(x: number, y: number, type: string, stream = 0) {
  let seed = Math.imul(x, 73856093) ^ Math.imul(y, 19349663) ^ Math.imul(stream, 83492791);
  for (let i = 0; i < type.length; i++) seed = Math.imul(seed ^ type.charCodeAt(i), 16777619);
  return () => {
    seed = (seed + 0x6d2b79f5) | 0;
    let value = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    value ^= value + Math.imul(value ^ (value >>> 7), 61 | value);
    return ((value ^ (value >>> 14)) >>> 0) / 4294967296;
  };
}
export function variedTerrainColor(type: string, x: number, y: number) {
  const appearance = terrainAppearance(type);
  return new THREE.Color(appearance.color).multiplyScalar(1 + (tileRandom(x, y, type)() * 2 - 1) * appearance.variation);
}
