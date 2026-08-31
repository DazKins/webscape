import * as THREE from "three";
import { cylinder, sphere, torus } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createFishingSpotModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "fishingSpot";

  const hitboxMaterial = new THREE.MeshBasicMaterial({
    transparent: true,
    opacity: 0,
    depthWrite: false,
    colorWrite: false,
  });
  const hitbox = new THREE.Mesh(new THREE.BoxGeometry(1, 0.08, 1), hitboxMaterial);
  hitbox.name = "fishingSpotHitbox";
  hitbox.position.y = 0.04;
  root.add(hitbox);

  for (const [radius, opacityOffset] of [[0.34, 0], [0.21, 0.01]] as const) {
    const ripple = torus(radius, 0.018, 5, 20, opacityOffset === 0 ? 0x8fd9e8 : 0xc6f1f4, {
      roughness: 0.25,
      metalness: 0.05,
    });
    ripple.rotation.x = Math.PI / 2;
    ripple.position.y = 0.035 + opacityOffset;
    ripple.scale.z = 0.72;
    root.add(ripple);
  }

  const bobber = sphere(0.075, 8, 6, 0xf3f0dc);
  bobber.position.set(0.08, 0.105, 0.02);
  root.add(bobber);
  const cap = cylinder(0.018, 0.018, 0.13, 6, 0xe44e3f);
  cap.position.set(0.08, 0.2, 0.02);
  root.add(cap);

  return createModelInstance(root);
};
