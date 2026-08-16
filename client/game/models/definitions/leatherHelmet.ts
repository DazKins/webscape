import * as THREE from "three";
import { box, sphereCap, torus } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createLeatherHelmetModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "leatherHelmet";

  const leather = 0x765035;
  const darkLeather = 0x4c3425;
  const rivet = 0xb99862;

  const crown = sphereCap(0.18, 10, 5, Math.PI / 2, leather);
  crown.scale.set(1, 0.64, 0.94);
  crown.position.y = 0.1;
  root.add(crown);

  const band = torus(0.171, 0.018, 6, 12, darkLeather);
  band.rotation.x = Math.PI / 2;
  band.position.y = 0.11;
  root.add(band);

  const noseGuard = box(0.035, 0.105, 0.025, darkLeather);
  noseGuard.position.set(0, 0.053, 0.174);
  root.add(noseGuard);

  for (const side of [-1, 1]) {
    const earGuard = box(0.055, 0.085, 0.09, leather);
    earGuard.position.set(side * 0.158, 0.055, 0);
    root.add(earGuard);

    const sideRivet = box(0.018, 0.018, 0.012, rivet, { metalness: 0.45 });
    sideRivet.position.set(side * 0.188, 0.11, 0.025);
    root.add(sideRivet);
  }

  return createModelInstance(root);
};
