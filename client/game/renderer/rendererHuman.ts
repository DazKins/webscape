import Entity from "../entity/entity";
import EntityRenderer, { type TerrainHeightSampler } from "./renderer";
import * as THREE from "three";
import { GLTFLoader, type GLTF } from "three/examples/jsm/loaders/GLTFLoader.js";
import { clone as cloneSkeleton } from "three/examples/jsm/utils/SkeletonUtils.js";
import { createReactCss2dObject, ReactCss2dObject } from "../../util/reactCss2dObject";
import EntityHealthBar from "../../ui/components/entityHealthBar";

const SERVER_TICK_SECONDS = 0.52;
const HUMAN_MODEL_URL = "/models/man.glb";
const HUMAN_MODEL_HEIGHT = 1.35;
const HUMAN_HEALTH_BAR_Y = 1.55;
const HUMAN_IDLE_ANIMATION_NAME = "idle";
const HUMAN_WALK_ANIMATION_NAME = "run";
const HUMAN_ANIMATION_FADE_SECONDS = 0.12;
const HUMAN_MODEL_FORWARD_ROTATION_OFFSET = 0;
const HUMAN_ROTATION_SPEED_RADIANS_PER_SECOND = 10;

const humanModelLoader = new GLTFLoader();
let humanModelPromise: Promise<GLTF> | null = null;

function loadHumanModel() {
  if (!humanModelPromise) {
    humanModelPromise = new Promise((resolve, reject) => {
      humanModelLoader.load(HUMAN_MODEL_URL, resolve, undefined, reject);
    });
  }

  return humanModelPromise;
}

function createFallbackHuman(color: string) {
  const fallback = new THREE.Group();

  const bodyGeometry = new THREE.CylinderGeometry(0.3, 0.3, 0.8, 16);
  const bodyMaterial = new THREE.MeshPhongMaterial({
    color: color,
  });
  const body = new THREE.Mesh(bodyGeometry, bodyMaterial);
  body.position.x = 0;
  body.position.y = 0.4;
  body.position.z = 0;
  fallback.add(body);

  const headGeometry = new THREE.SphereGeometry(0.25, 16, 16);
  const headMaterial = new THREE.MeshPhongMaterial({
    color: 0xffe0bd,
  });
  const head = new THREE.Mesh(headGeometry, headMaterial);
  head.position.x = 0;
  head.position.y = 0.9;
  head.position.z = 0;
  fallback.add(head);

  return fallback;
}

function applyStaticTPose(model: THREE.Object3D, animations: THREE.AnimationClip[]) {
  const tPoseClip = animations.find((clip) => clip.name.toLowerCase().includes("tpose"));
  if (!tPoseClip) {
    return;
  }

  const mixer = new THREE.AnimationMixer(model);
  const action = mixer.clipAction(tPoseClip);
  action.play();
  mixer.setTime(0);
  mixer.update(0);
}

function findAnimation(animations: THREE.AnimationClip[], name: string) {
  return animations.find((clip) => clip.name.toLowerCase() === name);
}

function normalizeHumanModel(model: THREE.Object3D) {
  model.updateMatrixWorld(true);
  let box = new THREE.Box3().setFromObject(model);
  const size = box.getSize(new THREE.Vector3());
  if (size.y <= 0) {
    return;
  }

  const scale = HUMAN_MODEL_HEIGHT / size.y;
  model.scale.setScalar(scale);
  model.updateMatrixWorld(true);

  box = new THREE.Box3().setFromObject(model);
  const center = box.getCenter(new THREE.Vector3());
  model.position.x += -center.x;
  model.position.y += -box.min.y;
  model.position.z += -center.z;
}

export default class RendererHuman extends EntityRenderer {
  mesh: THREE.Group;
  healthBar: ReactCss2dObject<{ currentHealth: number; maxHealth: number }> | null = null;
  private visualRoot: THREE.Group;
  private fallbackModel: THREE.Group;
  private model: THREE.Object3D | null = null;
  private mixer: THREE.AnimationMixer | null = null;
  private idleAction: THREE.AnimationAction | null = null;
  private walkAction: THREE.AnimationAction | null = null;
  private activeAction: THREE.AnimationAction | null = null;
  private isRemoved = false;
  private segmentStartX: number;
  private segmentStartZ: number;
  private segmentTargetX: number;
  private segmentTargetZ: number;
  private segmentElapsedSeconds = SERVER_TICK_SECONDS;
  private targetRotationY = HUMAN_MODEL_FORWARD_ROTATION_OFFSET;

  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler
  ) {
    super(scene, entity, terrainHeightSampler);

    let metadataComponent = entity.getComponent("metadata");
    if (!metadataComponent) {
      metadataComponent = {}
    }
    const color = metadataComponent.color ?? "#00ff00";

    this.mesh = new THREE.Group();
    this.mesh.userData.entityId = entity.getId();

    this.visualRoot = new THREE.Group();
    this.visualRoot.position.set(0.5, 0, 0.5);
    this.mesh.add(this.visualRoot);

    this.fallbackModel = createFallbackHuman(color);
    this.visualRoot.add(this.fallbackModel);
    void this.loadModel();

    const positionComponent = this.entity.getComponent("position");
    this.mesh.position.set(
      positionComponent.x,
      this.sampleVisualHeight(positionComponent.x + 0.5, positionComponent.y + 0.5),
      positionComponent.y
    );
    this.segmentStartX = positionComponent.x;
    this.segmentStartZ = positionComponent.y;
    this.segmentTargetX = positionComponent.x;
    this.segmentTargetZ = positionComponent.y;

    this.scene.add(this.mesh);
  }

  private async loadModel() {
    try {
      const humanModel = await loadHumanModel();
      if (this.isRemoved) {
        return;
      }

      const model = cloneSkeleton(humanModel.scene);
      this.setupAnimations(model, humanModel.animations);
      normalizeHumanModel(model);

      if (this.isRemoved) {
        return;
      }

      this.visualRoot.remove(this.fallbackModel);
      this.model = model;
      this.visualRoot.add(model);
    } catch (error) {
      console.error("failed to load human model:", error);
    }
  }

  private setupAnimations(model: THREE.Object3D, animations: THREE.AnimationClip[]) {
    const idleClip = findAnimation(animations, HUMAN_IDLE_ANIMATION_NAME);
    const walkClip = findAnimation(animations, HUMAN_WALK_ANIMATION_NAME);

    if (!idleClip && !walkClip) {
      applyStaticTPose(model, animations);
      return;
    }

    this.mixer = new THREE.AnimationMixer(model);
    this.idleAction = idleClip ? this.mixer.clipAction(idleClip) : null;
    this.walkAction = walkClip ? this.mixer.clipAction(walkClip) : null;
    this.updateMovementAnimation(this.isMoving());
    this.mixer.update(0);
  }

  private isMoving() {
    return this.segmentElapsedSeconds < SERVER_TICK_SECONDS;
  }

  private updateMovementAnimation(isMoving: boolean) {
    const nextAction = isMoving ? this.walkAction : this.idleAction;
    if (this.activeAction === nextAction) {
      return;
    }

    const previousAction = this.activeAction;
    this.activeAction = nextAction;

    if (!nextAction) {
      previousAction?.fadeOut(HUMAN_ANIMATION_FADE_SECONDS);
      return;
    }

    nextAction.reset();
    nextAction.enabled = true;
    nextAction.setEffectiveTimeScale(1);
    nextAction.setEffectiveWeight(1);
    nextAction.play();

    if (previousAction) {
      previousAction.crossFadeTo(nextAction, HUMAN_ANIMATION_FADE_SECONDS, false);
    } else {
      nextAction.fadeIn(HUMAN_ANIMATION_FADE_SECONDS);
    }
  }

  update(deltaSeconds: number) {
    const positionComponent = this.entity.getComponent("position");
    if (!positionComponent) {
      return;
    }

    const targetX = positionComponent.x;
    const targetZ = positionComponent.y;

    if (targetX !== this.segmentTargetX || targetZ !== this.segmentTargetZ) {
      this.segmentStartX = this.mesh.position.x;
      this.segmentStartZ = this.mesh.position.z;
      this.segmentTargetX = targetX;
      this.segmentTargetZ = targetZ;
      this.segmentElapsedSeconds = 0;
      this.faceTravelDirection(targetX - this.segmentStartX, targetZ - this.segmentStartZ);
    }

    this.segmentElapsedSeconds += deltaSeconds;
    const progress = Math.min(this.segmentElapsedSeconds / SERVER_TICK_SECONDS, 1);
    const renderedX =
      this.segmentStartX + (this.segmentTargetX - this.segmentStartX) * progress;
    const renderedZ =
      this.segmentStartZ + (this.segmentTargetZ - this.segmentStartZ) * progress;
    this.mesh.position.set(
      renderedX,
      this.sampleVisualHeight(renderedX + 0.5, renderedZ + 0.5),
      renderedZ,
    );
    this.updateFacing(deltaSeconds);
    this.updateMovementAnimation(this.isMoving());
    this.mixer?.update(deltaSeconds);

    // Update health bar if health component exists
    const healthComponent = this.entity.getComponent("health");
    if (healthComponent) {
      if (!this.healthBar) {
        // Create health bar if it doesn't exist
        this.healthBar = createReactCss2dObject(EntityHealthBar, {
          currentHealth: healthComponent.currentHealth,
          maxHealth: healthComponent.maxHealth,
        });
        this.healthBar.object.position.x = 0.5;
        this.healthBar.object.position.y = HUMAN_HEALTH_BAR_Y;
        this.healthBar.object.position.z = 0.5;
        this.mesh.add(this.healthBar.object);
      } else {
        // Update existing health bar props
        this.healthBar.updateProps({
          currentHealth: healthComponent.currentHealth,
          maxHealth: healthComponent.maxHealth,
        });
      }
    } else if (this.healthBar) {
      // Remove health bar if health component is removed
      this.mesh.remove(this.healthBar.object);
      this.healthBar = null;
    }
  }

  getObject3D(): THREE.Object3D {
    return this.mesh;
  }

  onRemove() {
    this.isRemoved = true;
    if (this.healthBar) {
      this.mesh.remove(this.healthBar.object);
    }
    if (this.model) {
      this.visualRoot.remove(this.model);
    }
    this.mixer?.stopAllAction();
    this.scene.remove(this.mesh);
  }

  private faceTravelDirection(deltaX: number, deltaZ: number) {
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
    const step = Math.min(Math.abs(delta), HUMAN_ROTATION_SPEED_RADIANS_PER_SECOND * deltaSeconds);
    this.visualRoot.rotation.y += Math.sign(delta) * step;
  }
}
