import * as THREE from "three";
import { box, sphere, cylinder } from "../primitives";
import { animation, createModelInstance, joint } from "../rig";
import type { ModelFactory } from "../types";

export const createDoorModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "door";
  const doorHinge = joint("doorHinge", root, [-0.09, 0, -0.41]);

  const slab = box(0.18, 1.25, 0.82, 0x8a5a34);
  slab.position.set(0.09, 0.625, 0.41);
  doorHinge.add(slab);

  for (const side of [-1, 1]) {
    const surfaceX = 0.09 + side * 0.097;
    for (let index = 0; index < 5; index++) {
      const plank = box(0.02, 1.19, 0.15, index % 2 ? 0x93633d : 0x9e7046);
      plank.position.set(surfaceX, 0.625, 0.09 + index * 0.16);
      doorHinge.add(plank);
    }
    for (const y of [0.24, 1.02]) {
      const strap = box(0.025, 0.065, 0.72, 0x454d50, { metalness: 0.45 });
      strap.position.set(surfaceX + side * 0.016, y, 0.4);
      doorHinge.add(strap);
      for (const z of [0.1, 0.69]) {
        const rivet = sphere(0.019, 6, 4, 0xb29d72, { metalness: 0.5 });
        rivet.position.set(surfaceX + side * 0.03, y, z);
        doorHinge.add(rivet);
      }
    }
    const handle = sphere(0.045, 8, 6, 0xd2b46d, { metalness: 0.35, roughness: 0.45 });
    handle.position.set(0.09 + side * 0.13, 0.62, 0.69);
    doorHinge.add(handle);
  }

  for (const y of [0.24, 1.02]) {
    const hinge = cylinder(0.034, 0.034, 0.14, 8, 0x454d50, { metalness: 0.45 });
    hinge.position.y = y;
    doorHinge.add(hinge);
  }

  const open = animation("open", 0.28, false, (phase) => ({
    doorHinge: { rotation: [0, -easeInOut(phase) * Math.PI / 2, 0] },
  }));
  return createModelInstance(root, { doorHinge }, [open]);
};

function easeInOut(value: number): number {
  return value * value * (3 - 2 * value);
}
