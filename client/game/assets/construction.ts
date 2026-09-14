import * as THREE from "three";
import { geometryAssets, sharedGeometry } from "../models/assetCache";
import { modelNames, modelRegistry, type ModelName } from "../models/registry";
import { addWallGeometry, type WorldWall } from "../renderer/rendererWall";
import { createTerrainSurfaceGeometry, createWaterSurfaceGeometry, sampleTerrainHeight, TERRAIN_HEIGHT_SCALE, type TerrainHeightGrid } from "../world/terrainHeight";
import type { ModelOptions } from "../models/types";
import { packGeometry, type GeometryData } from "./geometryData";

export type ChunkBuild = {
  sizeX: number; sizeY: number;
  // A one-tile border makes sampling independent of live main-thread world data.
  heights: number[];
  terrain: string[];
  walls: WorldWall[];
};
export type WallPart = { geometry: number; type: string; position: number[]; quaternion: number[] };
export type ChunkSurfaces = { terrain: GeometryData; water: GeometryData; wallGeometries: [string, GeometryData][]; walls: WallPart[] };
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
  const terrain = createTerrainSurfaceGeometry(grid, chunk.terrain, terrainColor);
  const water = createWaterSurfaceGeometry(grid, chunk.terrain);
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
  const surfaces = { terrain: packGeometry(terrain), water: packGeometry(water),
    wallGeometries: [...geometries.keys()].map((geometry): [string, GeometryData] => [geometry.userData.assetKey, packGeometry(geometry)]), walls };
  terrain.dispose(); water.dispose();
  return { kind: "chunk", surfaces };
}

export function installModelGeometry(geometries: [string, GeometryData][], unpack: (data: GeometryData) => THREE.BufferGeometry) {
  for (const [key, data] of geometries) sharedGeometry(key, () => unpack(data));
}

export function terrainColor(type: string) {
  switch (type) {
    case "grass": return 0x73964f;
    case "dirt": return 0x9a6b42;
    case "road": return 0xb8ab88;
    case "water": return 0x4f8fb8;
    case "stone": return 0x8b9296;
    default: return 0xe77d11;
  }
}
