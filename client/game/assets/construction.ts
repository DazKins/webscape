import * as THREE from "three";
import { createTerrainBlend } from "../world/terrainBlend";
import { variedTerrainColor } from "../world/terrainAppearance";
import { createTerrainDetails } from "../world/terrainDetails";
export { terrainColor } from "../world/terrainAppearance";
import { geometryAssets, sharedGeometry } from "../models/assetCache";
import { modelNames, modelRegistry, type ModelName } from "../models/registry";
import { addWallGeometry, type WorldWall } from "../renderer/rendererWall";
import { createTerrainSurfaceGeometry, createWaterSurfaceGeometry, sampleTerrainHeight, TERRAIN_HEIGHT_SCALE, type TerrainHeightGrid } from "../world/terrainHeight";
import type { ModelOptions } from "../models/types";
import { packGeometry, type GeometryData } from "./geometryData";

export type ChunkBuild = {
  sizeX: number; sizeY: number;
  originX?: number; originY?: number;
  // A one-tile border makes sampling independent of live main-thread world data.
  heights: number[];
  terrain: string[];
  terrainBorder: (string | undefined)[];
  walls: WorldWall[];
};
export type WallPart = { geometry: number; type: string; position: number[]; quaternion: number[] };
export type ChunkSurfaces = { terrain: GeometryData; details: GeometryData; water: GeometryData; wallGeometries: [string, GeometryData][]; walls: WallPart[] };
export type ConstructionRequest = { kind: "warmModels" } | { kind: "model"; name: ModelName; options: ModelOptions } | { kind: "chunk"; chunk: ChunkBuild };
export type ConstructionResult = { kind: "warmModels"; geometries: [string, GeometryData][] } | { kind: "chunk"; surfaces: ChunkSurfaces };

export function construct(request: ConstructionRequest): ConstructionResult {
  if (request.kind === "model") {
    const model = modelRegistry[request.name](request.options);
    const geometries = new Map<string, THREE.BufferGeometry>();
    model.root.traverse(object => {
      if (object instanceof THREE.Mesh && object.geometry.userData.assetKey) geometries.set(object.geometry.userData.assetKey, object.geometry);
    });
    const result: ConstructionResult = { kind: "warmModels", geometries: [...geometries].map(([key, geometry]) => [key, packGeometry(geometry)]) };
    model.dispose();
    return result;
  }
  if (request.kind === "warmModels") {
    for (const name of modelNames) modelRegistry[name]().dispose();
    for (const hairStyle of ["cropped", "swept", "bob", "curls"] as const) {
      modelRegistry.human({ appearance: { hairStyle, skinTone: "fair", hairColor: "darkBrown", tunicColor: "slateBlue", trousersColor: "navy", shoeColor: "darkBrown" } }).dispose();
    }
    return { kind: "warmModels", geometries: geometryAssets.values().map(([key, geometry]) => [key, packGeometry(geometry)]) };
  }
  const { chunk } = request;
  const grid: TerrainHeightGrid = {
    sizeX: chunk.sizeX, sizeY: chunk.sizeY,
    heights: Array.from({ length: chunk.sizeX * chunk.sizeY }, (_, index) => chunk.heights[(Math.floor(index / chunk.sizeX) + 1) * (chunk.sizeX + 2) + index % chunk.sizeX + 1]),
    sampleOutside: (x, y) => (chunk.heights[(y + 1) * (chunk.sizeX + 2) + x + 1] ?? 0) * TERRAIN_HEIGHT_SCALE,
  };
  const blend = createTerrainBlend(chunk.sizeX, chunk.sizeY, chunk.terrain, chunk.terrainBorder, chunk.originX ?? 0, chunk.originY ?? 0);
  const terrain = createTerrainSurfaceGeometry(grid, chunk.terrain, (type, x, y) => variedTerrainColor(type, x + (chunk.originX ?? 0), y + (chunk.originY ?? 0)), blend);
  const details = createTerrainDetails(grid, chunk.terrain, chunk.originX ?? 0, chunk.originY ?? 0, blend);
  const water = createWaterSurfaceGeometry(grid, chunk.terrain, blend, chunk.originX ?? 0, chunk.originY ?? 0);
  const root = new THREE.Group();
  addWallGeometry(root, chunk.walls, (x, z) => sampleTerrainHeight(grid, x, z), false);
  const geometries = new Map<THREE.BufferGeometry, number>();
  const walls: WallPart[] = [];
  root.traverse(object => {
    if (!(object instanceof THREE.Mesh)) return;
    if (!geometries.has(object.geometry)) geometries.set(object.geometry, geometries.size);
    walls.push({ geometry: geometries.get(object.geometry)!, type: object.material.userData.wallType,
      position: object.position.toArray(), quaternion: object.quaternion.toArray() });
  });
  const surfaces = { terrain: packGeometry(terrain), details: packGeometry(details), water: packGeometry(water),
    wallGeometries: [...geometries.keys()].map((geometry): [string, GeometryData] => [geometry.userData.assetKey, packGeometry(geometry)]), walls };
  terrain.dispose(); details.dispose(); water.dispose();
  return { kind: "chunk", surfaces };
}

export function installModelGeometry(geometries: [string, GeometryData][], unpack: (data: GeometryData) => THREE.BufferGeometry) {
  for (const [key, data] of geometries) sharedGeometry(key, () => unpack(data));
}
