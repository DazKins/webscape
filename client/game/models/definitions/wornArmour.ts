import * as THREE from "three";
import { box, cylinder, sphere, taperedBox, torus } from "../primitives";
import { createModelInstance, joint } from "../rig";
import type { ModelFactory } from "../types";

const iron = 0x818d98;
const trim = 0xb8c2ca;
const metal = { metalness: 0.5, roughness: 0.65 };

function place(parent: THREE.Group, mesh: THREE.Object3D, x: number, y: number, z: number) {
  mesh.position.set(x, y, z);
  parent.add(mesh);
  return mesh;
}

// Each detachable group is authored relative to its human joint pivot.
// The controller attaches these to the actual rig so bends and breathing agree.
export const createWornChainmailModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "chainmailChestplate";
  place(root, taperedBox(0.4, 0.245, 0.32, 0.205, 0.32, iron, metal), 0, 0.16, 0);
  const collar = torus(0.083, 0.022, 6, 10, trim, metal);
  collar.rotation.x = Math.PI / 2;
  place(root, collar, 0, 0.325, 0.006);
  // Low-poly links follow the taper on both faces of the mail shirt.
  for (const face of [-1, 1]) {
    for (let row = 0; row < 5; row++) {
      const y = 0.09 + row * 0.045;
      for (let col = 0; col < 6; col++) {
        place(root, torus(0.018, 0.0045, 4, 6, trim, metal),
          -0.132 + col * 0.05 + (row % 2) * 0.01, y, face * (0.108 + y * 0.0625));
      }
    }
  }
  place(root, box(0.34, 0.05, 0.23, 0x503a2a), 0, 0.035, 0);
  place(root, box(0.067, 0.056, 0.02, 0xc7a05a, metal), 0, 0.035, 0.125);
  const skirt = joint("mailSkirt", root);
  place(skirt, taperedBox(0.32, 0.22, 0.35, 0.245, 0.18, iron, metal), 0, 0, 0);
  for (const side of ["left", "right"]) {
    const sleeve = joint(`${side}MailSleeve`, root);
    place(sleeve, cylinder(0.084, 0.077, 0.17, 8, iron, metal), 0, -0.075, 0);
    place(sleeve, cylinder(0.079, 0.079, 0.025, 8, trim, metal), 0, -0.155, 0);
  }
  return createModelInstance(root);
};

export const createWornIronLeggingsModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "ironLeggings";
  for (const side of ["left", "right"]) {
    const thigh = joint(`${side}Cuisses`, root);
    place(thigh, cylinder(0.087, 0.074, 0.25, 8, iron, metal), 0, -0.125, 0);
    place(thigh, box(0.018, 0.19, 0.012, trim, metal), 0, -0.13, 0.083);
    const shin = joint(`${side}Greave`, root);
    place(shin, cylinder(0.071, 0.059, 0.225, 8, iron, metal), 0, -0.1125, 0);
    const knee = sphere(0.085, 8, 6, trim, metal);
    knee.scale.set(1, 0.85, 0.55);
    place(shin, knee, 0, -0.005, 0.047);
  }
  return createModelInstance(root);
};

export const createWornLeatherBootsModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "leatherBoots";
  for (const side of ["left", "right"]) {
    const shaft = joint(`${side}BootShaft`, root);
    // The shaft follows the shin; the sole follows the independently posed foot.
    place(shaft, cylinder(0.077, 0.065, 0.21, 8, 0x70482f), 0, -0.125, 0);
    const bootTop = sphere(0.076, 8, 6, 0x70482f);
    bootTop.scale.y = 0.65;
    place(shaft, bootTop, 0, -0.005, 0);
    place(shaft, cylinder(0.082, 0.081, 0.035, 8, 0xb18453), 0, -0.015, 0);
    for (const y of [-0.08, -0.13, -0.18]) {
      place(shaft, box(0.055, 0.012, 0.012, 0x392a21), 0, y, 0.075);
    }
    const foot = joint(`${side}BootFoot`, root);
    place(foot, box(0.142, 0.025, 0.225, 0x392a21), 0, -0.0275, 0.043);
    place(foot, taperedBox(0.128, 0.19, 0.14, 0.22, 0.065, 0x855738), 0, 0.01, 0.04);
    place(foot, box(0.145, 0.018, 0.032, 0xb18453), 0, 0.044, 0.023);
  }
  return createModelInstance(root);
};
