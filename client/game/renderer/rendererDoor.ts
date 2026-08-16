import * as THREE from "three";
import Entity from "../entity/entity";
import ModelEntityRenderer from "./modelEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";

const DOOR_ANIMATION_SECONDS = 0.28;

export default class RendererDoor extends ModelEntityRenderer {
  private openPhase: number;

  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler
  ) {
    super(scene, entity, terrainHeightSampler, "door");
    this.openPhase = entity.getComponent("openable")?.isOpen ? 1 : 0;
    this.modelInstance.seek("open", this.openPhase);
  }

  update(deltaSeconds: number) {
    super.update(deltaSeconds);

    const openable = this.entity.getComponent("openable");
    const isOpen = Boolean(openable?.isOpen);
    const phaseStep = deltaSeconds / DOOR_ANIMATION_SECONDS;
    this.openPhase = THREE.MathUtils.clamp(
      this.openPhase + (isOpen ? phaseStep : -phaseStep),
      0,
      1,
    );
    this.modelInstance.root.rotation.y = this.getOrientationRotation();
    this.modelInstance.seek("open", this.openPhase);
  }

  private getOrientationRotation() {
    const renderable = this.entity.getComponent("renderable");
    switch (renderable?.orientation) {
      case "east":
        return Math.PI / 2;
      case "south":
        return Math.PI;
      case "west":
        return -Math.PI / 2;
      case "north":
      default:
        return 0;
    }
  }
}
