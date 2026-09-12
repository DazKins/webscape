import * as THREE from "three";
import { box, cylinder, dodecahedron, sphere, sphereCap, taperedBox, torus } from "../primitives";
import { humanAppearanceColors, type HairStyle } from "../humanAppearance";
import { animation, createModelInstance, joint } from "../rig";
import type { ModelFactory, ModelPose } from "../types";

const TAU = Math.PI * 2;

export const HUMAN_CHOP_ANIMATION_SECONDS = 0.6;
export const HUMAN_CHOP_CONTACT_SECONDS = 0.5;
export const HUMAN_PICKUP_ANIMATION_SECONDS = 0.65;

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
  const leftHand = joint("leftHand", leftElbow, [0, -0.25, 0]);
  const leftHip = joint("leftHip", hips, [0.105, -0.07, 0]);
  const rightHip = joint("rightHip", hips, [-0.105, -0.07, 0]);
  const leftKnee = joint("leftKnee", leftHip, [0, -0.26, 0]);
  const rightKnee = joint("rightKnee", rightHip, [0, -0.26, 0]);

  const leftAnkle = joint("leftAnkle", leftKnee, [0, -0.235, 0]);
  const rightAnkle = joint("rightAnkle", rightKnee, [0, -0.235, 0]);

  const appearance = options.appearance;
  const tunicColor = appearance
    ? humanAppearanceColors.tunic[appearance.tunicColor]
    : options.color ?? 0x4f8ab8;
  const tunicShadow = new THREE.Color(tunicColor).multiplyScalar(0.72);
  const tunicHighlight = new THREE.Color(tunicColor).offsetHSL(0, -0.04, 0.1);
  const skinColor = appearance
    ? humanAppearanceColors.skin[appearance.skinTone]
    : 0xe3b58e;
  const hairColor = appearance
    ? humanAppearanceColors.hair[appearance.hairColor]
    : 0x4b3327;
  const trousersColor = appearance
    ? humanAppearanceColors.trousers[appearance.trousersColor]
    : 0x34435a;
  const bootColor = appearance
    ? humanAppearanceColors.shoes[appearance.shoeColor]
    : 0x553b2d;

  const pelvis = taperedBox(0.3, 0.2, 0.33, 0.22, 0.18, tunicShadow);
  hips.add(pelvis);

  const chest = taperedBox(0.38, 0.22, 0.3, 0.18, 0.32, tunicColor);
  chest.position.y = 0.16;
  torso.add(chest);

  const collar = torus(0.082, 0.018, 6, 10, tunicHighlight);
  collar.rotation.x = Math.PI / 2;
  collar.position.set(0, 0.325, 0.006);
  torso.add(collar);

  const belt = box(0.32, 0.05, 0.21, 0x543c2c);
  belt.position.y = 0.035;
  torso.add(belt);
  const buckle = box(0.065, 0.055, 0.022, 0xc7a05a, { metalness: 0.45 });
  buckle.position.set(0, 0.035, 0.114);
  torso.add(buckle);
  const buckleInset = box(0.033, 0.028, 0.006, 0x543c2c);
  buckleInset.position.set(0, 0.035, 0.128);
  torso.add(buckleInset);
  const placket = box(0.025, 0.16, 0.012, tunicShadow);
  placket.position.set(0, 0.23, 0.108);
  torso.add(placket);

  const neck = cylinder(0.06, 0.07, 0.09, 8, skinColor);
  neck.position.y = -0.035;
  head.add(neck);
  const face = sphere(0.15, 10, 8, skinColor);
  face.scale.set(0.88, 1.05, 0.9);
  face.position.y = 0.105;
  head.add(face);
  const hair = new THREE.Group();
  hair.name = "hair";
  head.add(hair);
  addHair(hair, appearance?.hairStyle ?? "cropped", hairColor);
  for (const eyeX of [-0.052, 0.052]) {
    const eye = sphere(0.014, 6, 4, 0x27241f);
    eye.position.set(eyeX, 0.125, 0.133);
    head.add(eye);
  }

  const nose = sphere(0.028, 6, 4, skinColor);
  nose.scale.set(0.75, 0.85, 1);
  nose.position.set(0, 0.09, 0.139);
  head.add(nose);
  for (const side of [-1, 1]) {
    const ear = sphere(0.032, 6, 4, skinColor);
    ear.scale.set(0.5, 1, 0.7);
    ear.position.set(side * 0.132, 0.09, 0);
    head.add(ear);
  }

  addArm(leftShoulder, leftElbow, leftHand, skinColor, tunicColor);
  addArm(rightShoulder, rightElbow, rightHand, skinColor, tunicColor);
  addLeg(leftHip, leftKnee, leftAnkle, trousersColor, bootColor);
  addLeg(rightHip, rightKnee, rightAnkle, trousersColor, bootColor);

  const joints = {
    hips,
    torso,
    head,
    leftShoulder,
    rightShoulder,
    leftElbow,
    rightElbow,
    rightHand,
    leftHand,
    leftHip,
    rightHip,
    leftKnee,
    rightKnee,
    leftAnkle,
    rightAnkle,
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
    const stride = Math.cos(phase * TAU);
    const bounce = -0.06 + Math.sin(phase * TAU) ** 2 * 0.025;
    return {
      hips: { position: [stride * 0.008, bounce, 0] },
      torso: { rotation: [0.045, stride * 0.035, stride * 0.012] },
      head: { rotation: [-0.035, -stride * 0.02, 0] },
      leftShoulder: { rotation: [-stride * 0.48, 0, -0.04] },
      rightShoulder: { rotation: [stride * 0.48, 0, 0.04] },
      leftElbow: { rotation: [-0.28 - Math.max(0, stride) * 0.2, 0, 0] },
      rightElbow: { rotation: [-0.28 - Math.max(0, -stride) * 0.2, 0, 0] },
      ...strideLegPose("left", phase, bounce),
      ...strideLegPose("right", phase + 0.5, bounce),
    };
  });

  const pickup = animation("pickup", HUMAN_PICKUP_ANIMATION_SECONDS, false, (phase): ModelPose => {
    const bend = THREE.MathUtils.smoothstep(phase, 0, 0.4)
      * (1 - THREE.MathUtils.smoothstep(phase, 0.55, 1));
    const legBend = 1.0 * bend;
    return {
      hips: { position: [0, -0.52 * (1 - Math.cos(legBend)), 0] },
      leftHip: { rotation: [-legBend, 0, 0] },
      rightHip: { rotation: [-legBend, 0, 0] },
      leftKnee: { rotation: [2 * legBend, 0, 0] },
      rightKnee: { rotation: [2 * legBend, 0, 0] },
      leftAnkle: { rotation: [-legBend, 0, 0] },
      rightAnkle: { rotation: [-legBend, 0, 0] },
      torso: { rotation: [1.05 * bend, 0, 0] },
      head: { rotation: [0.15 * bend, 0, 0] },
      rightShoulder: { rotation: [-0.85 * bend, 0, -0.12 * bend] },
      rightElbow: { rotation: [-0.15 * bend, 0, 0] },
      leftShoulder: { rotation: [-0.2 * bend, 0, -0.08 * bend] },
    };
  });

  const attack = animation("attack", 0.38, false, (phase): ModelPose => {
    const windUp = THREE.MathUtils.smoothstep(phase, 0, 0.28);
    const strike = THREE.MathUtils.smoothstep(phase, 0.28, 0.58);
    const recover = THREE.MathUtils.smoothstep(phase, 0.58, 1);
    const shoulderSwing = -1.9 * windUp + 0.9 * strike + recover;
    const elbowBend = THREE.MathUtils.lerp(0, -0.65, windUp)
      + THREE.MathUtils.lerp(0, 0.5, strike)
      + THREE.MathUtils.lerp(0, 0.15, recover);
    const torsoTwist = THREE.MathUtils.lerp(0, 0.22, windUp)
      + THREE.MathUtils.lerp(0, -0.42, strike)
      + THREE.MathUtils.lerp(0, 0.2, recover);
    const followThrough = strike * (1 - recover);
    return {
      hips: { rotation: [0, -torsoTwist * 0.25, 0] },
      torso: { rotation: [0.08 * followThrough, torsoTwist, -0.04 * followThrough] },
      rightShoulder: { rotation: [shoulderSwing, 0, 0.12 * (windUp - recover)] },
      rightElbow: { rotation: [elbowBend, 0, 0] },
      rightHand: { rotation: [-0.25 * windUp + 1.35 * strike - 1.1 * recover, 0, 0] },
      leftShoulder: { rotation: [-0.12 * followThrough, 0, -0.04 * followThrough] },
    };
  });

  const cast = animation("cast", 1.5, false, (phase): ModelPose => {
    const gather = THREE.MathUtils.smoothstep(phase, 0, 0.45);
    const release = THREE.MathUtils.smoothstep(phase, 0.45, 2 / 3);
    const recover = THREE.MathUtils.smoothstep(phase, 2 / 3, 1);
    const intensity = Math.max(gather, release) * (1 - recover);
    return {
      hips: { rotation: [-0.035 * intensity, 0, 0] },
      torso: { rotation: [-0.08 * intensity, -0.08 * intensity, 0] },
      head: { rotation: [0.07 * intensity, 0.05 * intensity, 0] },
      rightShoulder: {
        rotation: [-1.12 * intensity, 0.12 * intensity, 0.22 * intensity],
      },
      rightElbow: { rotation: [-0.72 * intensity, 0, 0] },
      rightHand: { rotation: [-0.22 * intensity, 0, 0.16 * intensity] },
      leftShoulder: {
        rotation: [-0.82 * intensity, -0.12 * intensity, -0.34 * intensity],
      },
      leftElbow: { rotation: [-1.0 * intensity, 0, 0] },
      leftHand: { rotation: [0.15 * intensity, 0, -0.2 * intensity] },
      ...staffArmPose(0.25 + 0.1 * intensity, 0.12 * intensity),
    };
  });

  const shoot = animation("shoot", 1.5, false, (phase): ModelPose => {
    const draw = THREE.MathUtils.smoothstep(phase, 0, 2 / 3);
    const recover = THREE.MathUtils.smoothstep(phase, 2 / 3, 1);
    const aim = draw * (1 - recover);
    return {
      hips: { rotation: [0, -0.08 * aim, 0] },
      torso: { rotation: [-0.04 * aim, -0.2 * aim, 0] },
      head: { rotation: [0.03 * aim, 0.15 * aim, 0] },
      leftShoulder: { rotation: [-1.42 * aim, -0.08 * aim, -0.08 * aim] },
      leftElbow: { rotation: [-0.12 * aim, 0, 0] },
      leftHand: { rotation: [2.1 + (1.54 - 2.1) * aim, 0, -0.12 * aim] },
      rightShoulder: { rotation: [-1.08 * aim, 0.55 * aim, 0.18 * aim] },
      rightElbow: { rotation: [-1.15 * aim, 0, 0] },
      rightHand: { rotation: [0, 0, 0.25 * aim] },
    };
  });

  const chop = animation("chop", HUMAN_CHOP_ANIMATION_SECONDS, false, (phase): ModelPose => {
    const windUpEnd = 0.3;
    const sideEnd = 0.48;
    const hitEnd = HUMAN_CHOP_CONTACT_SECONDS / HUMAN_CHOP_ANIMATION_SECONDS;
    const down = new THREE.Vector3(0, -1, 0);
    const windUpDirection = new THREE.Vector3(
      -Math.cos(THREE.MathUtils.degToRad(10)) / Math.sqrt(2),
      Math.sin(THREE.MathUtils.degToRad(10)),
      -Math.cos(THREE.MathUtils.degToRad(10)) / Math.sqrt(2),
    );
    const sideDirection = new THREE.Vector3(-1, 0, 0);
    const hitDirection = new THREE.Vector3(
      0,
      -Math.sin(THREE.MathUtils.degToRad(15)),
      Math.cos(THREE.MathUtils.degToRad(15)),
    );

    let armDirection: THREE.Vector3;
    if (phase <= windUpEnd) {
      const progress = THREE.MathUtils.smoothstep(phase, 0, windUpEnd);
      armDirection = down.clone().lerp(windUpDirection, progress).normalize();
    } else if (phase <= sideEnd) {
      const progress = THREE.MathUtils.smoothstep(phase, windUpEnd, sideEnd);
      armDirection = windUpDirection.clone().lerp(sideDirection, progress).normalize();
    } else if (phase <= hitEnd) {
      const progress = THREE.MathUtils.smoothstep(phase, sideEnd, hitEnd);
      armDirection = sideDirection.clone().lerp(hitDirection, progress).normalize();
    } else {
      const progress = THREE.MathUtils.smoothstep(phase, hitEnd, 1);
      armDirection = hitDirection.clone().lerp(down, progress).normalize();
    }

    const shoulderRotation = new THREE.Quaternion().setFromUnitVectors(down, armDirection);
    const shoulderEuler = new THREE.Euler().setFromQuaternion(shoulderRotation, "XYZ");
    const idleHandRotation = new THREE.Quaternion();
    const axeHalfTurn = new THREE.Quaternion().setFromAxisAngle(
      new THREE.Vector3(0, -Math.sin(0.08), Math.cos(0.08)),
      Math.PI,
    );
    const windUpHandRotation = new THREE.Quaternion().setFromEuler(
      new THREE.Euler(THREE.MathUtils.degToRad(60), 0, 0),
    ).multiply(axeHalfTurn);
    const hitHandRotation = new THREE.Quaternion().setFromEuler(
      new THREE.Euler(
        THREE.MathUtils.degToRad(-105),
        Math.PI / 2 - 0.08,
        Math.PI / 2,
      ),
    ).multiply(axeHalfTurn);
    let handRotation: THREE.Quaternion;
    if (phase <= windUpEnd) {
      const progress = THREE.MathUtils.smoothstep(phase, 0, windUpEnd);
      handRotation = new THREE.Quaternion().slerpQuaternions(
        idleHandRotation,
        windUpHandRotation,
        progress,
      );
    } else if (phase <= hitEnd) {
      const progress = THREE.MathUtils.smoothstep(phase, windUpEnd, hitEnd);
      handRotation = new THREE.Quaternion().slerpQuaternions(
        windUpHandRotation,
        hitHandRotation,
        progress,
      );
    } else {
      const progress = THREE.MathUtils.smoothstep(phase, hitEnd, 1);
      handRotation = new THREE.Quaternion().slerpQuaternions(
        hitHandRotation,
        idleHandRotation,
        progress,
      );
    }
    const handEuler = new THREE.Euler().setFromQuaternion(handRotation, "XYZ");
    const effort = Math.sin(phase * Math.PI) ** 2;
    return {
      torso: { rotation: [0.035 * effort, 0.065 * effort, 0] },
      head: { rotation: [-0.025 * effort, -0.04 * effort, 0] },
      leftShoulder: { rotation: [-0.18 * effort, 0, -0.12 * effort] },
      leftElbow: { rotation: [-0.3 * effort, 0, 0] },
      rightShoulder: { rotation: [shoulderEuler.x, shoulderEuler.y, shoulderEuler.z] },
      rightHand: { rotation: [handEuler.x, handEuler.y, handEuler.z] },
    };
  });

  const fishWait = animation("fishWait", 3.2, true, (phase): ModelPose => {
    const wave = Math.sin(phase * Math.PI * 2);
    return {
      hips: { position: [0, wave * 0.003, 0] },
      torso: { rotation: [0.025 + wave * 0.006, -0.025 + wave * 0.004, 0] },
      head: { rotation: [-0.015 + wave * 0.004, wave * 0.006, 0] },
      rightShoulder: { rotation: [-0.84 + wave * 0.018, 0.08, 0.16] },
      rightElbow: { rotation: [-0.7 + wave * 0.015, 0, 0] },
      rightHand: { rotation: [0.65 + wave * 0.025, 0, wave * 0.018] },
      leftShoulder: { rotation: [-0.32 - wave * 0.012, 0, -0.16] },
      leftElbow: { rotation: [-0.44 + wave * 0.012, 0, 0] },
    };
  });

  const fishAction = animation("fishAction", 0.5, false, (phase): ModelPose => {
    const pull = Math.sin(phase * Math.PI);
    const recover = Math.sin(Math.min(1, phase * 1.35) * Math.PI);
    return {
      hips: { rotation: [-0.08 * pull, 0, 0] },
      torso: { rotation: [0.025 - 0.22 * pull, -0.025 - 0.08 * pull, 0.06 * pull] },
      head: { rotation: [-0.015 + 0.12 * pull, 0, 0] },
      rightShoulder: { rotation: [-0.84 - 0.72 * pull, 0.08, 0.16] },
      rightElbow: { rotation: [-0.7 - 0.35 * recover, 0, 0] },
      rightHand: { rotation: [0.65 - 0.65 * pull, 0, 0.22 * pull] },
      leftShoulder: { rotation: [-0.32 - 0.28 * pull, 0, -0.16] },
      leftElbow: { rotation: [-0.44 - 0.2 * pull, 0, 0] },
    };
  });

  const bowIdle = animation("bowIdle", 2, true, (phase) => ({
    ...idle.sample(phase),
    leftHand: { rotation: [2.1, 0, 0] },
  }));
  const bowRun = animation("bowRun", 0.8, true, (phase) => ({
    ...run.sample(phase),
    leftShoulder: { rotation: [0, 0, -0.08] },
    leftElbow: { rotation: [-0.12, 0, 0] },
    leftHand: { rotation: [2.22, 0, 0] },
  }));
  const staffIdle = animation("staffIdle", 2, true, (phase) => ({
    ...idle.sample(phase), ...staffArmPose(0.25, 0),
  }));
  const staffRun = animation("staffRun", 0.8, true, (phase) => {
    // Lift the planted tip before the grip passes the body, keeping the
    // stride compact. The wrist cancels the arm's rotation throughout.
    const plantedUntil = 0.3;
    const planted = phase < plantedUntil;
    const swing = (phase - plantedUntil) / (1 - plantedUntil);
    const z = planted ? THREE.MathUtils.lerp(0.3, 0.1, THREE.MathUtils.smoothstep(phase, 0, plantedUntil))
      : THREE.MathUtils.lerp(0.1, 0.3, THREE.MathUtils.smoothstep(swing, 0, 1));
    const lift = planted ? 0 : Math.sin(swing * Math.PI) ** 2 * 0.06;
    const stridePose = run.sample(phase);
    return { ...stridePose, ...staffArmPose(z, lift, stridePose.hips.position?.[1] ?? 0) };
  });

  const instance = createModelInstance(root, joints, [
    idle,
    run,
    bowIdle,
    bowRun,
    staffIdle,
    staffRun,
    attack,
    pickup,
    cast,
    shoot,
    chop,
    fishWait,
    fishAction,
  ], {
    headwear,
    rightHand,
    leftHand,
  });
  instance.play("idle");
  return instance;
};

function addHair(
  head: THREE.Group,
  style: HairStyle,
  color: THREE.ColorRepresentation,
) {
  switch (style) {
    case "cropped": {
      const cap = sphereCap(0.156, 10, 4, Math.PI * 0.44, color);
      cap.scale.set(0.96, 1.08, 0.98);
      cap.position.set(0, 0.122, -0.004);
      head.add(cap);
      break;
    }
    case "swept": {
      const cap = sphereCap(0.158, 10, 4, Math.PI * 0.48, color);
      cap.scale.set(0.98, 1.06, 1);
      cap.position.set(0, 0.124, -0.006);
      head.add(cap);

      const fringe = taperedBox(0.16, 0.055, 0.055, 0.045, 0.075, color);
      fringe.position.set(0.028, 0.19, 0.124);
      fringe.rotation.z = -0.34;
      fringe.rotation.x = -0.12;
      head.add(fringe);
      break;
    }
    case "bob": {
      const cap = sphereCap(0.16, 10, 4, Math.PI * 0.52, color);
      cap.scale.set(1, 1.04, 1);
      cap.position.set(0, 0.122, -0.008);
      head.add(cap);
      for (const side of [-1, 1]) {
        const sideLock = taperedBox(0.045, 0.105, 0.035, 0.08, 0.19, color);
        sideLock.position.set(side * 0.137, 0.055, -0.012);
        sideLock.rotation.z = side * -0.035;
        head.add(sideLock);
      }
      const back = taperedBox(0.22, 0.045, 0.18, 0.035, 0.16, color);
      back.position.set(0, 0.065, -0.132);
      head.add(back);
      break;
    }
    case "curls": {
      const curlPositions: ReadonlyArray<readonly [number, number, number]> = [
        [-0.09, 0.22, -0.035], [0, 0.245, -0.04], [0.09, 0.22, -0.035],
        [-0.13, 0.15, -0.045], [-0.05, 0.17, -0.105], [0.05, 0.17, -0.105],
        [0.13, 0.15, -0.045], [-0.135, 0.075, -0.035], [0.135, 0.075, -0.035],
      ];
      for (const [x, y, z] of curlPositions) {
        const curl = dodecahedron(0.068, color);
        curl.position.set(x, y, z);
        curl.scale.y = 0.9;
        head.add(curl);
      }
      break;
    }
  }
}

function addArm(
  shoulder: THREE.Group,
  elbow: THREE.Group,
  wrist: THREE.Group,
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
  wrist.add(hand);
}

// Solve a two-segment arm in the character's vertical forward plane. The
// staff grip is 0.78 units above its tip, so a zero lift places it on the floor.
function staffArmPose(z: number, lift: number, bounce = 0): ModelPose {
  const y = 0.78 + lift - 1.0 - bounce;
  const upper = 0.28;
  const lower = 0.25;
  const elbow = -Math.acos(THREE.MathUtils.clamp(
    (y * y + z * z - upper * upper - lower * lower) / (2 * upper * lower), -1, 1,
  ));
  const shoulder = Math.atan2(-z, -y)
    - Math.atan2(lower * Math.sin(elbow), upper + lower * Math.cos(elbow));
  return {
    hips: { position: [0, bounce, 0], rotation: [0, 0, 0] },
    torso: { rotation: [0, 0, 0] },
    rightShoulder: { rotation: [shoulder, 0, 0] },
    rightElbow: { rotation: [elbow, 0, 0] },
    rightHand: { rotation: [-shoulder - elbow, 0, 0] },
  };
}

function addLeg(
  hip: THREE.Group,
  knee: THREE.Group,
  ankle: THREE.Group,
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
  foot.position.set(0, 0, 0.04);
  ankle.add(foot);
  const cuff = cylinder(0.066, 0.065, 0.045, 7, bootColor);
  cuff.position.y = -0.025;
  knee.add(cuff);
}

// Swing feet clear the ground; stance feet stay level while travelling backward.
function strideLegPose(side: "left" | "right", phase: number, bounce: number): ModelPose {
  const angle = phase * TAU;
  const z = -Math.cos(angle) * 0.22;
  const lift = Math.max(0, Math.sin(angle)) ** 2 * 0.1;
  const y = 0.045 + lift - (0.54 + bounce);
  const upper = 0.26;
  const lower = 0.235;
  const knee = Math.acos(THREE.MathUtils.clamp(
    (y * y + z * z - upper * upper - lower * lower) / (2 * upper * lower), -1, 1,
  ));
  const hip = Math.atan2(-z, -y)
    - Math.atan2(lower * Math.sin(knee), upper + lower * Math.cos(knee));
  return {
    [`${side}Hip`]: { rotation: [hip, 0, 0] },
    [`${side}Knee`]: { rotation: [knee, 0, 0] },
    [`${side}Ankle`]: { rotation: [-hip - knee, 0, 0] },
  };
}
