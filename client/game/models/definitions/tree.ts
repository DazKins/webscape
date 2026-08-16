import * as THREE from "three";
import { cone, cylinder } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createTreeModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "tree";

  const trunk = cylinder(0.12, 0.16, 0.8, 8, 0x7a4f2a);
  trunk.position.y = 0.4;
  root.add(trunk);

  const lowerLeaves = cone(0.48, 0.78, 9, 0x2f6b3b);
  lowerLeaves.position.y = 0.95;
  root.add(lowerLeaves);
  const upperLeaves = cone(0.36, 0.72, 9, 0x3d7b48);
  upperLeaves.position.y = 1.34;
  root.add(upperLeaves);

  return createModelInstance(root);
};
