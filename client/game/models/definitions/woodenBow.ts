import * as THREE from "three";
import { cylinder, sphere } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createWoodenBowModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "woodenBow";

  const wood = 0xc1813f;
  const darkWood = 0x51301c;
  const stringColor = 0xf0e5ca;
  const limbPoints = [
    new THREE.Vector3(-0.3, -0.45, 0),
    new THREE.Vector3(-0.16, -0.27, 0),
    new THREE.Vector3(-0.01, -0.07, 0),
    new THREE.Vector3(0, 0.07, 0),
    new THREE.Vector3(-0.06, 0.3, 0),
    new THREE.Vector3(-0.18, 0.52, 0),
    new THREE.Vector3(-0.3, 0.7, 0),
  ];
  for (let index = 0; index < limbPoints.length - 1; index++) {
    addSegment(root, limbPoints[index], limbPoints[index + 1], wood, 0.04);
  }

  const gripPosition = new THREE.Vector3(0, 0, 0);
  const grip = cylinder(0.055, 0.055, 0.2, 8, darkWood, { roughness: 0.95 });
  grip.position.copy(gripPosition);
  root.add(grip);

  addSegment(
    root,
    limbPoints[0],
    limbPoints[limbPoints.length - 1],
    stringColor,
    0.012,
  );

  for (const tipPosition of [limbPoints[0], limbPoints[limbPoints.length - 1]]) {
    const tip = sphere(0.05, 7, 5, darkWood, { roughness: 0.9 });
    tip.scale.set(0.8, 1.15, 0.8);
    tip.position.copy(tipPosition);
    root.add(tip);
  }

  const arrowRest = new THREE.Group();
  arrowRest.name = "bowArrowRest";
  arrowRest.position.copy(gripPosition).add(new THREE.Vector3(0, 0, 0.12));
  root.add(arrowRest);

  return createModelInstance(root);
};

function addSegment(
  parent: THREE.Object3D,
  start: THREE.Vector3,
  end: THREE.Vector3,
  color: THREE.ColorRepresentation,
  radius: number,
) {
  const direction = end.clone().sub(start);
  const segment = cylinder(radius, radius, direction.length(), 7, color, {
    roughness: 0.9,
  });
  segment.position.copy(start).add(end).multiplyScalar(0.5);
  segment.quaternion.setFromUnitVectors(
    new THREE.Vector3(0, 1, 0),
    direction.normalize(),
  );
  parent.add(segment);
}
