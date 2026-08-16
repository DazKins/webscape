import * as THREE from "three";
import Entity from "../entity/entity";
import ModelEntityRenderer from "./modelEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";

export default class RendererBuilding extends ModelEntityRenderer {
  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler
  ) {
    const metadata = entity.getComponent("metadata") ?? {};
    super(scene, entity, terrainHeightSampler, "building", {
      width: metadata.width ?? 1,
      height: metadata.height ?? 1,
    });
  }
}
