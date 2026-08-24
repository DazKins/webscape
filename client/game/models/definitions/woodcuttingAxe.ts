import * as THREE from "three";
import { box, cylinder } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createWoodcuttingAxeModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "woodcuttingAxe";

  const handle = cylinder(0.027, 0.034, 0.58, 8, 0x7b4b2a, {
    roughness: 0.94,
  });
  handle.position.y = 0.29;
  root.add(handle);

  const grip = cylinder(0.036, 0.039, 0.16, 8, 0x4a2d1d, {
    roughness: 0.9,
  });
  grip.position.y = 0.08;
  root.add(grip);

  const axeHead = new THREE.Group();
  axeHead.name = "axeHead";
  axeHead.position.y = 0.54;
  root.add(axeHead);

  const eye = box(0.13, 0.12, 0.1, 0x59636b, {
    roughness: 0.4,
    metalness: 0.56,
  });
  axeHead.add(eye);

  const blade = box(0.23, 0.2, 0.045, 0xaeb8be, {
    roughness: 0.3,
    metalness: 0.68,
  });
  blade.position.x = -0.15;
  blade.rotation.z = 0.13;
  axeHead.add(blade);

  const edge = box(0.025, 0.22, 0.052, 0xe1e7e9, {
    roughness: 0.24,
    metalness: 0.74,
  });
  edge.position.x = -0.275;
  edge.rotation.z = 0.13;
  axeHead.add(edge);

  const poll = box(0.12, 0.09, 0.085, 0x68737a, {
    roughness: 0.42,
    metalness: 0.52,
  });
  poll.position.x = 0.115;
  axeHead.add(poll);

  return createModelInstance(root);
};
