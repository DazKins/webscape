import * as THREE from "three";
import type { TerrainBlend } from "./terrainBlend";

export const TERRAIN_SEGMENTS = 4;
const tileCoordinates = (tile: number) => Array.from({ length: TERRAIN_SEGMENTS + 1 }, (_, i) => tile + i / TERRAIN_SEGMENTS);

// Keep the editor's MapViewport3D HEIGHT_SCALE in sync.
export const TERRAIN_HEIGHT_SCALE = 0.6;

const HIGHLIGHT_SURFACE_OFFSET = 0.025;
const WATER_SURFACE_OFFSET = 0.012;

export type TerrainHeightGrid = {
  sizeX: number;
  sizeY: number;
  heights: number[];
  sampleOutside?: (x: number, y: number) => number;
};

export function getTileHeight(grid: TerrainHeightGrid, x: number, y: number) {
  if (
    !Number.isInteger(x) ||
    !Number.isInteger(y) ||
    x < 0 ||
    y < 0 ||
    x >= grid.sizeX ||
    y >= grid.sizeY
  ) {
    return grid.sampleOutside?.(x, y) ?? 0;
  }

  const height = grid.heights[y * grid.sizeX + x];
  if (!Number.isInteger(height) || height < 0 || height > 10) {
    return 0;
  }

  return height * TERRAIN_HEIGHT_SCALE;
}

export function sampleTerrainHeight(
  grid: TerrainHeightGrid,
  worldX: number,
  worldZ: number
) {
  if (grid.sizeX <= 0 || grid.sizeY <= 0) {
    return 0;
  }

  const centerX = worldX - 0.5;
  const centerZ = worldZ - 0.5;
  const x0 = Math.floor(centerX);
  const z0 = Math.floor(centerZ);
  const x1 = x0 + 1;
  const z1 = z0 + 1;
  const tx = centerX - x0;
  const tz = centerZ - z0;

  const h00 = getTileHeight(grid, x0, z0);
  const h10 = getTileHeight(grid, x1, z0);
  const h01 = getTileHeight(grid, x0, z1);
  const h11 = getTileHeight(grid, x1, z1);

  return lerp(lerp(h00, h10, tx), lerp(h01, h11, tx), tz);
}

export function createTerrainSurfaceGeometry(
  grid: TerrainHeightGrid,
  terrain: string[],
  terrainColorForTile: (terrainType: string, x: number, y: number) => THREE.ColorRepresentation,
  blend?: TerrainBlend
) {
  const positions: number[] = [];
  const colors: number[] = [];
  const indices: number[] = [];
  const color = new THREE.Color();

  for (let tileY = 0; tileY < grid.sizeY; tileY += 1) {
    for (let tileX = 0; tileX < grid.sizeX; tileX += 1) {
      const xCoords = tileCoordinates(tileX);
      const zCoords = tileCoordinates(tileY);
      const vertexOffset = positions.length / 3;
      const terrainType = terrain[tileY * grid.sizeX + tileX] ?? "";

      color.set(terrainColorForTile(terrainType, tileX, tileY));

      for (const z of zCoords) {
        for (const x of xCoords) {
          positions.push(x, sampleTerrainHeight(grid, x, z), z);
          const blended = blend?.(x, z).color ?? color;
          colors.push(blended.r, blended.g, blended.b);
        }
      }

      const width = xCoords.length;
      for (let z = 0; z < zCoords.length - 1; z += 1) {
        for (let x = 0; x < xCoords.length - 1; x += 1) {
          const topLeft = vertexOffset + z * width + x;
          const topRight = topLeft + 1;
          const bottomLeft = topLeft + width;
          const bottomRight = bottomLeft + 1;
          indices.push(topLeft, bottomLeft, topRight, topRight, bottomLeft, bottomRight);
        }
      }
    }
  }

  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));
  geometry.setAttribute("color", new THREE.Float32BufferAttribute(colors, 3));
  geometry.setIndex(indices);
  setTerrainNormals(geometry, grid);
  return geometry;
}

export function createWaterSurfaceGeometry(
  grid: TerrainHeightGrid,
  terrain: string[],
  blend?: TerrainBlend,
  originX = 0,
  originY = 0
) {
  const positions: number[] = [];
  const uvs: number[] = [];
  const waterWeights: number[] = [];
  const indices: number[] = [];

  for (let tileY = 0; tileY < grid.sizeY; tileY += 1) {
    for (let tileX = 0; tileX < grid.sizeX; tileX += 1) {
      const terrainType = terrain[tileY * grid.sizeX + tileX] ?? "";
      if (terrainType !== "water" && !blend) {
        continue;
      }

      const xCoords = tileCoordinates(tileX);
      const zCoords = tileCoordinates(tileY);
      const vertexOffset = positions.length / 3;

      const weights = zCoords.flatMap(z => xCoords.map(x => blend?.(x, z).water ?? 1));
      if (!weights.some(weight => weight > 0)) continue;
      waterWeights.push(...weights);
      for (const z of zCoords) {
        for (const x of xCoords) {
          positions.push(
            x,
            sampleTerrainHeight(grid, x, z) + WATER_SURFACE_OFFSET,
            z
          );
          uvs.push(originX + x, originY + z);
        }
      }

      const width = xCoords.length;
      for (let z = 0; z < zCoords.length - 1; z += 1) {
        for (let x = 0; x < xCoords.length - 1; x += 1) {
          const topLeft = vertexOffset + z * width + x;
          const topRight = topLeft + 1;
          const bottomLeft = topLeft + width;
          const bottomRight = bottomLeft + 1;
          indices.push(topLeft, bottomLeft, topRight, topRight, bottomLeft, bottomRight);
        }
      }
    }
  }

  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));
  geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
  geometry.setAttribute("waterWeight", new THREE.Float32BufferAttribute(waterWeights, 1));
  geometry.setIndex(indices);
  setTerrainNormals(geometry, grid);
  return geometry;
}

export function createTileHighlightGeometry(
  grid: TerrainHeightGrid,
  tileX: number,
  tileY: number
) {
  const xCoords = tileCoordinates(tileX);
  const zCoords = tileCoordinates(tileY);
  const positions: number[] = [];
  const indices: number[] = [];

  for (const z of zCoords) {
    for (const x of xCoords) {
      positions.push(
        x,
        sampleTerrainHeight(grid, x, z) + HIGHLIGHT_SURFACE_OFFSET,
        z
      );
    }
  }

  const width = xCoords.length;
  for (let z = 0; z < zCoords.length - 1; z += 1) {
    for (let x = 0; x < xCoords.length - 1; x += 1) {
      const topLeft = z * width + x;
      const topRight = topLeft + 1;
      const bottomLeft = topLeft + width;
      const bottomRight = bottomLeft + 1;
      indices.push(topLeft, bottomLeft, topRight, topRight, bottomLeft, bottomRight);
    }
  }

  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));
  geometry.setIndex(indices);
  geometry.computeVertexNormals();
  return geometry;
}

function lerp(from: number, to: number, amount: number) {
  return from + (to - from) * amount;
}

function setTerrainNormals(geometry: THREE.BufferGeometry, grid: TerrainHeightGrid) {
  const positions = geometry.getAttribute("position");
  const normals = new Float32Array(positions.count * 3);
  const normal = new THREE.Vector3();

  // Tiles duplicate edge vertices; colour sampling agrees across their edges. Sample the
  // height field on both sides so those vertices still share smooth normals.
  // Half-tile steps also fit within the worker's one-tile chunk border.
  for (let i = 0; i < positions.count; i += 1) {
    const x = positions.getX(i);
    const z = positions.getZ(i);
    const dx = sampleTerrainHeight(grid, x + 0.5, z) - sampleTerrainHeight(grid, x - 0.5, z);
    const dz = sampleTerrainHeight(grid, x, z + 0.5) - sampleTerrainHeight(grid, x, z - 0.5);
    normal.set(-dx, 1, -dz).normalize().toArray(normals, i * 3);
  }

  geometry.setAttribute("normal", new THREE.BufferAttribute(normals, 3));
}
