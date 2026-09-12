import * as THREE from "three";
import { cylinder, sphere, torus } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createRewardDropModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "rewarddrop";
  const bag = sphere(0.28, 10, 7, 0x9d713e);
  bag.scale.set(1, 0.8, 0.85);
  bag.position.y = 0.225;
  root.add(bag);
  const neck = cylinder(0.1, 0.17, 0.15, 9, 0xbc9159);
  neck.position.y = 0.43;
  root.add(neck);
  const mouth = cylinder(0.13, 0.08, 0.09, 9, 0xbf965e);
  mouth.position.y = 0.52;
  root.add(mouth);
  const tie = torus(0.088, 0.018, 5, 10, 0xead09b);
  tie.rotation.x = Math.PI / 2;
  tie.position.y = 0.48;
  root.add(tie);
  for (const side of [-1, 1]) {
    const loop = torus(0.043, 0.01, 5, 8, 0xead09b);
    loop.position.set(side * 0.038, 0.46, 0.092);
    loop.rotation.z = side * 0.4;
    root.add(loop);
  }
  const coin = cylinder(0.075, 0.075, 0.018, 10, 0xe8bd55, { metalness: 0.5, roughness: 0.4 });
  coin.rotation.x = Math.PI / 2;
  coin.position.set(0, 0.25, 0.238);
  root.add(coin);
  return createModelInstance(root);
};
