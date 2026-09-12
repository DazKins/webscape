import * as THREE from "three";
import { box, cone, cylinder, torus } from "../primitives";
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
  for (let index = 0; index < 5; index++) {
    const angle = index * Math.PI * 2 / 5;
    const rootWood = cylinder(0.055, 0.09, 0.24, 5, 0x68442c);
    rootWood.position.set(Math.sin(angle) * 0.11, 0.1, Math.cos(angle) * 0.11);
    rootWood.rotation.set(Math.cos(angle) * 0.6, 0, -Math.sin(angle) * 0.6);
    standingTree.add(rootWood);
  }

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
  for (const radius of [0.045, 0.09]) {
    const ring = torus(radius, 0.004, 4, 8, 0x966537);
    ring.rotation.x = Math.PI / 2;
    ring.position.y = 0.234;
    stump.add(ring);
  }

  setTreeDamageStage(root, options.damageStage ?? 0);
  const hit = animation("hit", TREE_HIT_ANIMATION_SECONDS, false, (phase) => {
    const strength = (1 - phase) ** 2 * TREE_HIT_SHAKE_RADIANS;
    return {
      treeShake: {
        rotation: [
          Math.sin(phase * Math.PI * 4) * strength * 0.38,
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
  for (let tier = 0; tier < 3; tier++) {
    const leaves = cone((0.48 - tier * 0.115) * fullness, (0.75 - tier * 0.1) * fullness, 8, colors[tier === 2 ? 1 : 0]);
    // Alternating, slightly irregular skirts break up the stacked-cone silhouette.
    const positions = leaves.geometry.getAttribute("position");
    for (let index = 0; index < positions.count; index++) {
      const x = positions.getX(index);
      const z = positions.getZ(index);
      if (positions.getY(index) < 0) {
        const variation = Math.sin(Math.atan2(z, x) * 3 + tier) * 0.035;
        positions.setY(index, positions.getY(index) + variation);
      }
    }
    leaves.geometry.computeVertexNormals();
    leaves.position.set(tier * 0.012, 0.92 + tier * 0.29 - stage * 0.008, 0);
    leaves.rotation.y = tier * 0.4;
    canopy.add(leaves);
  }
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
