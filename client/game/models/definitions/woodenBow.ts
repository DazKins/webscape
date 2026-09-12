import * as THREE from "three";
import { cylinder, sphere } from "../primitives";
import { animation, createModelInstance, joint } from "../rig";
import type { ModelFactory } from "../types";

export const createWoodenBowModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "woodenBow";

  const wood = 0xc1813f;
  const darkWood = 0x51301c;
  const stringColor = 0xf0e5ca;
  const limbPoints = [
    new THREE.Vector3(-0.26, -0.55, 0),
    new THREE.Vector3(-0.12, -0.37, 0),
    new THREE.Vector3(-0.025, -0.18, 0),
    new THREE.Vector3(0, 0, 0),
    new THREE.Vector3(-0.025, 0.18, 0),
    new THREE.Vector3(-0.12, 0.37, 0),
    new THREE.Vector3(-0.26, 0.55, 0),
  ];
  for (let index = 0; index < limbPoints.length - 1; index++) {
    addSegment(root, limbPoints[index], limbPoints[index + 1], wood, 0.022 + (2 - Math.abs(index - 2.5)) * 0.004);
  }

  const gripPosition = new THREE.Vector3(0, 0, 0);
  const grip = cylinder(0.039, 0.039, 0.16, 8, darkWood, { roughness: 0.95 });
  grip.position.copy(gripPosition);
  root.add(grip);

  const lowerString = joint("lowerString", root, [-0.26, -0.55, 0]);
  const upperString = joint("upperString", root, [-0.26, 0.55, 0]);
  for (const [pivot, side] of [[lowerString, 1], [upperString, -1]] as const) {
    const string = cylinder(0.004, 0.004, 0.55, 5, stringColor);
    string.position.y = side * 0.275;
    pivot.add(string);
  }

  for (const tipPosition of [limbPoints[0], limbPoints[limbPoints.length - 1]]) {
    const tip = sphere(0.025, 7, 5, darkWood, { roughness: 0.9 });
    tip.scale.set(0.8, 1.15, 0.8);
    tip.position.copy(tipPosition);
    root.add(tip);
  }

  const arrowRest = new THREE.Group();
  arrowRest.name = "bowArrowRest";
  arrowRest.position.copy(gripPosition).add(new THREE.Vector3(0, 0, 0.12));
  root.add(arrowRest);

  const draw = animation("draw", 1.5, false, (phase) => {
    const tension = THREE.MathUtils.smoothstep(phase, 0, 2 / 3)
      * (1 - THREE.MathUtils.smoothstep(phase, 2 / 3, 0.74));
    const pull = tension * 0.18;
    const angle = Math.atan2(pull, 0.55);
    const stretch = Math.hypot(pull, 0.55) / 0.55;
    return {
      lowerString: { rotation: [0, 0, angle], scale: [1, stretch, 1] },
      upperString: { rotation: [0, 0, -angle], scale: [1, stretch, 1] },
    };
  });
  return createModelInstance(root, { lowerString, upperString }, [draw]);
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
