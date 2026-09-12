import * as THREE from "three";
import { box, taperedBox } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createBuildingModel: ModelFactory = (options = {}) => {
  const width = Math.max(0.8, options.width ?? 2);
  const depth = Math.max(0.8, options.height ?? 2);
  const root = new THREE.Group();
  root.name = "building";

  const body = box(width * 0.9, 1.2, depth * 0.9, 0xd2bf98);
  body.position.y = 0.6;
  root.add(body);

  // A rectangular hip roof follows both footprint dimensions.
  const roof = taperedBox(width * 0.22, depth * 0.22, width, depth, 0.55, 0x864a3b);
  roof.position.y = 1.475;
  root.add(roof);

  const foundation = box(width * 0.94, 0.14, depth * 0.94, 0x77746b);
  foundation.position.y = 0.07;
  root.add(foundation);
  for (const y of [0.2, 1.16]) {
    const beam = box(width * 0.93, 0.08, depth * 0.93, 0x614532);
    beam.position.y = y;
    root.add(beam);
  }
  for (const x of [-1, 1]) {
    for (const z of [-1, 1]) {
      const post = box(0.075, 1.08, 0.075, 0x614532);
      post.position.set(x * width * 0.44, 0.64, z * depth * 0.44);
      root.add(post);
    }
    const windowFrame = box(0.035, 0.46, depth * 0.32, 0x614532);
    windowFrame.position.set(x * width * 0.456, 0.76, 0);
    root.add(windowFrame);
    const glass = box(0.04, 0.35, depth * 0.25, 0x526c70, { roughness: 0.35 });
    glass.position.copy(windowFrame.position);
    root.add(glass);
    const mullion = box(0.045, 0.37, 0.035, 0x614532);
    mullion.position.copy(windowFrame.position);
    root.add(mullion);
    const sill = box(0.09, 0.045, depth * 0.36, 0x8d7050);
    sill.position.set(x * width * 0.456, 0.52, 0);
    root.add(sill);
  }

  const doorway = box(Math.min(0.34, width * 0.3), 0.68, 0.035, 0x57402e);
  doorway.position.set(0, 0.34, depth * 0.455);
  root.add(doorway);
  const doorFrame = box(Math.min(0.44, width * 0.4), 0.075, 0.07, 0x614532);
  doorFrame.position.set(0, 0.71, depth * 0.457);
  root.add(doorFrame);
  const latch = box(0.03, 0.055, 0.018, 0xc4a05b, { metalness: 0.5 });
  latch.position.set(0.09, 0.36, depth * 0.455 + 0.024);
  root.add(latch);

  return createModelInstance(root);
};
