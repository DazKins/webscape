import * as THREE from "three";
import { box, cone, cylinder } from "../primitives";
import { animation, createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

const DAMAGE_STAGE_COUNT = 4;
export const TREE_HIT_ANIMATION_SECONDS = 0.28;
const TREE_HIT_SHAKE_RADIANS = THREE.MathUtils.degToRad(4.5);
export function setTreeDamageStage(root: THREE.Object3D, requestedStage: number) {
  const stage = THREE.MathUtils.clamp(Math.floor(requestedStage), 0, DAMAGE_STAGE_COUNT);
  for (let index = 0; index <= DAMAGE_STAGE_COUNT; index += 1) {
    const canopy = root.getObjectByName(`canopyDamage${index}`);
    if (canopy) canopy.visible = stage === index;

    if (index > 0) {
      const chips = root.getObjectByName(`woodChips${index}`);
      if (chips) chips.visible = stage >= index;

      const marks = root.getObjectByName(`chopMarks${index}`);
      if (marks) marks.visible = stage >= index;
    }
  }
}

export const createTreeModel: ModelFactory = (options = {}) => {
  const root = new THREE.Group();
  root.name = "tree";

  const shakePivot = new THREE.Group();
  shakePivot.name = "treeShake";
  root.add(shakePivot);

  const standingTree = new THREE.Group();
  standingTree.name = "standingTree";
  shakePivot.add(standingTree);

  const trunk = cylinder(0.12, 0.16, 0.8, 8, 0x7a4f2a);
  trunk.name = "treeTrunk";
  trunk.position.y = 0.4;
  standingTree.add(trunk);

  const canopyColors = [
    [0x2f6b3b, 0x3d7b48],
    [0x396b39, 0x4a7945],
    [0x41683a, 0x527541],
    [0x4c6438, 0x5d713f],
    [0x586037, 0x686b3c],
  ] as const;
  for (let stage = 0; stage <= DAMAGE_STAGE_COUNT; stage += 1) {
    const canopy = createDamagedCanopy(stage, canopyColors[stage]);
    canopy.name = `canopyDamage${stage}`;
    canopy.visible = stage === 0;
    standingTree.add(canopy);
  }

  for (let stage = 1; stage <= DAMAGE_STAGE_COUNT; stage += 1) {
    const marks = createChopMarks(stage);
    marks.name = `chopMarks${stage}`;
    marks.visible = false;
    standingTree.add(marks);

    const chips = createWoodChips(stage);
    chips.name = `woodChips${stage}`;
    chips.visible = false;
    standingTree.add(chips);
  }

  const stump = new THREE.Group();
  stump.name = "stump";
  stump.visible = false;
  shakePivot.add(stump);

  const stumpWood = cylinder(0.14, 0.18, 0.22, 8, 0x7a4f2a);
  stumpWood.position.y = 0.11;
  stump.add(stumpWood);

  const cutSurface = cylinder(0.125, 0.125, 0.012, 8, 0xc39258);
  cutSurface.position.y = 0.226;
  stump.add(cutSurface);

  setTreeDamageStage(root, options.damageStage ?? 0);
  const hit = animation("hit", TREE_HIT_ANIMATION_SECONDS, false, (phase) => {
    const strength = (1 - phase) * TREE_HIT_SHAKE_RADIANS;
    return {
      treeShake: {
        rotation: [
          Math.sin(phase * Math.PI * 4 + Math.PI / 3) * strength * 0.38,
          0,
          Math.sin(phase * Math.PI * 6) * strength,
        ],
      },
    };
  });
  return createModelInstance(root, { treeShake: shakePivot }, [hit]);
};

function createDamagedCanopy(
  stage: number,
  colors: readonly [THREE.ColorRepresentation, THREE.ColorRepresentation],
) {
  const canopy = new THREE.Group();
  const fullness = 1 - stage * 0.035;
  const lowerLeaves = cone(0.48 * fullness, 0.78 * fullness, 9, colors[0]);
  lowerLeaves.position.y = 0.95;
  canopy.add(lowerLeaves);
  const upperLeaves = cone(0.36 * fullness, 0.72 * fullness, 9, colors[1]);
  upperLeaves.position.y = 1.34 - stage * 0.012;
  canopy.add(upperLeaves);
  return canopy;
}

function createChopMarks(stage: number) {
  const marks = new THREE.Group();
  const surfaceRadius = 0.151;
  const y = 0.3 + stage * 0.04;
  for (let side = 0; side < 4; side += 1) {
    const angle = side * Math.PI / 2;
    const mark = box(0.09, 0.028, 0.009, stage % 2 === 0 ? 0xc58d4d : 0xd9ad70);
    mark.position.set(
      Math.sin(angle) * surfaceRadius,
      y,
      Math.cos(angle) * surfaceRadius,
    );
    mark.rotation.y = angle;
    mark.rotation.z = (stage + side) % 2 === 0 ? -0.28 : 0.28;
    marks.add(mark);
  }
  return marks;
}

function createWoodChips(stage: number) {
  const chips = new THREE.Group();
  const angles = [0.45, 2.15, 3.75, 5.35];
  const angle = angles[stage - 1];
  for (let index = 0; index < 2; index += 1) {
    const chip = box(
      0.045 - index * 0.009,
      0.012,
      0.022 + index * 0.008,
      index === 0 ? 0xd0a064 : 0xb9793f,
    );
    const distance = 0.19 + index * 0.055;
    chip.position.set(
      Math.sin(angle + index * 0.28) * distance,
      0.008,
      Math.cos(angle + index * 0.28) * distance,
    );
    chip.rotation.y = angle + index * 0.7;
    chips.add(chip);
  }

  return chips;
}
