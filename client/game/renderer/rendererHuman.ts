import * as THREE from "three";
import Entity from "../entity/entity";
import { createModel, type HumanAppearance, type ModelInstance } from "../models";
import EntityHealthBar from "../../ui/components/entityHealthBar";
import { createReactCss2dObject, type ReactCss2dObject } from "../../util/reactCss2dObject";
import EntityRenderer, { type TerrainHeightSampler } from "./renderer";
import EquipmentAttachmentController, {
  type EquippedComponent,
} from "./equipmentAttachmentController";
import {
  HUMAN_CHOP_ANIMATION_SECONDS,
  HUMAN_CHOP_CONTACT_SECONDS,
} from "../models/definitions/human";

const POSITION_INTERPOLATION_SECONDS = 0.5;
const SERVER_TICK_SECONDS = 0.5;
const HUMAN_HEALTH_BAR_Y = 1.55;
const HUMAN_IDLE_ANIMATION_NAME = "idle";
const HUMAN_RUN_ANIMATION_NAME = "run";
const HUMAN_ANIMATION_FADE_SECONDS = 0.12;
const HUMAN_ATTACK_ANIMATION_SECONDS = 0.38;
const HUMAN_FISH_WAIT_ANIMATION_SECONDS = 3.2;
const HUMAN_IDLE_ANIMATION_SECONDS = 2;
const HUMAN_RUN_ANIMATION_SECONDS = 0.8;
const HUMAN_MODEL_FORWARD_ROTATION_OFFSET = 0;
const HUMAN_ROTATION_SPEED_RADIANS_PER_SECOND = 10;

export default class RendererHuman extends EntityRenderer {
  mesh: THREE.Group;
  healthBar: ReactCss2dObject<{ currentHealth: number; maxHealth: number }> | null = null;
  private readonly visualRoot: THREE.Group;
  private readonly modelInstance: ModelInstance;
  private readonly equipmentAttachments: EquipmentAttachmentController;
  private readonly hair: THREE.Object3D | undefined;
  private segmentStartX: number;
  private segmentStartZ: number;
  private segmentTargetX: number;
  private segmentTargetZ: number;
  private segmentElapsedSeconds = POSITION_INTERPOLATION_SECONDS;
  private targetRotationY = HUMAN_MODEL_FORWARD_ROTATION_OFFSET;
  private attackAnimationSecondsRemaining = 0;
  private previousFishingPhaseKey: string | null = null;
  private previousWoodcuttingPhaseKey: string | null = null;
  private previousLocomotionPhaseKey: string | null = null;
  private previousCombatPhaseKey: string | null = null;
  private readonly resolveEntity: (entityId: string) => Entity | undefined;
  private readonly getEstimatedServerTick: () => number;

  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler,
    resolveEntity: (entityId: string) => Entity | undefined = () => undefined,
    getEstimatedServerTick: () => number = () => 0,
  ) {
    super(scene, entity, terrainHeightSampler);
    this.resolveEntity = resolveEntity;
    this.getEstimatedServerTick = getEstimatedServerTick;

    const appearance = entity.getComponent("appearance") as HumanAppearance;
    this.mesh = new THREE.Group();
    this.mesh.userData.entityId = entity.getId();
    this.visualRoot = new THREE.Group();
    this.visualRoot.position.set(0.5, 0, 0.5);
    this.mesh.add(this.visualRoot);

    this.modelInstance = createModel("human", { appearance });
    this.hair = this.modelInstance.root.getObjectByName("hair");
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
    const progress = Math.min(this.segmentElapsedSeconds / POSITION_INTERPOLATION_SECONDS, 1);
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
    const isFishing = this.updateFishingAnimation();
    if (isFishing) {
      this.previousWoodcuttingPhaseKey = null;
      this.previousLocomotionPhaseKey = null;
    } else {
      const isWoodcutting = this.updateWoodcuttingAnimation();
      if (isWoodcutting) {
        this.previousLocomotionPhaseKey = null;
      } else if (this.updateCombatAnimation()) {
        this.previousLocomotionPhaseKey = null;
      } else if (this.attackAnimationSecondsRemaining > 0) {
        this.previousLocomotionPhaseKey = null;
        this.modelInstance.play("attack", HUMAN_ANIMATION_FADE_SECONDS);
        this.attackAnimationSecondsRemaining = Math.max(
          0,
          this.attackAnimationSecondsRemaining - deltaSeconds,
        );
      } else {
        this.updateLocomotionAnimation();
      }
    }
    this.modelInstance.update(deltaSeconds);
    const equipped = this.entity.getComponent("equipped") as EquippedComponent | undefined;
    if (this.hair) {
      this.hair.visible = !equipped?.slots?.head;
    }
    this.equipmentAttachments.update(equipped, deltaSeconds);
    this.updateHealthBar();
  }

  getObject3D(): THREE.Object3D {
    return this.mesh;
  }

  playAttackAnimation() {
    this.attackAnimationSecondsRemaining = HUMAN_ATTACK_ANIMATION_SECONDS;
    this.modelInstance.play("attack", HUMAN_ANIMATION_FADE_SECONDS);
  }

  getProjectileOrigin(projectileType: string): THREE.Vector3 | null {
    const weapon = this.equipmentAttachments.getAttachmentObject("weapon");
    if (projectileType === "magicBolt") {
      return weapon?.getObjectByName("magicStaffCrown")
        ?.getWorldPosition(new THREE.Vector3()) ?? null;
    }
    if (projectileType === "arrow") {
      return weapon?.getObjectByName("bowArrowRest")
        ?.getWorldPosition(new THREE.Vector3()) ?? null;
    }
    return null;
  }

  onRemove() {
    if (this.healthBar) {
      this.mesh.remove(this.healthBar.object);
      this.healthBar.dispose();
    }
    this.equipmentAttachments.dispose();
    this.modelInstance.dispose();
    this.scene.remove(this.mesh);
  }

  private isMoving() {
    return this.segmentElapsedSeconds < POSITION_INTERPOLATION_SECONDS;
  }

  private updateLocomotionAnimation() {
    const equipped = this.entity.getComponent("equipped") as EquippedComponent | undefined;
    const weapon = equipped?.slots?.weapon?.renderModel;
    const idleName = weapon === "woodenBow" ? "bowIdle"
      : weapon === "magicStaff" ? "staffIdle" : HUMAN_IDLE_ANIMATION_NAME;
    const runName = weapon === "woodenBow" ? "bowRun"
      : weapon === "magicStaff" ? "staffRun" : HUMAN_RUN_ANIMATION_NAME;
    const locomotion = this.entity.getComponent("locomotion");
    const phase = this.locomotionPhase(locomotion);
    const phaseStartedTick = this.locomotionPhaseStartedTick(locomotion);
    if (phase === null || phaseStartedTick === null) {
      this.previousLocomotionPhaseKey = null;
      this.modelInstance.play(
        this.isMoving() ? runName : idleName,
        HUMAN_ANIMATION_FADE_SECONDS,
      );
      return;
    }

    const animationName = phase === "moving"
      ? runName
      : idleName;
    const animationDuration = phase === "moving"
      ? HUMAN_RUN_ANIMATION_SECONDS
      : HUMAN_IDLE_ANIMATION_SECONDS;
    const phaseKey = `${animationName}:${phase}:${phaseStartedTick}`;
    if (phaseKey !== this.previousLocomotionPhaseKey) {
      const phaseAgeTicks = Math.max(0, this.getEstimatedServerTick() - phaseStartedTick);
      const elapsedSeconds = phaseAgeTicks * SERVER_TICK_SECONDS;
      this.modelInstance.playAt(
        animationName,
        (elapsedSeconds % animationDuration) / animationDuration,
        HUMAN_ANIMATION_FADE_SECONDS,
      );
      this.previousLocomotionPhaseKey = phaseKey;
    } else {
      this.modelInstance.play(animationName, HUMAN_ANIMATION_FADE_SECONDS);
    }
  }

  private updateCombatAnimation(): boolean {
    const combat = this.entity.getComponent("combatstate");
    if (
      typeof combat !== "object" || combat === null ||
      (combat.attackMethod !== "magic" && combat.attackMethod !== "ranged") ||
      (combat.phase !== "casting" && combat.phase !== "recovering") ||
      typeof combat.phaseStartedTick !== "number"
    ) {
      this.previousCombatPhaseKey = null;
      return false;
    }

    const phaseAgeTicks = Math.max(
      0,
      this.getEstimatedServerTick() - combat.phaseStartedTick,
    );
    const windUpTicks = typeof combat.windUpTicks === "number"
      ? Math.max(1, combat.windUpTicks)
      : 2;
    const normalizedTime = combat.phase === "casting"
      ? Math.min(2 / 3, (phaseAgeTicks / windUpTicks) * (2 / 3))
      : Math.min(1, 2 / 3 + phaseAgeTicks / 3);
    const phaseKey = `${combat.phase}:${combat.phaseStartedTick}`;
    const animationName = combat.attackMethod === "ranged" ? "shoot" : "cast";
    if (phaseKey !== this.previousCombatPhaseKey) {
      this.modelInstance.playAt(animationName, normalizedTime, HUMAN_ANIMATION_FADE_SECONDS);
      this.previousCombatPhaseKey = phaseKey;
    } else {
      this.modelInstance.playAt(animationName, normalizedTime);
    }
    return true;
  }

  private locomotionPhase(locomotion: unknown): "idle" | "moving" | null {
    if (typeof locomotion !== "object" || locomotion === null || !("phase" in locomotion)) {
      return null;
    }
    return locomotion.phase === "idle" || locomotion.phase === "moving"
      ? locomotion.phase
      : null;
  }

  private locomotionPhaseStartedTick(locomotion: unknown): number | null {
    if (
      typeof locomotion !== "object" ||
      locomotion === null ||
      !("phaseStartedTick" in locomotion)
    ) {
      return null;
    }
    return typeof locomotion.phaseStartedTick === "number" &&
      Number.isFinite(locomotion.phaseStartedTick)
      ? locomotion.phaseStartedTick
      : null;
  }

  private updateFishingAnimation(): boolean {
    const fishing = this.entity.getComponent("fishing");
    const phase = this.fishingPhase(fishing);
    const phaseStartedTick = this.fishingPhaseStartedTick(fishing);
    if (phase === null || phaseStartedTick === null) {
      this.previousFishingPhaseKey = null;
      return false;
    }

    const estimatedTick = this.getEstimatedServerTick();
    const phaseAgeTicks = Math.max(0, estimatedTick - phaseStartedTick);
    const phaseKey = `${phase}:${phaseStartedTick}`;
    if (phaseKey !== this.previousFishingPhaseKey) {
      if (phase === "waiting") {
        const elapsedSeconds = phaseAgeTicks * SERVER_TICK_SECONDS;
        this.modelInstance.seek(
          "fishWait",
          (elapsedSeconds % HUMAN_FISH_WAIT_ANIMATION_SECONDS) /
            HUMAN_FISH_WAIT_ANIMATION_SECONDS,
        );
      } else {
        this.modelInstance.seek("fishAction", Math.min(phaseAgeTicks, 1));
      }
      this.previousFishingPhaseKey = phaseKey;
    }

    if (phase === "waiting" || phaseAgeTicks >= 1) {
      this.modelInstance.play("fishWait", HUMAN_ANIMATION_FADE_SECONDS);
    } else {
      this.modelInstance.play("fishAction", HUMAN_ANIMATION_FADE_SECONDS);
    }
    return true;
  }

  private updateWoodcuttingAnimation(): boolean {
    const woodcutting = this.entity.getComponent("woodcutting");
    const phase = this.woodcuttingPhase(woodcutting);
    const phaseStartedTick = this.woodcuttingPhaseStartedTick(woodcutting);
    if (phase === null || phaseStartedTick === null) {
      this.previousWoodcuttingPhaseKey = null;
      return false;
    }

    const phaseAgeTicks = Math.max(0, this.getEstimatedServerTick() - phaseStartedTick);
    const phaseAgeSeconds = phaseAgeTicks * SERVER_TICK_SECONDS;
    const contactPhase = HUMAN_CHOP_CONTACT_SECONDS / HUMAN_CHOP_ANIMATION_SECONDS;
    const phaseKey = `${phase}:${phaseStartedTick}`;
    if (phaseKey !== this.previousWoodcuttingPhaseKey) {
      if (phase === "swinging") {
        this.modelInstance.playAt(
          "chop",
          Math.min(phaseAgeSeconds / HUMAN_CHOP_ANIMATION_SECONDS, contactPhase),
          HUMAN_ANIMATION_FADE_SECONDS,
        );
      } else {
        const elapsedSeconds = HUMAN_CHOP_CONTACT_SECONDS + phaseAgeSeconds;
        if (elapsedSeconds < HUMAN_CHOP_ANIMATION_SECONDS) {
          this.modelInstance.playAt(
            "chop",
            elapsedSeconds / HUMAN_CHOP_ANIMATION_SECONDS,
            HUMAN_ANIMATION_FADE_SECONDS,
          );
        } else {
          this.modelInstance.play(HUMAN_IDLE_ANIMATION_NAME, HUMAN_ANIMATION_FADE_SECONDS);
        }
      }
      this.previousWoodcuttingPhaseKey = phaseKey;
    } else if (phase === "swinging" && phaseAgeSeconds < HUMAN_CHOP_CONTACT_SECONDS) {
      this.modelInstance.play("chop", HUMAN_ANIMATION_FADE_SECONDS);
    } else if (phase === "swinging") {
      this.modelInstance.playAt("chop", contactPhase);
    } else if (
      HUMAN_CHOP_CONTACT_SECONDS + phaseAgeSeconds <
      HUMAN_CHOP_ANIMATION_SECONDS
    ) {
      this.modelInstance.play("chop", HUMAN_ANIMATION_FADE_SECONDS);
    } else {
      this.modelInstance.play(HUMAN_IDLE_ANIMATION_NAME, HUMAN_ANIMATION_FADE_SECONDS);
    }
    return true;
  }

  private woodcuttingPhase(woodcutting: unknown): "swinging" | "recovering" | null {
    if (
      typeof woodcutting !== "object" ||
      woodcutting === null ||
      !("phase" in woodcutting)
    ) {
      return null;
    }
    return woodcutting.phase === "swinging" || woodcutting.phase === "recovering"
      ? woodcutting.phase
      : null;
  }

  private woodcuttingPhaseStartedTick(woodcutting: unknown): number | null {
    if (
      typeof woodcutting !== "object" ||
      woodcutting === null ||
      !("phaseStartedTick" in woodcutting)
    ) {
      return null;
    }
    return typeof woodcutting.phaseStartedTick === "number" &&
      Number.isFinite(woodcutting.phaseStartedTick)
      ? woodcutting.phaseStartedTick
      : null;
  }

  private fishingPhase(fishing: unknown): "casting" | "waiting" | "reeling" | null {
    if (typeof fishing !== "object" || fishing === null || !("phase" in fishing)) {
      return null;
    }
    return fishing.phase === "casting" || fishing.phase === "waiting" || fishing.phase === "reeling"
      ? fishing.phase
      : null;
  }

  private fishingPhaseStartedTick(fishing: unknown): number | null {
    if (typeof fishing !== "object" || fishing === null || !("phaseStartedTick" in fishing)) {
      return null;
    }
    return typeof fishing.phaseStartedTick === "number" &&
      Number.isFinite(fishing.phaseStartedTick)
      ? fishing.phaseStartedTick
      : null;
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
      this.healthBar.dispose();
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
