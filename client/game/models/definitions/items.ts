import * as THREE from "three";
import { box, cone, cylinder, dodecahedron, sphere, torus } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

// Small inventory objects share the same low-poly palette as the equipment.
function itemModel(name: string, build: (root: THREE.Group) => void): ModelFactory {
  return () => {
    const root = new THREE.Group();
    root.name = name;
    build(root);
    const bounds = new THREE.Box3().setFromObject(root);
    const center = bounds.getCenter(new THREE.Vector3());
    for (const child of root.children) {
      child.position.add(new THREE.Vector3(-center.x, -bounds.min.y, -center.z));
    }
    return createModelInstance(root);
  };
}

function place(root: THREE.Group, object: THREE.Object3D, x: number, y: number, z: number) {
  object.position.set(x, y, z);
  root.add(object);
  return object;
}

export const createHealthPotionModel = itemModel("healthPotion", (root) => {
  const bottle = sphere(0.12, 8, 6, 0x991e46);
  bottle.scale.y = 1.2;
  place(root, bottle, 0, 0.14, 0);
  place(root, cylinder(0.05, 0.065, 0.09, 8, 0xd4c6dc), 0, 0.29, 0);
  place(root, cylinder(0.052, 0.05, 0.045, 8, 0x9c7244), 0, 0.35, 0);
  place(root, box(0.12, 0.09, 0.015, 0xefdfb7), 0, 0.14, 0.114);
  place(root, box(0.07, 0.022, 0.018, 0xa62f43), 0, 0.14, 0.125);
  place(root, box(0.022, 0.065, 0.018, 0xa62f43), 0, 0.14, 0.125);
});

export const createBreadModel = itemModel("bread", (root) => {
  const loaf = sphere(0.18, 10, 6, 0xc38943);
  loaf.scale.set(0.75, 0.6, 1.5);
  place(root, loaf, 0, 0.09, 0);
  for (const z of [-0.12, 0, 0.12]) {
    const score = box(0.16, 0.014, 0.025, 0xf0c47a);
    score.rotation.y = -0.3;
    place(root, score, 0, 0.191 - Math.abs(z) * 0.15, z);
  }
});

export const createAppleModel = itemModel("apple", (root) => {
  place(root, sphere(0.13, 9, 7, 0xb33429), 0, 0.13, 0);
  place(root, cylinder(0.012, 0.017, 0.085, 5, 0x65452c), 0, 0.28, 0);
  const leaf = sphere(0.06, 5, 3, 0x58853b);
  leaf.scale.set(1, 0.15, 0.5);
  leaf.rotation.z = 0.4;
  place(root, leaf, 0.05, 0.29, 0);
});

export const createIronOreModel = itemModel("ironOre", (root) => {
  const rock = dodecahedron(0.2, 0x565761);
  rock.scale.set(1, 0.72, 0.9);
  place(root, rock, 0, 0.12, 0);
  for (const [x, z] of [[-0.08, 0.05], [0.09, -0.04], [0, 0.1]]) {
    place(root, dodecahedron(0.065, 0xc1a79a, { metalness: 0.6 }), x, 0.23, z);
  }
});

export const createStoneItemModel = itemModel("stone", (root) => {
  const stone = dodecahedron(0.2, 0x929995);
  stone.scale.set(1.2, 0.6, 0.9);
  root.add(stone);
});

export const createWoodModel = itemModel("wood", (root) => {
  for (const x of [-0.07, 0.07]) {
    place(root, box(0.11, 0.055, 0.5, 0xb0824e), x, 0.035, 0);
    place(root, box(0.012, 0.008, 0.45, 0x7b512e), x + 0.02, 0.065, 0);
  }
});

export const createLogsModel = itemModel("logs", (root) => {
  for (const [x, y] of [[-0.09, 0.09], [0.09, 0.09], [0, 0.23]]) {
    const log = cylinder(0.09, 0.095, 0.46, 8, 0x71462d);
    log.rotation.x = Math.PI / 2;
    place(root, log, x, y, 0);
    for (const z of [-0.235, 0.235]) {
      const end = cylinder(0.075, 0.075, 0.01, 8, 0xd0a067);
      end.rotation.x = Math.PI / 2;
      place(root, end, x, y, z);
      place(root, torus(0.041, 0.005, 4, 8, 0x95633d), x, y, z * 1.03);
    }
  }
});

export const createArrowModel = itemModel("arrow", (root) => {
  const shaft = cylinder(0.014, 0.014, 0.65, 7, 0x8a5a32);
  shaft.rotation.x = Math.PI / 2;
  root.add(shaft);
  const point = cone(0.045, 0.13, 5, 0xcbd1d4, { metalness: 0.45 });
  point.rotation.x = Math.PI / 2;
  place(root, point, 0, 0, 0.39);
  for (const side of [-1, 1]) {
    const feather = box(0.075, 0.01, 0.13, 0xc44b3d);
    feather.rotation.z = side * 0.55;
    place(root, feather, 0, 0, -0.27);
  }
});

export const createFishModel = itemModel("fish", (root) => {
  const body = sphere(0.14, 10, 6, 0x79b1b8);
  body.scale.set(1.7, 0.4, 0.7);
  root.add(body);
  const tail = cone(0.105, 0.15, 3, 0x507f90);
  tail.rotation.z = Math.PI / 2;
  tail.scale.z = 0.35;
  place(root, tail, -0.25, 0, 0);
  place(root, sphere(0.018, 6, 4, 0x172e33), 0.14, 0.052, 0.037);
  const fin = cone(0.055, 0.1, 3, 0x507f90);
  fin.rotation.x = Math.PI / 2;
  place(root, fin, 0, 0.01, 0.09);
});

export const createMysteriousKeyModel = itemModel("mysteriousKey", (root) => {
  const ring = torus(0.065, 0.019, 6, 10, 0xc7a34f, { metalness: 0.6 });
  ring.rotation.x = Math.PI / 2;
  place(root, ring, 0, 0.02, -0.13);
  place(root, box(0.03, 0.03, 0.25, 0xc7a34f), 0, 0.02, 0.045);
  for (const z of [0.09, 0.15]) place(root, box(0.08, 0.03, 0.028, 0xc7a34f), 0.03, 0.02, z);
});

export const createAncientScrollModel = itemModel("ancientScroll", (root) => {
  place(root, box(0.25, 0.02, 0.33, 0xe0c994), 0, 0.045, 0);
  for (const z of [-0.16, 0.16]) {
    const roll = cylinder(0.045, 0.045, 0.29, 8, 0xebd5a4);
    roll.rotation.z = Math.PI / 2;
    place(root, roll, 0, 0.045, z);
  }
  for (const z of [-0.07, -0.025, 0.02, 0.065]) place(root, box(0.15, 0.006, 0.01, 0x876542), 0, 0.058, z);
  place(root, cylinder(0.035, 0.035, 0.01, 8, 0x9c3834), 0.05, 0.065, 0.08);
});

export const createChainmailChestplateModel = itemModel("chainmailChestplate", (root) => {
  place(root, box(0.34, 0.1, 0.4, 0x8b969e, { metalness: 0.5 }), 0, 0.06, 0);
  for (const x of [-0.23, 0.23]) place(root, box(0.14, 0.09, 0.17, 0x737f89), x, 0.06, -0.1);
  for (let row = 0; row < 5; row++) {
    for (let col = 0; col < 5; col++) {
      const ring = torus(0.024, 0.006, 4, 6, 0xc0c6cb, { metalness: 0.6 });
      ring.rotation.x = Math.PI / 2;
      place(root, ring, -0.12 + col * 0.06, 0.115, -0.12 + row * 0.065);
    }
  }
});

export const createIronLeggingsModel = itemModel("ironLeggings", (root) => {
  place(root, box(0.32, 0.1, 0.1, 0x5d6772), 0, 0.06, -0.2);
  for (const x of [-0.085, 0.085]) {
    place(root, box(0.14, 0.09, 0.45, 0x8c969f, { metalness: 0.5 }), x, 0.06, 0.025);
    place(root, box(0.15, 0.035, 0.1, 0xb6bec4), x, 0.12, 0.065);
  }
});

export const createLeatherBootsModel = itemModel("leatherBoots", (root) => {
  for (const x of [-0.095, 0.095]) {
    place(root, box(0.14, 0.05, 0.27, 0x392a21), x, 0.03, 0.035);
    place(root, box(0.135, 0.11, 0.24, 0x855738), x, 0.095, 0.04);
    place(root, box(0.125, 0.22, 0.13, 0x70482f), x, 0.18, -0.015);
    place(root, box(0.14, 0.035, 0.145, 0xb18453), x, 0.28, -0.015);
  }
});

export const createWoodenShieldModel = itemModel("woodenShield", (root) => {
  place(root, cylinder(0.26, 0.26, 0.06, 10, 0x92704a), 0, 0.04, 0);
  const rim = torus(0.25, 0.022, 6, 10, 0x8b969c, { metalness: 0.5 });
  rim.rotation.x = Math.PI / 2;
  place(root, rim, 0, 0.06, 0);
  place(root, cylinder(0.06, 0.075, 0.04, 8, 0xb1bbc0), 0, 0.09, 0);
  for (const x of [-0.11, 0.11]) place(root, box(0.012, 0.006, 0.43, 0x62472e), x, 0.074, 0);
});

// Visible diagnostic for future content without a registered item model.
export const createUnknownItemModel = itemModel("unknownItem", (root) => {
  root.add(dodecahedron(0.16, 0xc675db));
});
