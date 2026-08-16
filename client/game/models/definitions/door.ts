import * as THREE from "three";
import { box, sphere } from "../primitives";
import { animation, createModelInstance, joint } from "../rig";
import type { ModelFactory } from "../types";

export const createDoorModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "door";
  const doorHinge = joint("doorHinge", root, [-0.09, 0, -0.41]);

  const slab = box(0.18, 1.25, 0.82, 0x8a5a34);
  slab.position.set(0.09, 0.625, 0.41);
  doorHinge.add(slab);

  const inset = box(0.022, 0.86, 0.55, 0x6f4326);
  inset.position.set(0.101, 0.64, 0.41);
  doorHinge.add(inset);

  const handle = sphere(0.05, 8, 6, 0xd2b46d, { metalness: 0.35, roughness: 0.45 });
  handle.position.set(0.21, 0.62, 0.69);
  doorHinge.add(handle);

  const open = animation("open", 0.28, false, (phase) => ({
    doorHinge: { rotation: [0, -easeInOut(phase) * Math.PI / 2, 0] },
  }));
  return createModelInstance(root, { doorHinge }, [open]);
};

function easeInOut(value: number): number {
  return value * value * (3 - 2 * value);
}
