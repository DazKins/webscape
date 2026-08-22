import * as THREE from "three";
import { box, cylinder, sphere, sphereCap, taperedBox, torus } from "../primitives";
import { animation, createModelInstance, joint } from "../rig";
import type { ModelFactory, ModelPose } from "../types";

const TAU = Math.PI * 2;

export const createHumanModel: ModelFactory = (options = {}) => {
  const root = new THREE.Group();
  root.name = "human";

  const hips = joint("hips", root, [0, 0.61, 0]);
  const torso = joint("torso", hips, [0, 0.08, 0]);
  const head = joint("head", torso, [0, 0.4, 0]);
  const headwear = joint("headwear", head, [0, 0.07, 0]);
  const leftShoulder = joint("leftShoulder", torso, [0.23, 0.31, 0]);
  const rightShoulder = joint("rightShoulder", torso, [-0.23, 0.31, 0]);
  const leftElbow = joint("leftElbow", leftShoulder, [0, -0.28, 0]);
  const rightElbow = joint("rightElbow", rightShoulder, [0, -0.28, 0]);
  const rightHand = joint("rightHand", rightElbow, [0, -0.25, 0]);
  const leftHip = joint("leftHip", hips, [0.105, -0.07, 0]);
  const rightHip = joint("rightHip", hips, [-0.105, -0.07, 0]);
  const leftKnee = joint("leftKnee", leftHip, [0, -0.26, 0]);
  const rightKnee = joint("rightKnee", rightHip, [0, -0.26, 0]);

  const tunicColor = options.color ?? 0x4f8ab8;
  const tunicShadow = new THREE.Color(tunicColor).multiplyScalar(0.72);
  const tunicHighlight = new THREE.Color(tunicColor).offsetHSL(0, -0.04, 0.1);
  const skinColor = 0xe3b58e;
  const trousersColor = 0x34435a;
  const bootColor = 0x553b2d;

  const pelvis = box(0.34, 0.18, 0.22, trousersColor);
  hips.add(pelvis);

  const chest = taperedBox(0.38, 0.22, 0.3, 0.18, 0.32, tunicColor);
  chest.position.y = 0.16;
  torso.add(chest);

  const collar = torus(0.082, 0.018, 6, 10, tunicHighlight);
  collar.rotation.x = Math.PI / 2;
  collar.position.set(0, 0.325, 0.006);
  torso.add(collar);

  const belt = box(0.31, 0.06, 0.2, tunicShadow);
  belt.position.y = 0.035;
  torso.add(belt);

  const neck = cylinder(0.06, 0.07, 0.09, 8, skinColor);
  neck.position.y = -0.035;
  head.add(neck);
  const face = sphere(0.15, 10, 8, skinColor);
  face.scale.set(0.88, 1.05, 0.9);
  face.position.y = 0.105;
  head.add(face);
  const hair = sphereCap(0.156, 10, 4, Math.PI * 0.44, 0x4b3327);
  hair.scale.set(0.96, 1.08, 0.98);
  hair.position.set(0, 0.122, -0.004);
  head.add(hair);
  for (const eyeX of [-0.052, 0.052]) {
    const eye = sphere(0.014, 6, 4, 0x27241f);
    eye.position.set(eyeX, 0.125, 0.133);
    head.add(eye);
  }

  addArm(leftShoulder, leftElbow, skinColor, tunicColor);
  addArm(rightShoulder, rightElbow, skinColor, tunicColor);
  addLeg(leftHip, leftKnee, trousersColor, bootColor);
  addLeg(rightHip, rightKnee, trousersColor, bootColor);

  const joints = {
    hips,
    torso,
    head,
    leftShoulder,
    rightShoulder,
    leftElbow,
    rightElbow,
    leftHip,
    rightHip,
    leftKnee,
    rightKnee,
  };

  const idle = animation("idle", 2, true, (phase): ModelPose => {
    const wave = Math.sin(phase * TAU);
    const breath = (Math.sin(phase * TAU - Math.PI / 2) + 1) * 0.5;
    return {
      hips: { position: [0, wave * 0.008, 0] },
      torso: {
        rotation: [0, wave * 0.018, wave * 0.012],
        scale: [1 + breath * 0.008, 1 + breath * 0.015, 1 + breath * 0.008],
      },
      head: { rotation: [wave * 0.012, -wave * 0.025, 0] },
      leftShoulder: { rotation: [wave * 0.035, 0, -0.035] },
      rightShoulder: { rotation: [-wave * 0.035, 0, 0.035] },
    };
  });

  const run = animation("run", 0.8, true, (phase): ModelPose => {
    const stride = Math.sin(phase * TAU);
    const strideOpposite = Math.sin(phase * TAU + Math.PI);
    const bob = Math.abs(Math.sin(phase * TAU));
    const verticalBounce = Math.cos(phase * TAU * 2);
    const weightShift = stride;
    return {
      hips: {
        position: [weightShift * 0.012, verticalBounce * 0.022, 0],
        rotation: [0.035, stride * 0.025, -weightShift * 0.012],
      },
      torso: { rotation: [-0.055, -stride * 0.04, -weightShift * 0.018] },
      head: { rotation: [0.02 + bob * 0.01, stride * 0.015, weightShift * 0.008] },
      leftShoulder: { rotation: [strideOpposite * 0.48, 0, -0.025] },
      rightShoulder: { rotation: [stride * 0.48, 0, 0.025] },
      leftElbow: { rotation: [-0.12 - Math.max(0, stride) * 0.2, 0, 0] },
      rightElbow: { rotation: [-0.12 - Math.max(0, strideOpposite) * 0.2, 0, 0] },
      leftHip: { rotation: [stride * 0.48, 0, 0] },
      rightHip: { rotation: [strideOpposite * 0.48, 0, 0] },
      leftKnee: { rotation: [Math.max(0, -stride) * 0.58, 0, 0] },
      rightKnee: { rotation: [Math.max(0, -strideOpposite) * 0.58, 0, 0] },
    };
  });

  const instance = createModelInstance(root, joints, [idle, run], { headwear, rightHand });
  instance.play("idle");
  return instance;
};

function addArm(
  shoulder: THREE.Group,
  elbow: THREE.Group,
  skinColor: THREE.ColorRepresentation,
  tunicColor: THREE.ColorRepresentation,
) {
  const sleeve = cylinder(0.07, 0.065, 0.16, 7, tunicColor);
  sleeve.position.y = -0.08;
  shoulder.add(sleeve);
  const upperArm = cylinder(0.057, 0.052, 0.14, 7, skinColor);
  upperArm.position.y = -0.22;
  shoulder.add(upperArm);
  const forearm = cylinder(0.052, 0.045, 0.23, 7, skinColor);
  forearm.position.y = -0.115;
  elbow.add(forearm);
  const hand = sphere(0.058, 7, 5, skinColor);
  hand.position.y = -0.25;
  elbow.add(hand);
}

function addLeg(
  hip: THREE.Group,
  knee: THREE.Group,
  trousersColor: THREE.ColorRepresentation,
  bootColor: THREE.ColorRepresentation,
) {
  const thigh = cylinder(0.078, 0.066, 0.25, 7, trousersColor);
  thigh.position.y = -0.125;
  hip.add(thigh);
  const shin = cylinder(0.063, 0.055, 0.22, 7, bootColor);
  shin.position.y = -0.11;
  knee.add(shin);
  const foot = box(0.12, 0.08, 0.2, bootColor);
  foot.position.set(0, -0.235, 0.04);
  knee.add(foot);
}
