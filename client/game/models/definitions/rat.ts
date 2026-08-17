import * as THREE from "three";
import { cone, cylinder, sphere } from "../primitives";
import { animation, createModelInstance, joint } from "../rig";
import type { ModelFactory, ModelPose } from "../types";

const TAU = Math.PI * 2;

export const createRatModel: ModelFactory = (options = {}) => {
  const root = new THREE.Group();
  root.name = "rat";

  const body = joint("body", root, [0, 0.3, -0.02]);
  const head = joint("head", body, [0, 0.035, 0.3]);
  const leftEar = joint("leftEar", head, [0.105, 0.12, 0.02]);
  const rightEar = joint("rightEar", head, [-0.105, 0.12, 0.02]);
  const tailBase = joint("tailBase", body, [0, 0, -0.36]);
  const tailMiddle = joint("tailMiddle", tailBase, [0, 0, -0.25]);
  const tailTip = joint("tailTip", tailMiddle, [0, 0, -0.21]);
  const leftFrontLeg = joint("leftFrontLeg", body, [0.16, -0.14, 0.2]);
  const rightFrontLeg = joint("rightFrontLeg", body, [-0.16, -0.14, 0.2]);
  const leftBackLeg = joint("leftBackLeg", body, [0.2, -0.13, -0.22]);
  const rightBackLeg = joint("rightBackLeg", body, [-0.2, -0.13, -0.22]);

  const furColor = options.color ?? 0x74675f;
  const darkFurColor = new THREE.Color(furColor).multiplyScalar(0.7);
  const bellyColor = new THREE.Color(furColor).offsetHSL(0.02, -0.12, 0.13);
  const skinColor = 0xd49391;
  const innerEarColor = 0xeab0ad;
  const eyeColor = 0x171313;

  const torso = sphere(0.28, 9, 6, furColor);
  torso.scale.set(0.88, 0.72, 1.36);
  body.add(torso);

  const belly = sphere(0.2, 8, 5, bellyColor);
  belly.scale.set(0.86, 0.42, 1.25);
  belly.position.set(0, -0.135, 0.02);
  body.add(belly);

  const face = sphere(0.2, 9, 6, furColor);
  face.scale.set(0.78, 0.76, 1.05);
  face.position.z = 0.075;
  head.add(face);

  const muzzle = cone(0.105, 0.28, 8, darkFurColor);
  muzzle.rotation.x = Math.PI / 2;
  muzzle.position.set(0, -0.035, 0.235);
  head.add(muzzle);

  const nose = sphere(0.052, 7, 5, 0x4a2b30);
  nose.scale.set(1.1, 0.8, 0.8);
  nose.position.set(0, -0.035, 0.39);
  head.add(nose);

  addEar(leftEar, furColor, innerEarColor);
  addEar(rightEar, furColor, innerEarColor);

  for (const eyeX of [-0.105, 0.105]) {
    const eye = sphere(0.03, 7, 5, eyeColor, { roughness: 0.3 });
    eye.position.set(eyeX, 0.045, 0.205);
    head.add(eye);

    const catchlight = sphere(0.008, 5, 4, 0xffffff, { emissive: 0x333333 });
    catchlight.position.set(eyeX + Math.sign(eyeX) * 0.004, 0.055, 0.23);
    head.add(catchlight);
  }

  addWhiskers(head);
  addFoot(leftFrontLeg, skinColor, 0.11);
  addFoot(rightFrontLeg, skinColor, 0.11);
  addFoot(leftBackLeg, skinColor, 0.15);
  addFoot(rightBackLeg, skinColor, 0.15);
  addTailSegment(tailBase, 0.045, 0.035, 0.25, skinColor);
  addTailSegment(tailMiddle, 0.035, 0.024, 0.21, skinColor);
  addTailSegment(tailTip, 0.024, 0.009, 0.16, skinColor);

  const joints = {
    body,
    head,
    leftEar,
    rightEar,
    tailBase,
    tailMiddle,
    tailTip,
    leftFrontLeg,
    rightFrontLeg,
    leftBackLeg,
    rightBackLeg,
  };

  const idle = animation("idle", 1.8, true, (phase): ModelPose => {
    const sniff = Math.sin(phase * TAU * 2);
    const sway = Math.sin(phase * TAU);
    return {
      body: { position: [0, Math.max(0, sniff) * 0.008, 0] },
      head: { rotation: [sniff * 0.055, sway * 0.08, sway * 0.018] },
      leftEar: { rotation: [0, 0, sniff * 0.08] },
      rightEar: { rotation: [0, 0, -sniff * 0.08] },
      tailBase: { rotation: [0, sway * 0.16, 0] },
      tailMiddle: { rotation: [0, -sway * 0.2, 0] },
      tailTip: { rotation: [0, sway * 0.25, 0] },
    };
  });

  const run = animation("run", 0.42, true, (phase): ModelPose => {
    const stride = Math.sin(phase * TAU);
    const oppositeStride = Math.sin(phase * TAU + Math.PI);
    const bob = Math.abs(stride);
    return {
      body: {
        position: [0, bob * 0.035, 0],
        rotation: [-0.04 + bob * 0.06, stride * 0.025, stride * 0.018],
      },
      head: { rotation: [-bob * 0.07, -stride * 0.035, 0] },
      leftFrontLeg: { rotation: [stride * 0.65, 0, 0] },
      rightFrontLeg: { rotation: [oppositeStride * 0.65, 0, 0] },
      leftBackLeg: { rotation: [oppositeStride * 0.55, 0, 0] },
      rightBackLeg: { rotation: [stride * 0.55, 0, 0] },
      tailBase: { rotation: [0.08, -stride * 0.22, 0] },
      tailMiddle: { rotation: [0.06, stride * 0.3, 0] },
      tailTip: { rotation: [0.04, -stride * 0.38, 0] },
    };
  });

  const instance = createModelInstance(root, joints, [idle, run]);
  instance.play("idle");
  return instance;
};

function addEar(
  earJoint: THREE.Group,
  furColor: THREE.ColorRepresentation,
  innerEarColor: THREE.ColorRepresentation,
) {
  const outerEar = sphere(0.105, 8, 5, furColor);
  outerEar.scale.set(0.9, 1.15, 0.32);
  earJoint.add(outerEar);

  const innerEar = sphere(0.073, 7, 5, innerEarColor);
  innerEar.scale.set(0.88, 1.1, 0.22);
  innerEar.position.z = 0.032;
  earJoint.add(innerEar);
}

function addFoot(legJoint: THREE.Group, color: THREE.ColorRepresentation, length: number) {
  const leg = cylinder(0.035, 0.025, 0.13, 6, color);
  leg.position.y = -0.055;
  leg.rotation.z = Math.PI * 0.08;
  legJoint.add(leg);

  const paw = sphere(0.055, 7, 4, color);
  paw.scale.set(0.8, 0.42, length / 0.055);
  paw.position.set(0, -0.12, length * 0.38);
  legJoint.add(paw);
}

function addTailSegment(
  tailJoint: THREE.Group,
  radiusTop: number,
  radiusBottom: number,
  length: number,
  color: THREE.ColorRepresentation,
) {
  const segment = cylinder(radiusTop, radiusBottom, length, 7, color);
  segment.rotation.x = Math.PI / 2;
  segment.position.z = -length / 2;
  tailJoint.add(segment);
}

function addWhiskers(head: THREE.Group) {
  const whiskerColor = 0xdbc9bc;
  for (const side of [-1, 1]) {
    for (const [index, y] of [-0.065, -0.02, 0.025].entries()) {
      const whisker = cylinder(0.004, 0.002, 0.28, 5, whiskerColor);
      whisker.rotation.z = Math.PI / 2 + side * (index - 1) * 0.08;
      whisker.rotation.x = side * 0.2;
      whisker.position.set(side * 0.18, y, 0.27);
      head.add(whisker);
    }
  }
}
