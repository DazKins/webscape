import * as THREE from "three";
import { box, cone } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createBuildingModel: ModelFactory = (options = {}) => {
  const width = Math.max(0.8, options.width ?? 2);
  const depth = Math.max(0.8, options.height ?? 2);
  const root = new THREE.Group();
  root.name = "building";

  const body = box(width * 0.9, 1.2, depth * 0.9, 0xb8ab88);
  body.position.y = 0.6;
  root.add(body);

  const roof = cone(Math.max(width, depth) * 0.7, 0.55, 4, 0x7f4231);
  roof.rotation.y = Math.PI / 4;
  roof.position.y = 1.475;
  root.add(roof);

  const doorway = box(Math.min(0.34, width * 0.3), 0.68, 0.035, 0x57402e);
  doorway.position.set(0, 0.34, depth * 0.455);
  root.add(doorway);

  return createModelInstance(root);
};
