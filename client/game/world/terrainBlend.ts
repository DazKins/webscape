import * as THREE from "three";
import { terrainAppearance, variedTerrainColor } from "./terrainAppearance";

export type TerrainBlend = ReturnType<typeof createTerrainBlend>;

// A snapshot includes diagonal neighbours. Undefined means unloaded, not a terrain type.
export function createTerrainBlend(sizeX: number, sizeY: number, terrain: string[], border: (string | undefined)[], originX: number, originY: number) {
  const tiles = new Map<string, { type: string; color: THREE.Color }>();
  for (let y = -1; y <= sizeY; y++) for (let x = -1; x <= sizeX; x++) {
    const type = x >= 0 && y >= 0 && x < sizeX && y < sizeY
      ? terrain[y * sizeX + x] : border[(y + 1) * (sizeX + 2) + x + 1];
    if (type !== undefined) tiles.set(`${x},${y}`, { type, color: variedTerrainColor(type, originX + x, originY + y) });
  }
  return (x: number, y: number) => {
    const worldX = originX + x, worldY = originY + y;
    // Continuous world-space variation gives both sides of a chunk seam identical edges.
    const sx = x + 0.045 * Math.sin(worldY * 5.3 + Math.sin(worldX * 2.1));
    const sy = y + 0.045 * Math.sin(worldX * 4.7 + Math.sin(worldY * 2.7));
    const left = Math.floor(sx - 0.5), top = Math.floor(sy - 0.5);
    const color = new THREE.Color(0, 0, 0);
    let total = 0, water = 0, grass = 0;
    for (let ty = top; ty <= top + 1; ty++) for (let tx = left; tx <= left + 1; tx++) {
      const tile = tiles.get(`${tx},${ty}`);
      if (!tile) continue;
      const width = tile.type === "road" ? 0.18 : tile.type === "water" ? 0.28 : 0.32;
      const weight = influence(Math.abs(sx - tx - 0.5), width) * influence(Math.abs(sy - ty - 0.5), width);
      color.r += tile.color.r * weight;
      color.g += tile.color.g * weight;
      color.b += tile.color.b * weight;
      total += weight;
      if (tile.type === "water") water += weight;
      if (terrainAppearance(tile.type).detail === "grass") grass += weight;
    }
    if (total > 0) color.multiplyScalar(1 / total);
    return { color, water: total > 0 ? water / total : 0, grass: total > 0 ? grass / total : 0 };
  };
}

function influence(distance: number, width: number) {
  const t = THREE.MathUtils.clamp((0.5 + width - distance) / (2 * width), 0, 1);
  return t * t * (3 - 2 * t);
}
