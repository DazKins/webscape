import * as THREE from "three";
import Entity from "../entity/entity";

export type TerrainHeightSampler = (worldX: number, worldZ: number) => number;

const flatTerrainHeightSampler: TerrainHeightSampler = () => 0;

export default class EntityRenderer {
  scene: THREE.Scene;
  entity: Entity;
  private terrainHeightSampler: TerrainHeightSampler;

  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler: TerrainHeightSampler = flatTerrainHeightSampler
  ) {
    this.scene = scene;
    this.entity = entity;
    this.terrainHeightSampler = terrainHeightSampler;
  }

  update(_deltaSeconds: number) {
    throw Error("update not implemented");
  }

  getObject3D(): THREE.Object3D | null {
    return null;
  }

  onRemove() {
    throw Error("onRemove not implemented");
  }

  protected sampleVisualHeight(worldX: number, worldZ: number) {
    return this.terrainHeightSampler(worldX, worldZ);
  }
}
