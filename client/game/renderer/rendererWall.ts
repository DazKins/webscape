import * as THREE from "three";

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

export function addWallGeometry(scene: THREE.Scene, walls: WorldWall[]) {
  const wallTiles = createWallTileIndex(walls);
  const materials = new Map<string, THREE.MeshPhongMaterial>();

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
  material: THREE.MeshPhongMaterial,
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
  material: THREE.MeshPhongMaterial,
  x: number,
  y: number
) {
  addBox(scene, material, x + 0.5, y + 0.5, PILLAR_SIZE, PILLAR_HEIGHT, PILLAR_SIZE);
}

function addBox(
  scene: THREE.Scene,
  material: THREE.MeshPhongMaterial,
  x: number,
  z: number,
  width: number,
  height: number,
  depth: number
) {
  const wall = new THREE.Mesh(new THREE.BoxGeometry(width, height, depth), material);
  wall.position.set(x, height / 2, z);
  scene.add(wall);
}

function getWallMaterial(
  materials: Map<string, THREE.MeshPhongMaterial>,
  type: string
): THREE.MeshPhongMaterial {
  const existing = materials.get(type);
  if (existing) {
    return existing;
  }

  const material = new THREE.MeshPhongMaterial({ color: wallColor(type) });
  materials.set(type, material);
  return material;
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
