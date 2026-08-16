import * as THREE from "three";
import type {
  JointPose,
  ModelInstance,
  ModelPose,
  ProceduralAnimation,
} from "./types";

export type JointMap = Readonly<Record<string, THREE.Group>>;
export type SocketMap = Readonly<Record<string, THREE.Object3D>>;

type Transform = {
  position: THREE.Vector3;
  quaternion: THREE.Quaternion;
  scale: THREE.Vector3;
};

type Playback = {
  animation: ProceduralAnimation;
  time: number;
};

export function joint(
  name: string,
  parent: THREE.Object3D,
  position: readonly [number, number, number] = [0, 0, 0],
): THREE.Group {
  const result = new THREE.Group();
  result.name = name;
  result.position.set(...position);
  parent.add(result);
  return result;
}

export function animation(
  name: string,
  duration: number,
  loop: boolean,
  sample: (normalizedTime: number) => ModelPose,
): ProceduralAnimation {
  return { name, duration, loop, sample };
}

export function createModelInstance(
  root: THREE.Group,
  joints: JointMap = {},
  animations: readonly ProceduralAnimation[] = [],
  sockets: SocketMap = {},
): ModelInstance {
  return new RiggedModelInstance(root, joints, animations, sockets);
}

class RiggedModelInstance implements ModelInstance {
  readonly animations: readonly ProceduralAnimation[];
  private readonly bindPose: Readonly<Record<string, Transform>>;
  private readonly geometries: ReadonlySet<THREE.BufferGeometry>;
  private readonly materials: ReadonlySet<THREE.Material>;
  private active: Playback | null = null;
  private previous: Playback | null = null;
  private fadeSeconds = 0;
  private fadeElapsed = 0;
  private isDisposed = false;

  constructor(
    readonly root: THREE.Group,
    private readonly joints: JointMap,
    animations: readonly ProceduralAnimation[],
    private readonly sockets: SocketMap,
  ) {
    this.animations = animations;
    this.bindPose = captureBindPose(joints);
    const resources = captureResources(root);
    this.geometries = resources.geometries;
    this.materials = resources.materials;
  }

  getSocket(name: string): THREE.Object3D | undefined {
    this.assertAvailable();
    return this.sockets[name];
  }

  play(animationName: string, fadeSeconds = 0): void {
    this.assertAvailable();
    if (this.active?.animation.name === animationName) {
      return;
    }

    const animationDefinition = this.getAnimation(animationName);
    this.previous = this.active;
    this.active = { animation: animationDefinition, time: 0 };
    this.fadeSeconds = Math.max(0, fadeSeconds);
    this.fadeElapsed = 0;
    this.applyCurrentPose();
  }

  update(deltaSeconds: number): void {
    this.assertAvailable();
    const elapsed = Math.max(0, deltaSeconds);
    if (this.active) {
      this.active.time += elapsed;
    }
    if (this.previous) {
      this.previous.time += elapsed;
      this.fadeElapsed += elapsed;
    }
    this.applyCurrentPose();
  }

  seek(animationName: string, normalizedTime: number): void {
    this.assertAvailable();
    const animationDefinition = this.getAnimation(animationName);
    const phase = THREE.MathUtils.clamp(normalizedTime, 0, 1);
    this.active = {
      animation: animationDefinition,
      time: phase * animationDefinition.duration,
    };
    this.previous = null;
    this.fadeSeconds = 0;
    this.fadeElapsed = 0;
    this.applyPose(animationDefinition.sample(phase));
  }

  dispose(): void {
    if (this.isDisposed) {
      return;
    }

    for (const geometry of this.geometries) {
      geometry.dispose();
    }
    for (const material of this.materials) {
      material.dispose();
    }

    this.root.removeFromParent();
    this.isDisposed = true;
  }

  private getAnimation(name: string): ProceduralAnimation {
    const result = this.animations.find((candidate) => candidate.name === name);
    if (!result) {
      throw new Error(`model does not define animation: ${name}`);
    }
    return result;
  }

  private applyCurrentPose(): void {
    if (!this.active) {
      this.applyPose({});
      return;
    }

    const activePose = this.active.animation.sample(playbackPhase(this.active));
    if (!this.previous) {
      this.applyPose(activePose);
      return;
    }

    const previousPose = this.previous.animation.sample(playbackPhase(this.previous));
    const fade = this.fadeSeconds === 0 ? 1 : Math.min(1, this.fadeElapsed / this.fadeSeconds);
    this.applyBlendedPose(previousPose, activePose, fade);
    if (fade >= 1) {
      this.previous = null;
    }
  }

  private applyPose(pose: ModelPose): void {
    for (const [name, object] of Object.entries(this.joints)) {
      applyTransform(object, poseTransform(this.bindPose[name], pose[name]));
    }
  }

  private applyBlendedPose(from: ModelPose, to: ModelPose, alpha: number): void {
    for (const [name, object] of Object.entries(this.joints)) {
      const fromTransform = poseTransform(this.bindPose[name], from[name]);
      const toTransform = poseTransform(this.bindPose[name], to[name]);
      object.position.lerpVectors(fromTransform.position, toTransform.position, alpha);
      object.quaternion.slerpQuaternions(fromTransform.quaternion, toTransform.quaternion, alpha);
      object.scale.lerpVectors(fromTransform.scale, toTransform.scale, alpha);
    }
  }

  private assertAvailable(): void {
    if (this.isDisposed) {
      throw new Error("cannot use a disposed model instance");
    }
  }
}

function captureResources(root: THREE.Object3D): {
  geometries: ReadonlySet<THREE.BufferGeometry>;
  materials: ReadonlySet<THREE.Material>;
} {
  const geometries = new Set<THREE.BufferGeometry>();
  const materials = new Set<THREE.Material>();
  root.traverse((object) => {
    if (!(object instanceof THREE.Mesh)) {
      return;
    }
    geometries.add(object.geometry);
    const objectMaterials = Array.isArray(object.material) ? object.material : [object.material];
    for (const material of objectMaterials) {
      materials.add(material);
    }
  });
  return { geometries, materials };
}

function captureBindPose(joints: JointMap): Readonly<Record<string, Transform>> {
  return Object.fromEntries(
    Object.entries(joints).map(([name, object]) => [
      name,
      {
        position: object.position.clone(),
        quaternion: object.quaternion.clone(),
        scale: object.scale.clone(),
      },
    ]),
  );
}

function poseTransform(bind: Transform, pose: JointPose | undefined): Transform {
  const position = bind.position.clone();
  if (pose?.position) {
    position.add(new THREE.Vector3(...pose.position));
  }

  const quaternion = bind.quaternion.clone();
  if (pose?.rotation) {
    quaternion.multiply(new THREE.Quaternion().setFromEuler(new THREE.Euler(...pose.rotation)));
  }

  const scale = bind.scale.clone();
  if (pose?.scale) {
    scale.multiply(new THREE.Vector3(...pose.scale));
  }
  return { position, quaternion, scale };
}

function applyTransform(object: THREE.Object3D, transform: Transform): void {
  object.position.copy(transform.position);
  object.quaternion.copy(transform.quaternion);
  object.scale.copy(transform.scale);
}

function playbackPhase(playback: Playback): number {
  const { animation: definition, time } = playback;
  if (definition.duration <= 0) {
    return 0;
  }
  if (!definition.loop) {
    return Math.min(1, time / definition.duration);
  }
  return (time % definition.duration) / definition.duration;
}
