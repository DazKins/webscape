import * as THREE from "three";
import { sampleTerrainHeight, type TerrainHeightGrid } from "./terrainHeight";
import { terrainAppearance, tileRandom } from "./terrainAppearance";

// One opaque, vertex-coloured mesh per chunk; no per-detail objects or frame updates.
export function createTerrainDetails(grid: TerrainHeightGrid, terrain: string[], originX: number, originY: number) {
  const positions: number[] = [], colors: number[] = [];
  const color = new THREE.Color();
  const vertex = (x: number, y: number, z: number) => {
    positions.push(x, y, z); colors.push(color.r, color.g, color.b);
  };
  // Match the rendered half-tile triangles, including on non-planar slopes.
  const height = (x: number, z: number) => {
    const left = Math.floor(x * 2) / 2, top = Math.floor(z * 2) / 2;
    const u = (x - left) * 2, v = (z - top) * 2;
    const a = sampleTerrainHeight(grid, left, top), b = sampleTerrainHeight(grid, left + 0.5, top);
    const c = sampleTerrainHeight(grid, left, top + 0.5), d = sampleTerrainHeight(grid, left + 0.5, top + 0.5);
    return u + v <= 1 ? a + (b - a) * u + (c - a) * v : d + (c - d) * (1 - u) + (b - d) * (1 - v);
  };
  for (let y = 0; y < grid.sizeY; y++) for (let x = 0; x < grid.sizeX; x++) {
    const type = terrain[y * grid.sizeX + x] ?? "";
    const appearance = terrainAppearance(type);
    if (!appearance.detail) continue;
    const random = tileRandom(originX + x, originY + y, type, 1);
    const minimum = appearance.minCount ?? 0;
    const count = minimum + Math.floor(random() * ((appearance.count ?? 0) - minimum + 1));
    for (let i = 0; i < count; i++) {
      const cx = x + 0.12 + random() * 0.76, cz = y + 0.12 + random() * 0.76;
      const size = (appearance.height ?? 0) * (0.65 + random() * 0.7);
      const rotation = random() * Math.PI * 2;
      color.set(appearance.color).multiplyScalar(0.75 + random() * 0.35);
      if (appearance.detail === "grass") {
        for (let blade = 0; blade < 3; blade++) {
          const angle = rotation + blade * Math.PI / 3;
          const dx = Math.cos(angle) * 0.025, dz = Math.sin(angle) * 0.025;
          vertex(cx - dx, height(cx - dx, cz - dz), cz - dz);
          vertex(cx + dx, height(cx + dx, cz + dz), cz + dz);
          vertex(cx + dx * 1.4, height(cx, cz) + size * (0.7 + random() * 0.3), cz + dz * 1.4);
        }
      } else {
        for (let face = 0; face < 5; face++) {
          const a = rotation + face * Math.PI * 2 / 5, b = a + Math.PI * 2 / 5;
          const ax = cx + Math.cos(a) * size, az = cz + Math.sin(a) * size * 0.8;
          const bx = cx + Math.cos(b) * size, bz = cz + Math.sin(b) * size * 0.8;
          vertex(ax, height(ax, az), az);
          vertex(cx, height(cx, cz) + size * 0.7, cz);
          vertex(bx, height(bx, bz), bz);
        }
      }
    }
  }
  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));
  geometry.setAttribute("color", new THREE.Float32BufferAttribute(colors, 3));
  geometry.computeVertexNormals();
  return geometry;
}
