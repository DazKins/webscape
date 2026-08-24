import type * as THREE from "three";

export type ModelOptions = {
  color?: THREE.ColorRepresentation;
  width?: number;
  height?: number;
  damageStage?: number;
};

export type JointPose = {
  position?: readonly [number, number, number];
  rotation?: readonly [number, number, number];
  scale?: readonly [number, number, number];
};

export type ModelPose = Readonly<Record<string, JointPose>>;

export type ProceduralAnimation = {
  name: string;
  duration: number;
  loop: boolean;
  sample(normalizedTime: number): ModelPose;
};

export type ModelInstance = {
  root: THREE.Group;
  animations: readonly ProceduralAnimation[];
  getSocket(name: string): THREE.Object3D | undefined;
  play(animationName: string, fadeSeconds?: number): void;
  update(deltaSeconds: number): void;
  seek(animationName: string, normalizedTime: number): void;
  dispose(): void;
};

export type ModelFactory = (options?: ModelOptions) => ModelInstance;
