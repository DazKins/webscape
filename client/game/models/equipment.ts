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
  ironSword: {
    equipped: {
      socket: "rightHand",
      position: [0, 0.01, -0.125],
      rotation: [Math.PI / 2 + 0.08, Math.PI / 2, 0],
    },
    dropped: {
      rotation: [0, 0, Math.PI / 2],
    },
  },
  woodcuttingAxe: {
    equipped: {
      socket: "rightHand",
      position: [0, 0.01, -0.13],
      rotation: [Math.PI / 2 + 0.08, Math.PI / 2, 0],
    },
    dropped: {
      rotation: [0, 0, Math.PI / 2],
    },
  },
  fishingRod: {
    equipped: {
      socket: "rightHand",
      position: [0, 0.01, -0.13],
      rotation: [Math.PI / 2 + 0.08, Math.PI / 2, 0],
    },
    dropped: {
      rotation: [0, 0, Math.PI / 2],
    },
  },
  magicStaff: {
    equipped: {
      socket: "rightHand",
      position: [0, -0.48, 0.065],
      rotation: [0, 0, 0],
    },
    dropped: {
      rotation: [0, 0, Math.PI / 2],
    },
  },
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
