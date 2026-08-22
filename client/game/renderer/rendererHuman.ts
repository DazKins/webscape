import * as THREE from "three";
import Entity from "../entity/entity";
import { createModel, type ModelInstance } from "../models";
import EntityHealthBar from "../../ui/components/entityHealthBar";
import { createReactCss2dObject, type ReactCss2dObject } from "../../util/reactCss2dObject";
import EntityRenderer, { type TerrainHeightSampler } from "./renderer";
import EquipmentAttachmentController from "./equipmentAttachmentController";

const SERVER_TICK_SECONDS = 0.52;
const HUMAN_HEALTH_BAR_Y = 1.55;
const HUMAN_IDLE_ANIMATION_NAME = "idle";
const HUMAN_RUN_ANIMATION_NAME = "run";
const HUMAN_ANIMATION_FADE_SECONDS = 0.12;
const HUMAN_ATTACK_ANIMATION_SECONDS = 0.38;
const HUMAN_MODEL_FORWARD_ROTATION_OFFSET = 0;
const HUMAN_ROTATION_SPEED_RADIANS_PER_SECOND = 10;

export default class RendererHuman extends EntityRenderer {
  mesh: THREE.Group;
  healthBar: ReactCss2dObject<{ currentHealth: number; maxHealth: number }> | null = null;
  private readonly visualRoot: THREE.Group;
  private readonly modelInstance: ModelInstance;
  private readonly equipmentAttachments: EquipmentAttachmentController;
  private segmentStartX: number;
  private segmentStartZ: number;
  private segmentTargetX: number;
  private segmentTargetZ: number;
  private segmentElapsedSeconds = SERVER_TICK_SECONDS;
  private targetRotationY = HUMAN_MODEL_FORWARD_ROTATION_OFFSET;
  private attackAnimationSecondsRemaining = 0;
  private readonly resolveEntity: (entityId: string) => Entity | undefined;

  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler,
    resolveEntity: (entityId: string) => Entity | undefined = () => undefined,
  ) {
    super(scene, entity, terrainHeightSampler);
    this.resolveEntity = resolveEntity;

    const metadata = entity.getComponent("metadata") ?? {};
    this.mesh = new THREE.Group();
    this.mesh.userData.entityId = entity.getId();
    this.visualRoot = new THREE.Group();
    this.visualRoot.position.set(0.5, 0, 0.5);
    this.mesh.add(this.visualRoot);

    this.modelInstance = createModel("human", { color: metadata.color ?? "#00ff00" });
    this.visualRoot.add(this.modelInstance.root);
    this.equipmentAttachments = new EquipmentAttachmentController(this.modelInstance);

    const position = entity.getComponent("position");
    this.mesh.position.set(
      position.x,
      this.sampleVisualHeight(position.x + 0.5, position.y + 0.5),
      position.y,
    );
    this.segmentStartX = position.x;
    this.segmentStartZ = position.y;
    this.segmentTargetX = position.x;
    this.segmentTargetZ = position.y;
    this.scene.add(this.mesh);
  }

  update(deltaSeconds: number) {
    const position = this.entity.getComponent("position");
    if (!position) {
      return;
    }

    const targetX = position.x;
    const targetZ = position.y;
    if (targetX !== this.segmentTargetX || targetZ !== this.segmentTargetZ) {
      this.segmentStartX = this.mesh.position.x;
      this.segmentStartZ = this.mesh.position.z;
      this.segmentTargetX = targetX;
      this.segmentTargetZ = targetZ;
      this.segmentElapsedSeconds = 0;
      this.faceDirection(targetX - this.segmentStartX, targetZ - this.segmentStartZ);
    }

    this.segmentElapsedSeconds += deltaSeconds;
    const progress = Math.min(this.segmentElapsedSeconds / SERVER_TICK_SECONDS, 1);
    const renderedX = THREE.MathUtils.lerp(this.segmentStartX, this.segmentTargetX, progress);
    const renderedZ = THREE.MathUtils.lerp(this.segmentStartZ, this.segmentTargetZ, progress);
    this.mesh.position.set(
      renderedX,
      this.sampleVisualHeight(renderedX + 0.5, renderedZ + 0.5),
      renderedZ,
    );
    if (!this.isMoving()) {
      this.faceSynchronizedTarget(targetX, targetZ);
    }
    this.updateFacing(deltaSeconds);
    if (this.attackAnimationSecondsRemaining > 0) {
      this.modelInstance.play("attack", HUMAN_ANIMATION_FADE_SECONDS);
      this.attackAnimationSecondsRemaining = Math.max(
        0,
        this.attackAnimationSecondsRemaining - deltaSeconds,
      );
    } else {
      this.modelInstance.play(
        this.isMoving() ? HUMAN_RUN_ANIMATION_NAME : HUMAN_IDLE_ANIMATION_NAME,
        HUMAN_ANIMATION_FADE_SECONDS,
      );
    }
    this.modelInstance.update(deltaSeconds);
    this.equipmentAttachments.update(this.entity.getComponent("equipped"), deltaSeconds);
    this.updateHealthBar();
  }

  getObject3D(): THREE.Object3D {
    return this.mesh;
  }

  playAttackAnimation() {
    this.attackAnimationSecondsRemaining = HUMAN_ATTACK_ANIMATION_SECONDS;
    this.modelInstance.play("attack", HUMAN_ANIMATION_FADE_SECONDS);
  }

  onRemove() {
    if (this.healthBar) {
      this.mesh.remove(this.healthBar.object);
    }
    this.equipmentAttachments.dispose();
    this.modelInstance.dispose();
    this.scene.remove(this.mesh);
  }

  private isMoving() {
    return this.segmentElapsedSeconds < SERVER_TICK_SECONDS;
  }

  private updateHealthBar() {
    const health = this.entity.getComponent("health");
    if (health) {
      if (!this.healthBar) {
        this.healthBar = createReactCss2dObject(EntityHealthBar, {
          currentHealth: health.currentHealth,
          maxHealth: health.maxHealth,
        });
        this.healthBar.object.position.set(0.5, HUMAN_HEALTH_BAR_Y, 0.5);
        this.mesh.add(this.healthBar.object);
      } else {
        this.healthBar.updateProps({
          currentHealth: health.currentHealth,
          maxHealth: health.maxHealth,
        });
      }
    } else if (this.healthBar) {
      this.mesh.remove(this.healthBar.object);
      this.healthBar = null;
    }
  }

  private faceSynchronizedTarget(currentX: number, currentZ: number) {
    const facing = this.entity.getComponent("facing");
    if (!facing?.targetEntityId) {
      return;
    }

    const targetPosition = this.resolveEntity(facing.targetEntityId)?.getComponent("position");
    if (!targetPosition) {
      return;
    }

    this.faceDirection(targetPosition.x - currentX, targetPosition.y - currentZ);
  }

  private faceDirection(deltaX: number, deltaZ: number) {
    if (deltaX === 0 && deltaZ === 0) {
      return;
    }
    this.targetRotationY = Math.atan2(deltaX, deltaZ) + HUMAN_MODEL_FORWARD_ROTATION_OFFSET;
  }

  private updateFacing(deltaSeconds: number) {
    const delta = Math.atan2(
      Math.sin(this.targetRotationY - this.visualRoot.rotation.y),
      Math.cos(this.targetRotationY - this.visualRoot.rotation.y),
    );
    const step = Math.min(
      Math.abs(delta),
      HUMAN_ROTATION_SPEED_RADIANS_PER_SECOND * deltaSeconds,
    );
    this.visualRoot.rotation.y += Math.sign(delta) * step;
  }
}
