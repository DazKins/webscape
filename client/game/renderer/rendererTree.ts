import * as THREE from "three";
import Entity from "../entity/entity";
import {
  setTreeDamageStage,
  TREE_HIT_ANIMATION_SECONDS,
} from "../models/definitions/tree";
import ModelEntityRenderer from "./modelEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";

const TREE_DAMAGE_STAGE_COUNT = 4;

export default class RendererTree extends ModelEntityRenderer {
  private readonly standingTree: THREE.Object3D | undefined;
  private readonly stump: THREE.Object3D | undefined;
  private previousDurability: number | undefined;
  private shakeElapsedSeconds = TREE_HIT_ANIMATION_SECONDS;

  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler,
  ) {
    super(scene, entity, terrainHeightSampler, "tree");
    this.standingTree = this.modelInstance.root.getObjectByName("standingTree");
    this.stump = this.modelInstance.root.getObjectByName("stump");
    this.previousDurability = this.getDurability();
    this.updateTreeState(0);
  }

  update(deltaSeconds: number) {
    super.update(deltaSeconds);
    this.updateTreeState(deltaSeconds);
  }

  private updateTreeState(deltaSeconds: number) {
    const woodcuttable = this.entity.getComponent("woodcuttable");
    const durability = this.getDurability();
    if (
      durability !== undefined &&
      this.previousDurability !== undefined &&
      durability < this.previousDurability
    ) {
      this.shakeElapsedSeconds = 0;
      this.modelInstance.seek("hit", 0);
    }
    this.previousDurability = durability;

    const maximum = Number(woodcuttable?.maxDurability);
    const current = Number(woodcuttable?.currentDurability);
    const damageRatio = Number.isFinite(maximum) && maximum > 0 && Number.isFinite(current)
      ? THREE.MathUtils.clamp(1 - current / maximum, 0, 1)
      : 0;
    const damageStage = Math.min(
      TREE_DAMAGE_STAGE_COUNT,
      Math.ceil(damageRatio * TREE_DAMAGE_STAGE_COUNT),
    );
    setTreeDamageStage(this.modelInstance.root, damageStage);

    this.shakeElapsedSeconds = Math.min(
      TREE_HIT_ANIMATION_SECONDS,
      this.shakeElapsedSeconds + Math.max(0, deltaSeconds),
    );
    const isShaking = this.shakeElapsedSeconds < TREE_HIT_ANIMATION_SECONDS;

    const depleted = Boolean(woodcuttable?.depleted);
    const showStandingTree = !depleted || isShaking;
    if (this.standingTree) this.standingTree.visible = showStandingTree;
    if (this.stump) this.stump.visible = depleted && !isShaking;
  }

  private getDurability(): number | undefined {
    const durability = Number(
      this.entity.getComponent("woodcuttable")?.currentDurability,
    );
    return Number.isFinite(durability) ? durability : undefined;
  }

}
