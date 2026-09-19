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
    // Named model groups that follow separate host joints, in joint-local space.
    parts?: Readonly<Record<string, string>>;
    hideParts?: readonly string[];
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
      rotation: [Math.PI / 2, 0, Math.PI / 2],
    },
  },
  woodcuttingAxe: {
    equipped: {
      socket: "rightHand",
      position: [0, 0.01, -0.13],
      rotation: [Math.PI / 2 + 0.08, Math.PI / 2, 0],
    },
    dropped: {
      rotation: [Math.PI / 2, 0, Math.PI / 2],
    },
  },
  fishingRod: {
    equipped: {
      socket: "rightHand",
      position: [0, 0.01, -0.13],
      rotation: [Math.PI / 2 + 0.08, Math.PI / 2, 0],
    },
    dropped: {
      rotation: [Math.PI / 2, 0, Math.PI / 2],
    },
  },
  magicStaff: {
    equipped: {
      socket: "rightHand",
      position: [0, -0.78, 0],
      rotation: [0, 0, 0],
    },
    dropped: {
      rotation: [Math.PI / 2, 0, Math.PI / 2],
    },
  },
  woodenBow: {
    equipped: {
      socket: "leftHand",
      position: [0, 0, 0],
      rotation: [0, -Math.PI / 2, 0],
      scale: [0.85, 0.85, 0.85],
    },
    dropped: {
      rotation: [Math.PI / 2, 0, 0],
    },
  },
  leatherHelmet: {
    equipped: {
      socket: "headwear",
      hideParts: ["hair"],
    },
    dropped: {
      rotation: [0, 0, 0],
    },
  },
  chainmailChestplate: {
    equipped: {
      socket: "torso",
      parts: { mailSkirt: "hips", leftMailSleeve: "leftShoulder", rightMailSleeve: "rightShoulder" },
      hideParts: ["tunic", "tunicSkirt", "leftSleeve", "rightSleeve"],
    },
  },
  ironLeggings: {
    equipped: {
      socket: "hips",
      parts: { leftCuisses: "leftHip", rightCuisses: "rightHip", leftGreave: "leftKnee", rightGreave: "rightKnee" },
      hideParts: ["leftTrousers", "rightTrousers", "leftShin", "rightShin", "leftCuff", "rightCuff"],
    },
  },
  leatherBoots: {
    equipped: {
      socket: "hips",
      parts: { leftBootShaft: "leftKnee", rightBootShaft: "rightKnee", leftBootFoot: "leftAnkle", rightBootFoot: "rightAnkle" },
      hideParts: ["leftShin", "rightShin", "leftCuff", "rightCuff", "leftFoot", "rightFoot"],
    },
  },
  woodenShield: {
    equipped: {
      socket: "leftElbow",
      position: [0.09, -0.14, 0],
      rotation: [0, 0, -Math.PI / 2],
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
