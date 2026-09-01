import * as THREE from "three";
import { cylinder, dodecahedron, sphere } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createMagicStaffModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "magicStaff";

  const wood = 0x6f482b;
  const darkWood = 0x3f281b;
  const shaftPoints = [
    new THREE.Vector3(0, 0, 0),
    new THREE.Vector3(0.012, 0.31, 0.006),
    new THREE.Vector3(-0.01, 0.62, -0.004),
    new THREE.Vector3(0.014, 0.93, 0.007),
    new THREE.Vector3(0.026, 1.22, 0),
  ];
  for (let index = 0; index < shaftPoints.length - 1; index++) {
    addConnectedShaftSegment(
      root,
      shaftPoints[index],
      shaftPoints[index + 1],
      index % 2 === 0 ? wood : darkWood,
      0.04 - index * 0.002,
    );
  }

  for (const [index, point] of shaftPoints.slice(1, -1).entries()) {
    const knot = sphere(0.045 - index * 0.003, 7, 5, index % 2 === 0 ? darkWood : wood, {
      roughness: 0.98,
    });
    knot.position.copy(point);
    knot.scale.set(1.12, 0.75, 0.96);
    root.add(knot);
  }

  const crownBase = sphere(0.065, 7, 5, darkWood, { roughness: 0.95 });
  crownBase.position.copy(shaftPoints[shaftPoints.length - 1]);
  crownBase.scale.set(1.05, 0.8, 1.05);
  root.add(crownBase);

  const crown = new THREE.Group();
  crown.name = "magicStaffCrown";
  crown.position.set(0.026, 1.31, 0);
  const crystal = dodecahedron(0.115, 0xd9f6ff, {
    roughness: 0.22,
    emissive: 0x8edcff,
  });
  crystal.scale.set(0.78, 1.25, 0.78);
  crown.add(crystal);
  const glow = sphere(0.065, 8, 6, 0xf3fdff, {
    roughness: 0.1,
    emissive: 0xc8f5ff,
  });
  crown.add(glow);
  root.add(crown);

  return createModelInstance(root);
};

function addConnectedShaftSegment(
  parent: THREE.Object3D,
  start: THREE.Vector3,
  end: THREE.Vector3,
  color: THREE.ColorRepresentation,
  radius: number,
) {
  const direction = end.clone().sub(start);
  const length = direction.length();
  const segment = cylinder(
    radius * 0.9,
    radius,
    length + 0.025,
    7,
    color,
    { roughness: 0.97 },
  );
  segment.position.copy(start).add(end).multiplyScalar(0.5);
  segment.quaternion.setFromUnitVectors(
    new THREE.Vector3(0, 1, 0),
    direction.normalize(),
  );
  parent.add(segment);
}
