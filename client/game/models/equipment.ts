import * as THREE from "three";
import type { ModelName } from "./registry";

export type ModelTransform = {
  position?: readonly [number, number, number];
  rotation?: readonly [number, number, number];
  scale?: readonly [number, number, number];
};

export type EquipmentPresentation = {
  equipped: ModelTransform & {
    socket: string;
  };
  dropped?: ModelTransform;
};

const equipmentPresentations = {
  leatherHelmet: {
    equipped: {
      socket: "headwear",
    },
    dropped: {
      rotation: [0, 0, 0],
    },
  },
} satisfies Partial<Record<ModelName, EquipmentPresentation>>;

export function getEquipmentPresentation(
  renderModel: string,
): EquipmentPresentation | undefined {
  return equipmentPresentations[renderModel as keyof typeof equipmentPresentations];
}

export function applyModelTransform(
  object: THREE.Object3D,
  transform: ModelTransform,
): void {
  object.position.set(...(transform.position ?? [0, 0, 0]));
  object.rotation.set(...(transform.rotation ?? [0, 0, 0]));
  object.scale.set(...(transform.scale ?? [1, 1, 1]));
}
