import { constructionClient } from "../assets/constructionClient";
import type { ChunkBuild, ChunkSurfaces } from "../assets/construction";
import { unpackGeometry } from "../assets/geometryData";
import { retainResource, releaseResource, sharedGeometry } from "../models/assetCache";
import * as THREE from "three";
import Input from "../../input";
import Camera from "../camera";
import type { DeviceProfile, ViewportSize } from "../../responsive";
import { wallMaterial, type WorldWall } from "../renderer/rendererWall";
import {
  createTileHighlightGeometry,
  getTileHeight,
  sampleTerrainHeight,
  type TerrainHeightGrid,
} from "./terrainHeight";

export { terrainColor } from "../assets/construction";

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
  revision: number;
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
  private readonly pointerRaycaster = new THREE.Raycaster();
  private readonly pointerNdc = new THREE.Vector2();
  private readonly pointerViewProjection = new THREE.Matrix4();
  private readonly previousPointerViewProjection = new THREE.Matrix4();
  private readonly previousPointerNdc = new THREE.Vector2(Infinity, Infinity);
  private pointerTile: { x: number; y: number } | undefined;
  private pointerTileValid = false;
  private disposed = false;
  private building = false;
  private dirty = new Set<string>();
  private completed = new Map<string, { visual: ChunkVisual; revision: number; surfaces: ChunkSurfaces }>();

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
    this.pointerTileValid = false;
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
    // Queue builds after registering the full update, so each snapshot includes neighbors.
    for (const coordinate of affected.values()) {
      const visual = this.chunks.get(chunkKey(coordinate));
      if (visual) { visual.revision++; this.dirty.add(chunkKey(coordinate)); }
    }
    this.startNextBuild();
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
    const innerCamera = camera.getInnerCamera();
    // Picking happens before rendering, so refresh the camera's world matrix here.
    innerCamera.updateMatrixWorld();
    this.pointerNdc.set((pointer.x / viewport.width) * 2 - 1, -(pointer.y / viewport.height) * 2 + 1);
    this.pointerViewProjection.multiplyMatrices(innerCamera.projectionMatrix, innerCamera.matrixWorldInverse);
    if (
      this.pointerTileValid &&
      this.pointerNdc.equals(this.previousPointerNdc) &&
      this.pointerViewProjection.equals(this.previousPointerViewProjection)
    ) return this.pointerTile;

    this.previousPointerNdc.copy(this.pointerNdc);
    this.previousPointerViewProjection.copy(this.pointerViewProjection);
    this.pointerTileValid = true;
    this.pointerRaycaster.setFromCamera(this.pointerNdc, innerCamera);
    const hits = this.pointerRaycaster.intersectObjects([...this.chunks.values()].map((chunk) => chunk.terrainMesh), false);
    this.pointerTile = undefined;
    if (hits.length > 0) {
      const x = Math.floor(hits[0].point.x);
      const y = Math.floor(hits[0].point.z);
      const converted = this.globalToChunk(x, y);
      if (this.chunks.has(chunkKey(converted.coordinate))) this.pointerTile = { x, y };
    }
    return this.pointerTile;
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
    // Attach at most one completed chunk per frame to limit scene/GPU upload bursts.
    for (const [key, result] of this.completed) {
      this.completed.delete(key);
      if (this.chunks.get(key) !== result.visual || result.visual.revision !== result.revision) continue;
      this.installSurfaces(result.visual, result.surfaces);
      break;
    }
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

  dispose() {
    this.disposed = true;
    this.dirty.clear();
    this.completed.clear();
    for (const chunk of [...this.chunks.values()]) {
      this.disposeChunk(chunk.data.coordinate);
    }
    this.scene.remove(this.highlightMesh);
    this.highlightMesh.geometry.dispose();
    const material = this.highlightMesh.material;
    if (Array.isArray(material)) {
      material.forEach((value) => value.dispose());
    } else {
      material.dispose();
    }
  }

  private createChunk(data: ChunkLoad): ChunkVisual {
    const root = new THREE.Group();
    root.position.set(data.coordinate.x * this.chunkSize.x, 0, data.coordinate.y * this.chunkSize.y);
    this.scene.add(root);
    const grid = this.createGrid(data);
    const terrainMesh = new THREE.Mesh(new THREE.BufferGeometry(), new THREE.MeshPhongMaterial({ vertexColors: true, side: THREE.DoubleSide }));
    root.add(terrainMesh);
    return { data, root, terrainMesh, grid, revision: 0 };
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

  private startNextBuild() {
    if (this.disposed || this.building) return;
    for (const key of this.dirty) {
      this.dirty.delete(key);
      const visual = this.chunks.get(key);
      if (!visual) continue;
      const revision = visual.revision;
      const heights: number[] = [];
      const terrainBorder: (string | undefined)[] = [];
      for (let y = -1; y <= this.chunkSize.y; y++) {
        for (let x = -1; x <= this.chunkSize.x; x++) {
          // Worker input uses authored height units, including diagonal neighbors.
          const global = this.globalToChunk(visual.data.coordinate.x * this.chunkSize.x + x, visual.data.coordinate.y * this.chunkSize.y + y);
          const neighbor = this.chunks.get(chunkKey(global.coordinate));
          terrainBorder.push(neighbor?.data.terrain[global.local.y * this.chunkSize.x + global.local.x]);
          heights.push(neighbor?.data.heights[global.local.y * this.chunkSize.x + global.local.x] ?? 0);
        }
      }
      const chunk: ChunkBuild = { sizeX: this.chunkSize.x, sizeY: this.chunkSize.y,
        originX: visual.data.coordinate.x * this.chunkSize.x, originY: visual.data.coordinate.y * this.chunkSize.y,
        heights, terrainBorder, terrain: visual.data.terrain, walls: visual.data.walls ?? [] };
      this.building = true;
      constructionClient.run({ kind: "chunk", chunk }).then(result => {
        if (!this.disposed && this.chunks.get(key) === visual && visual.revision === revision && result.kind === "chunk") {
          this.completed.set(key, { visual, revision, surfaces: result.surfaces });
        }
      }).catch(error => console.error("Chunk construction failed", error)).finally(() => {
        this.building = false;
        this.startNextBuild();
      });
      break;
    }
  }

  private installSurfaces(chunk: ChunkVisual, surfaces: ChunkSurfaces) {
    chunk.terrainMesh.geometry.dispose();
    chunk.terrainMesh.geometry = unpackGeometry(surfaces.terrain);
    for (const name of ["chunkWater", "chunkWalls", "chunkDetails"]) {
      const old = chunk.root.getObjectByName(name);
      if (old) { disposeObject(old); chunk.root.remove(old); }
    }
    if (surfaces.details.attributes.position.array.length > 0) {
      const details = new THREE.Mesh(unpackGeometry(surfaces.details), new THREE.MeshPhongMaterial({ vertexColors: true, side: THREE.DoubleSide }));
      details.name = "chunkDetails";
      chunk.root.add(details);
    }
    chunk.waterMaterial = undefined;
    if (surfaces.water.attributes.position.array.length > 0) {
      chunk.waterMaterial = createWaterGlintMaterial();
      const water = new THREE.Mesh(unpackGeometry(surfaces.water), chunk.waterMaterial);
      water.name = "chunkWater";
      water.renderOrder = 1;
      chunk.root.add(water);
    }
    const walls = new THREE.Group();
    walls.name = "chunkWalls";
    const geometries = surfaces.wallGeometries.map(([key, data]) => sharedGeometry(key, () => unpackGeometry(data)));
    for (const part of surfaces.walls) {
      const geometry = geometries[part.geometry];
      const material = wallMaterial(part.type);
      retainResource(geometry); retainResource(material);
      const mesh = new THREE.Mesh(geometry, material);
      mesh.position.fromArray(part.position);
      mesh.quaternion.fromArray(part.quaternion);
      mesh.castShadow = true; mesh.receiveShadow = true;
      walls.add(mesh);
    }
    chunk.root.add(walls);
    this.pointerTileValid = false;
  }

  private disposeChunk(coordinate: ChunkCoordinate) {
    const key = chunkKey(coordinate);
    const chunk = this.chunks.get(key);
    if (!chunk) return;
    this.scene.remove(chunk.root);
    disposeObject(chunk.root);
    this.chunks.delete(key);
    this.dirty.delete(key);
    this.completed.delete(key);
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
    if (mesh.geometry) releaseResource(mesh.geometry);
    const materials = Array.isArray(mesh.material) ? mesh.material : mesh.material ? [mesh.material] : [];
    for (const material of materials) {
      releaseResource(material);
    }
  });
}

function createWaterGlintMaterial() {
  return new THREE.ShaderMaterial({
    uniforms: { time: { value: 0 } },
    vertexShader: `attribute float waterWeight; varying float vWaterWeight; varying vec2 vWaterUv; void main(){vWaterWeight=waterWeight;vWaterUv=uv;gl_Position=projectionMatrix*modelViewMatrix*vec4(position,1.0);}`,
    fragmentShader: `uniform float time; varying float vWaterWeight; varying vec2 vWaterUv; const float TAU=6.28318530718; float hash(vec2 p){return fract(sin(dot(p,vec2(127.1,311.7)))*43758.5453123);} float noise(vec2 p){vec2 c=floor(p);vec2 l=fract(p);l=l*l*(3.0-2.0*l);return mix(mix(hash(c),hash(c+vec2(1.,0.)),l.x),mix(hash(c+vec2(0.,1.)),hash(c+vec2(1.,1.)),l.x),l.y);} void main(){vec2 p=vWaterUv*3.4;float s=noise(p)*.52+noise(p*2.17+vec2(13.4,8.2))*.31+noise(p*4.61+vec2(5.7,21.9))*.17;float g=smoothstep(.58,.88,s)*(.62+.38*sin(time+noise(vWaterUv*.75+vec2(19.,3.))*TAU));g=clamp(g,0.,1.);gl_FragColor=vec4(mix(vec3(.58,.82,.92),vec3(.96,.99,1.),g),(.035+g*.12)*smoothstep(0.,1.,vWaterWeight));}`,
    transparent: true, depthWrite: false, side: THREE.DoubleSide,
  });
}

export default World;
