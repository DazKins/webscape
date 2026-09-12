import * as THREE from "three";
import Entity from "../entity/entity";
import ModelEntityRenderer from "./modelEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";
import { CHEST_OPEN_SECONDS } from "../models/definitions/chest";

export default class RendererChest extends ModelEntityRenderer {
  private openPhase: number;
  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler
  ) {
    super(scene, entity, terrainHeightSampler, "chest");
    this.openPhase = entity.getComponent("openable")?.isOpen ? 1 : 0;
    this.modelInstance.seek("open", this.openPhase);
  }

  update(deltaSeconds: number) {
    super.update(deltaSeconds);
    const direction = this.entity.getComponent("openable")?.isOpen ? 1 : -1;
    this.openPhase = THREE.MathUtils.clamp(
      this.openPhase + direction * Math.max(0, deltaSeconds) / CHEST_OPEN_SECONDS, 0, 1,
    );
    this.modelInstance.seek("open", this.openPhase);
  }
}
