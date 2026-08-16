import * as THREE from "three";
import Entity from "../entity/entity";
import ModelEntityRenderer from "./modelEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";

export default class RendererTree extends ModelEntityRenderer {
  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler
  ) {
    super(scene, entity, terrainHeightSampler, "tree");
  }
}
