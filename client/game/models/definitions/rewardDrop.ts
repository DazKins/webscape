import * as THREE from "three";
import { box, sphere, torus } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createRewardDropModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "rewarddrop";

  const bag = sphere(0.28, 12, 8, 0xc58a3a);
  bag.scale.set(1, 0.75, 0.85);
  bag.position.y = 0.28;
  root.add(bag);
  const tie = torus(0.15, 0.025, 7, 12, 0xf0d38a, { metalness: 0.2 });
  tie.rotation.x = Math.PI / 2;
  tie.position.y = 0.52;
  root.add(tie);
  const glint = box(0.12, 0.12, 0.04, 0xfff0a8, { emissive: 0x665500 });
  glint.position.set(-0.14, 0.38, 0.27);
  root.add(glint);

  return createModelInstance(root);
};
