import * as THREE from "three";
import { box, cone, cylinder, sphere } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createIronSwordModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "ironSword";

  const iron = 0xd2d8dc;
  const ironEdge = 0xf2f4f5;
  const leather = 0x5d3927;
  const brass = 0xb68a3b;

  const pommel = sphere(0.055, 8, 6, brass, {
    roughness: 0.38,
    metalness: 0.65,
  });
  pommel.scale.y = 0.72;
  pommel.position.y = 0.04;
  root.add(pommel);

  const grip = cylinder(0.032, 0.038, 0.15, 8, leather, {
    roughness: 0.92,
  });
  grip.position.y = 0.125;
  root.add(grip);

  const guard = box(0.3, 0.045, 0.055, brass, {
    roughness: 0.38,
    metalness: 0.65,
  });
  guard.position.y = 0.215;
  root.add(guard);

  for (const side of [-1, 1]) {
    const guardTip = sphere(0.035, 7, 5, brass, {
      roughness: 0.38,
      metalness: 0.65,
    });
    guardTip.scale.set(0.75, 1, 0.8);
    guardTip.position.set(side * 0.155, 0.215, 0);
    root.add(guardTip);
  }

  const blade = box(0.09, 0.43, 0.035, iron, {
    roughness: 0.36,
    metalness: 0.48,
  });
  blade.position.y = 0.4525;
  root.add(blade);

  const fuller = box(0.018, 0.42, 0.039, ironEdge, {
    roughness: 0.3,
    metalness: 0.55,
  });
  fuller.position.y = 0.4525;
  root.add(fuller);

  const point = cone(0.064, 0.14, 4, iron, {
    roughness: 0.36,
    metalness: 0.48,
  });
  point.rotation.y = Math.PI / 4;
  point.scale.z = 0.42;
  point.position.y = 0.7375;
  root.add(point);

  return createModelInstance(root);
};
