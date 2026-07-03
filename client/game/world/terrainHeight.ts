import * as THREE from "three";

export const TERRAIN_HEIGHT_SCALE = 0.2;

const HIGHLIGHT_SURFACE_OFFSET = 0.025;
const WATER_SURFACE_OFFSET = 0.012;

export type TerrainHeightGrid = {
  sizeX: number;
  sizeY: number;
  heights: number[];
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
    return 0;
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

  const sampleX = clamp(worldX, 0.5, grid.sizeX - 0.5);
  const sampleZ = clamp(worldZ, 0.5, grid.sizeY - 0.5);
  const centerX = sampleX - 0.5;
  const centerZ = sampleZ - 0.5;
  const x0 = Math.floor(centerX);
  const z0 = Math.floor(centerZ);
  const x1 = Math.min(x0 + 1, grid.sizeX - 1);
  const z1 = Math.min(z0 + 1, grid.sizeY - 1);
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
  terrainColorForTile: (terrainType: string) => THREE.ColorRepresentation
) {
  const positions: number[] = [];
  const colors: number[] = [];
  const indices: number[] = [];
  const color = new THREE.Color();

  for (let tileY = 0; tileY < grid.sizeY; tileY += 1) {
    for (let tileX = 0; tileX < grid.sizeX; tileX += 1) {
      const xCoords = [tileX, tileX + 0.5, tileX + 1];
      const zCoords = [tileY, tileY + 0.5, tileY + 1];
      const vertexOffset = positions.length / 3;
      const terrainType = terrain[tileY * grid.sizeX + tileX] ?? "";

      color.set(terrainColorForTile(terrainType));

      for (const z of zCoords) {
        for (const x of xCoords) {
          positions.push(x, sampleTerrainHeight(grid, x, z), z);
          colors.push(color.r, color.g, color.b);
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
  geometry.computeVertexNormals();
  return geometry;
}

export function createWaterSurfaceGeometry(
  grid: TerrainHeightGrid,
  terrain: string[]
) {
  const positions: number[] = [];
  const uvs: number[] = [];
  const indices: number[] = [];

  for (let tileY = 0; tileY < grid.sizeY; tileY += 1) {
    for (let tileX = 0; tileX < grid.sizeX; tileX += 1) {
      const terrainType = terrain[tileY * grid.sizeX + tileX] ?? "";
      if (terrainType !== "water") {
        continue;
      }

      const xCoords = [tileX, tileX + 0.5, tileX + 1];
      const zCoords = [tileY, tileY + 0.5, tileY + 1];
      const vertexOffset = positions.length / 3;

      for (const z of zCoords) {
        for (const x of xCoords) {
          positions.push(
            x,
            sampleTerrainHeight(grid, x, z) + WATER_SURFACE_OFFSET,
            z
          );
          uvs.push(x, z);
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
  geometry.setIndex(indices);
  geometry.computeVertexNormals();
  return geometry;
}

export function createTileHighlightGeometry(
  grid: TerrainHeightGrid,
  tileX: number,
  tileY: number
) {
  const xCoords = [tileX, tileX + 0.5, tileX + 1];
  const zCoords = [tileY, tileY + 0.5, tileY + 1];
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

function clamp(value: number, min: number, max: number) {
  if (max < min) {
    return min;
  }

  return Math.min(Math.max(value, min), max);
}

function lerp(from: number, to: number, amount: number) {
  return from + (to - from) * amount;
}
