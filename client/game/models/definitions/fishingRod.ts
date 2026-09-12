import * as THREE from "three";
import { cylinder, sphere, torus } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createFishingRodModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "fishingRod";

  const rod = new THREE.Group();
  rod.name = "fishingRodRoll";
  rod.rotation.y = Math.PI / 2;
  root.add(rod);

  const grip = cylinder(0.03, 0.038, 0.22, 8, 0x4c3020);
  grip.position.y = 0.11;
  rod.add(grip);

  const shaft = cylinder(0.012, 0.025, 0.86, 8, 0x8a5d32, { roughness: 0.8 });
  shaft.position.y = 0.65;
  shaft.position.x = 0.035;
  shaft.rotation.z = -0.08;
  rod.add(shaft);

  const tip = sphere(0.025, 7, 5, 0xb99355);
  tip.position.set(0.07, 1.08, 0);
  rod.add(tip);

  const reel = torus(0.09, 0.018, 6, 12, 0x6f7b82, {
    roughness: 0.35,
    metalness: 0.55,
  });
  reel.position.set(-0.055, 0.29, 0);
  reel.rotation.y = Math.PI / 2;
  rod.add(reel);

  const reelHub = cylinder(0.022, 0.022, 0.08, 8, 0xc3c9cc, {
    roughness: 0.3,
    metalness: 0.6,
  });
  reelHub.position.copy(reel.position);
  reelHub.rotation.z = Math.PI / 2;
  rod.add(reelHub);
  for (const y of [0.48, 0.75, 1.02]) {
    const guide = torus(0.024, 0.005, 4, 8, 0xb1bac0, { metalness: 0.5 });
    guide.rotation.x = Math.PI / 2;
    guide.position.set((y - 0.22) * 0.08, y, 0.025);
    rod.add(guide);
  }
  const line = cylinder(0.0025, 0.0025, 0.74, 4, 0xd9d6bb);
  line.position.set(0.028, 0.67, 0.03);
  line.rotation.z = -0.08;
  rod.add(line);
  const crank = cylinder(0.01, 0.01, 0.09, 6, 0xb1bac0, { metalness: 0.5 });
  crank.position.set(-0.1, 0.25, 0);
  rod.add(crank);
  const knob = sphere(0.024, 6, 4, 0x4c3020);
  knob.position.set(-0.11, 0.21, 0);
  rod.add(knob);

  return createModelInstance(root);
};
