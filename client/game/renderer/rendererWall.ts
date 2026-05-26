import * as THREE from "three";
import { RoundedBoxGeometry } from "three/examples/jsm/geometries/RoundedBoxGeometry.js";

export type WorldWall = {
  id: string;
  type: string;
  x: number;
  y: number;
};

type WallTileIndex = ReadonlyMap<string, string>;

const WALL_HEIGHT = 1.0;
const WALL_THICKNESS = 0.34;
const PILLAR_SIZE = 0.58;
const PILLAR_HEIGHT = 1.08;
const BEVEL_RADIUS = 0.055;

export function addWallGeometry(scene: THREE.Scene, walls: WorldWall[]) {
  const wallTiles = createWallTileIndex(walls);
  const materials = new Map<string, THREE.MeshStandardMaterial>();

  for (const wall of walls) {
    const material = getWallMaterial(materials, wall.type);
    const neighbours = {
      north: hasWallTile(wallTiles, wall.x, wall.y - 1, wall.type),
      east: hasWallTile(wallTiles, wall.x + 1, wall.y, wall.type),
      south: hasWallTile(wallTiles, wall.x, wall.y + 1, wall.type),
      west: hasWallTile(wallTiles, wall.x - 1, wall.y, wall.type),
    };

    addWallTile(scene, material, wall.x, wall.y, neighbours);
  }
}

function createWallTileIndex(walls: WorldWall[]): WallTileIndex {
  const wallTiles = new Map<string, string>();
  for (const wall of walls) {
    wallTiles.set(wallTileKey(wall.x, wall.y), wall.type);
  }
  return wallTiles;
}

function addWallTile(
  scene: THREE.Scene,
  material: THREE.MeshStandardMaterial,
  x: number,
  y: number,
  neighbours: { north: boolean; east: boolean; south: boolean; west: boolean }
) {
  const horizontal = neighbours.east || neighbours.west;
  const vertical = neighbours.north || neighbours.south;
  const connectionCount = Object.values(neighbours).filter(Boolean).length;

  if (!horizontal && !vertical) {
    addPillar(scene, material, x, y);
    return;
  }

  if (horizontal) {
    const minX = neighbours.west ? 0 : 0.5 - PILLAR_SIZE / 2;
    const maxX = neighbours.east ? 1 : 0.5 + PILLAR_SIZE / 2;
    addBox(
      scene,
      material,
      x + (minX + maxX) / 2,
      y + 0.5,
      maxX - minX,
      WALL_HEIGHT,
      WALL_THICKNESS
    );
  }

  if (vertical) {
    const minY = neighbours.north ? 0 : 0.5 - PILLAR_SIZE / 2;
    const maxY = neighbours.south ? 1 : 0.5 + PILLAR_SIZE / 2;
    addBox(
      scene,
      material,
      x + 0.5,
      y + (minY + maxY) / 2,
      WALL_THICKNESS,
      WALL_HEIGHT,
      maxY - minY
    );
  }

  if (connectionCount !== 2 || horizontal === vertical) {
    addPillar(scene, material, x, y);
  }
}

function addPillar(
  scene: THREE.Scene,
  material: THREE.MeshStandardMaterial,
  x: number,
  y: number
) {
  addBox(scene, material, x + 0.5, y + 0.5, PILLAR_SIZE, PILLAR_HEIGHT, PILLAR_SIZE);
}

function addBox(
  scene: THREE.Scene,
  material: THREE.MeshStandardMaterial,
  x: number,
  z: number,
  width: number,
  height: number,
  depth: number
) {
  const wall = new THREE.Mesh(
    new RoundedBoxGeometry(width, height, depth, 2, BEVEL_RADIUS),
    material
  );
  wall.position.set(x, height / 2, z);
  scene.add(wall);
}

function getWallMaterial(
  materials: Map<string, THREE.MeshStandardMaterial>,
  type: string
): THREE.MeshStandardMaterial {
  const existing = materials.get(type);
  if (existing) {
    return existing;
  }

  const material = new THREE.MeshStandardMaterial({
    color: 0xffffff,
    map: createWallTexture(type),
    roughness: 0.78,
    metalness: 0.0,
  });
  materials.set(type, material);
  return material;
}

function createWallTexture(type: string): THREE.CanvasTexture {
  const size = 128;
  const canvas = document.createElement("canvas");
  canvas.width = size;
  canvas.height = size;

  const ctx = canvas.getContext("2d");
  if (!ctx) {
    throw new Error("failed to create wall texture");
  }

  const base = new THREE.Color(wallColor(type));
  ctx.fillStyle = colorToCss(base);
  ctx.fillRect(0, 0, size, size);

  for (let y = 0; y < size; y += 1) {
    const shade = (noise(y, type.length) - 0.5) * 0.09;
    ctx.fillStyle = colorToCss(offsetColor(base, shade));
    ctx.fillRect(0, y, size, 1);
  }

  for (let x = 0; x < size; x += 8) {
    const shade = (noise(x, type.charCodeAt(0) || 1) - 0.5) * 0.12;
    ctx.fillStyle = colorToCss(offsetColor(base, shade));
    ctx.globalAlpha = 0.22;
    ctx.fillRect(x, 0, 2, size);
  }
  ctx.globalAlpha = 1;

  const texture = new THREE.CanvasTexture(canvas);
  texture.wrapS = THREE.RepeatWrapping;
  texture.wrapT = THREE.RepeatWrapping;
  texture.colorSpace = THREE.SRGBColorSpace;
  texture.anisotropy = 4;
  return texture;
}

function offsetColor(color: THREE.Color, amount: number): THREE.Color {
  return new THREE.Color(
    clamp(color.r + amount),
    clamp(color.g + amount),
    clamp(color.b + amount)
  );
}

function colorToCss(color: THREE.Color): string {
  return `rgb(${Math.round(color.r * 255)}, ${Math.round(color.g * 255)}, ${Math.round(color.b * 255)})`;
}

function clamp(value: number): number {
  return Math.min(1, Math.max(0, value));
}

function noise(value: number, seed: number): number {
  const x = Math.sin(value * 12.9898 + seed * 78.233) * 43758.5453;
  return x - Math.floor(x);
}

function hasWallTile(wallTiles: WallTileIndex, x: number, y: number, type: string): boolean {
  return wallTiles.get(wallTileKey(x, y)) === type;
}

function wallTileKey(x: number, y: number): string {
  return `${x},${y}`;
}

function wallColor(type: string): number {
  switch (type) {
    case "wood":
      return 0x8a5a34;
    case "stone":
      return 0x8b9296;
    default:
      return 0xd3d3d3;
  }
}
