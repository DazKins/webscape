import * as THREE from "three";
import Input from "../../input";
import Camera from "../camera";
import type { DeviceProfile, ViewportSize } from "../../responsive";
import { addWallGeometry, type WorldWall } from "../renderer/rendererWall";
import {
  createTerrainSurfaceGeometry,
  createTileHighlightGeometry,
  createWaterSurfaceGeometry,
  getTileHeight,
  sampleTerrainHeight,
  type TerrainHeightGrid,
} from "./terrainHeight";

const WATER_GLINT_SPEED = 1.15;

export type ChunkCoordinate = { x: number; y: number };
export type ChunkLoad = {
  coordinate: ChunkCoordinate;
  terrain: string[];
  heights: number[];
  walls: WorldWall[];
};
export type ChunkUpdate = { load?: ChunkLoad[]; unload?: ChunkCoordinate[] };

type ChunkVisual = {
  data: ChunkLoad;
  root: THREE.Group;
  terrainMesh: THREE.Mesh;
  waterMaterial?: THREE.ShaderMaterial;
  grid: TerrainHeightGrid;
};

class World {
  readonly scene: THREE.Scene;
  readonly input: Input;
  readonly chunkSize: ChunkCoordinate;
  readonly chunks = new Map<string, ChunkVisual>();
  readonly highlightMesh: THREE.Mesh;
  highlightedTile: { x: number; y: number } | undefined;
  selectedTileSeconds = 0;
  waterAnimationTime = 0;

  constructor(scene: THREE.Scene, chunkSize: ChunkCoordinate, input: Input) {
    this.scene = scene;
    this.chunkSize = chunkSize;
    this.input = input;
    this.highlightMesh = new THREE.Mesh(
      new THREE.PlaneGeometry(1, 1),
      new THREE.MeshBasicMaterial({ color: 0xd8d8d8, transparent: true, opacity: 0.32, side: THREE.DoubleSide })
    );
    this.highlightMesh.renderOrder = 2;
    this.highlightMesh.visible = false;
    scene.add(this.highlightMesh);
  }

  applyChunkUpdate(update: ChunkUpdate) {
    const affected = new Map<string, ChunkCoordinate>();
    for (const coordinate of update.unload ?? []) {
      this.disposeChunk(coordinate);
      this.addAffected(affected, coordinate);
      if (this.highlightedTile && chunkKey(this.globalToChunk(this.highlightedTile.x, this.highlightedTile.y).coordinate) === chunkKey(coordinate)) {
        this.highlightMesh.visible = false;
        this.highlightedTile = undefined;
        this.selectedTileSeconds = 0;
      }
    }
    for (const data of update.load ?? []) {
      this.disposeChunk(data.coordinate);
      this.chunks.set(chunkKey(data.coordinate), this.createChunk(data));
      this.addAffected(affected, data.coordinate);
    }
    // Terrain vertices sample adjacent chunks, so rebuild changed chunks and their neighbors.
    for (const coordinate of affected.values()) {
      const visual = this.chunks.get(chunkKey(coordinate));
      if (visual) this.rebuildSurfaces(visual);
    }
  }

  getVisualHeightAtWorldPosition(worldX: number, worldZ: number) {
    const tile = this.globalToChunk(Math.floor(worldX), Math.floor(worldZ));
    const visual = this.chunks.get(chunkKey(tile.coordinate));
    if (!visual) return 0;
    return sampleTerrainHeight(visual.grid, worldX - tile.origin.x, worldZ - tile.origin.y);
  }

  getVisualHeightAtTile(tileX: number, tileY: number) {
    const tile = this.globalToChunk(tileX, tileY);
    const visual = this.chunks.get(chunkKey(tile.coordinate));
    return visual ? getTileHeight(visual.grid, tile.local.x, tile.local.y) : 0;
  }

  getPointerTile(camera: Camera, viewport: ViewportSize) {
    if (this.input.isPointerBlocked()) return undefined;
    const pointer = this.input.getPointerPosition();
    const raycaster = new THREE.Raycaster();
    raycaster.setFromCamera(
      new THREE.Vector2((pointer.x / viewport.width) * 2 - 1, -(pointer.y / viewport.height) * 2 + 1),
      camera.getInnerCamera()
    );
    const hits = raycaster.intersectObjects([...this.chunks.values()].map((chunk) => chunk.terrainMesh), false);
    if (hits.length === 0) return undefined;
    const x = Math.floor(hits[0].point.x);
    const y = Math.floor(hits[0].point.z);
    const converted = this.globalToChunk(x, y);
    return this.chunks.has(chunkKey(converted.coordinate)) ? { x, y } : undefined;
  }

  showTileIndicator(tile: { x: number; y: number }) {
    const converted = this.globalToChunk(tile.x, tile.y);
    const visual = this.chunks.get(chunkKey(converted.coordinate));
    if (!visual) return;
    this.highlightMesh.geometry.dispose();
    this.highlightMesh.geometry = createTileHighlightGeometry(visual.grid, converted.local.x, converted.local.y);
    this.highlightMesh.position.set(converted.origin.x, 0, converted.origin.y);
    this.highlightMesh.visible = true;
    this.highlightedTile = tile;
    this.selectedTileSeconds = 0.34;
  }

  update(camera: Camera, deltaSeconds: number, profile: DeviceProfile) {
    this.waterAnimationTime += deltaSeconds;
    for (const chunk of this.chunks.values()) {
      if (chunk.waterMaterial) chunk.waterMaterial.uniforms.time.value = this.waterAnimationTime * WATER_GLINT_SPEED;
    }
    if (profile.canHover && !profile.isCoarsePointer) {
      const hovered = this.getPointerTile(camera, profile);
      if (hovered && (!this.highlightedTile || hovered.x !== this.highlightedTile.x || hovered.y !== this.highlightedTile.y)) {
        this.showTileIndicator(hovered);
      } else if (!hovered) {
        this.highlightMesh.visible = false;
        this.highlightedTile = undefined;
      }
      return;
    }
    this.selectedTileSeconds = Math.max(0, this.selectedTileSeconds - deltaSeconds);
    this.highlightMesh.visible = this.selectedTileSeconds > 0;
    if (!this.highlightMesh.visible) this.highlightedTile = undefined;
  }

  private createChunk(data: ChunkLoad): ChunkVisual {
    const root = new THREE.Group();
    root.position.set(data.coordinate.x * this.chunkSize.x, 0, data.coordinate.y * this.chunkSize.y);
    this.scene.add(root);
    const grid = this.createGrid(data);
    const terrainMesh = new THREE.Mesh(new THREE.BufferGeometry(), new THREE.MeshPhongMaterial({ vertexColors: true, side: THREE.DoubleSide }));
    root.add(terrainMesh);
    addWallGeometry(root, data.walls ?? [], (x, z) => sampleTerrainHeight(grid, x, z));
    const visual = { data, root, terrainMesh, grid } as ChunkVisual;
    this.rebuildSurfaces(visual);
    return visual;
  }

  private createGrid(data: ChunkLoad): TerrainHeightGrid {
    return {
      sizeX: this.chunkSize.x,
      sizeY: this.chunkSize.y,
      heights: data.heights,
      sampleOutside: (x, y) => {
        const globalX = data.coordinate.x * this.chunkSize.x + x;
        const globalY = data.coordinate.y * this.chunkSize.y + y;
        const converted = this.globalToChunk(globalX, globalY);
        const neighbor = this.chunks.get(chunkKey(converted.coordinate));
        if (!neighbor) return 0;
        return getTileHeight(neighbor.grid, converted.local.x, converted.local.y);
      },
    };
  }

  private rebuildSurfaces(chunk: ChunkVisual) {
    chunk.terrainMesh.geometry.dispose();
    chunk.terrainMesh.geometry = createTerrainSurfaceGeometry(chunk.grid, chunk.data.terrain, terrainColor);
    const oldWater = chunk.root.getObjectByName("chunkWater") as THREE.Mesh | undefined;
    if (oldWater) { disposeObject(oldWater); chunk.root.remove(oldWater); }
    chunk.waterMaterial = undefined;
    const waterGeometry = createWaterSurfaceGeometry(chunk.grid, chunk.data.terrain);
    if ((waterGeometry.getAttribute("position")?.count ?? 0) === 0) { waterGeometry.dispose(); return; }
    chunk.waterMaterial = createWaterGlintMaterial();
    const water = new THREE.Mesh(waterGeometry, chunk.waterMaterial);
    water.name = "chunkWater";
    water.renderOrder = 1;
    chunk.root.add(water);
  }

  private disposeChunk(coordinate: ChunkCoordinate) {
    const key = chunkKey(coordinate);
    const chunk = this.chunks.get(key);
    if (!chunk) return;
    this.scene.remove(chunk.root);
    disposeObject(chunk.root);
    this.chunks.delete(key);
  }

  private addAffected(result: Map<string, ChunkCoordinate>, coordinate: ChunkCoordinate) {
    for (let y = coordinate.y - 1; y <= coordinate.y + 1; y += 1) {
      for (let x = coordinate.x - 1; x <= coordinate.x + 1; x += 1) {
        const next = { x, y }; result.set(chunkKey(next), next);
      }
    }
  }

  private globalToChunk(x: number, y: number) {
    const coordinate = { x: Math.floor(x / this.chunkSize.x), y: Math.floor(y / this.chunkSize.y) };
    const origin = { x: coordinate.x * this.chunkSize.x, y: coordinate.y * this.chunkSize.y };
    return { coordinate, origin, local: { x: x - origin.x, y: y - origin.y } };
  }
}

function chunkKey(coordinate: ChunkCoordinate) { return `${coordinate.x},${coordinate.y}`; }

function disposeObject(object: THREE.Object3D) {
  object.traverse((child) => {
    const mesh = child as THREE.Mesh;
    mesh.geometry?.dispose();
    const materials = Array.isArray(mesh.material) ? mesh.material : mesh.material ? [mesh.material] : [];
    for (const material of materials) {
      for (const value of Object.values(material)) if (value instanceof THREE.Texture) value.dispose();
      material.dispose();
    }
  });
}

function terrainColor(type: string) {
  switch (type) {
    case "grass": return 0x73964f;
    case "dirt": return 0x9a6b42;
    case "road": return 0xb8ab88;
    case "water": return 0x4f8fb8;
    case "stone": return 0x8b9296;
    default: return 0xe77d11;
  }
}

function createWaterGlintMaterial() {
  return new THREE.ShaderMaterial({
    uniforms: { time: { value: 0 } },
    vertexShader: `varying vec2 vWaterUv; void main(){vWaterUv=uv;gl_Position=projectionMatrix*modelViewMatrix*vec4(position,1.0);}`,
    fragmentShader: `uniform float time; varying vec2 vWaterUv; const float TAU=6.28318530718; float hash(vec2 p){return fract(sin(dot(p,vec2(127.1,311.7)))*43758.5453123);} float noise(vec2 p){vec2 c=floor(p);vec2 l=fract(p);l=l*l*(3.0-2.0*l);return mix(mix(hash(c),hash(c+vec2(1.,0.)),l.x),mix(hash(c+vec2(0.,1.)),hash(c+vec2(1.,1.)),l.x),l.y);} void main(){vec2 p=vWaterUv*3.4;float s=noise(p)*.52+noise(p*2.17+vec2(13.4,8.2))*.31+noise(p*4.61+vec2(5.7,21.9))*.17;float g=smoothstep(.58,.88,s)*(.62+.38*sin(time+noise(vWaterUv*.75+vec2(19.,3.))*TAU));g=clamp(g,0.,1.);gl_FragColor=vec4(mix(vec3(.58,.82,.92),vec3(.96,.99,1.),g),.035+g*.12);}`,
    transparent: true, depthWrite: false, side: THREE.DoubleSide,
  });
}

export default World;
