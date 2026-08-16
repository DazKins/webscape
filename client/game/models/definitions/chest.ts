import * as THREE from "three";
import { box } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createChestModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "chest";

  const body = box(0.72, 0.42, 0.54, 0x9a5d2e);
  body.position.y = 0.21;
  root.add(body);
  const lid = box(0.78, 0.16, 0.6, 0xb47b37);
  lid.position.y = 0.5;
  root.add(lid);
  const band = box(0.13, 0.58, 0.62, 0x725034, { metalness: 0.15 });
  band.position.y = 0.29;
  root.add(band);
  const latch = box(0.12, 0.12, 0.04, 0xd2b46d, { metalness: 0.35 });
  latch.position.set(0, 0.38, 0.31);
  root.add(latch);

  return createModelInstance(root);
};
