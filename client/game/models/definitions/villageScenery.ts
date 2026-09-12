import * as THREE from "three";
import { box, cone, cylinder, sphere } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

const timber = 0x70452c;
const brass = 0xba914c;

function place(root: THREE.Group, mesh: THREE.Object3D, x: number, y: number, z: number) {
  mesh.position.set(x, y, z);
  root.add(mesh);
  return mesh;
}

export const createLanternModel: ModelFactory = () => {
  const root = new THREE.Group();
  place(root, box(0.34, 0.16, 0.34, 0x737b7b), 0, 0.08, 0);
  place(root, cylinder(0.055, 0.08, 1.6, 6, timber), 0, 0.88, 0);
  place(root, box(0.34, 0.08, 0.34, brass), 0, 1.68, 0);
  place(root, box(0.23, 0.3, 0.23, 0xffd68a, { emissive: 0xb47624 }), 0, 1.87, 0);
  for (const x of [-0.14, 0.14]) for (const z of [-0.14, 0.14]) {
    place(root, box(0.035, 0.34, 0.035, brass), x, 1.88, z);
  }
  place(root, cone(0.28, 0.22, 4, 0x365d61), 0, 2.14, 0);
  return createModelInstance(root);
};

export const createBenchModel: ModelFactory = () => {
  const root = new THREE.Group();
  for (const x of [-0.32, 0.32]) {
    place(root, box(0.12, 0.48, 0.48, timber), x, 0.24, 0);
    place(root, box(0.1, 0.8, 0.1, timber), x, 0.4, -0.2);
  }
  place(root, box(0.95, 0.1, 0.5, 0xa97544), 0, 0.5, 0);
  place(root, box(0.95, 0.25, 0.08, 0xa97544), 0, 0.76, -0.2);
  return createModelInstance(root);
};

export const createTavernTableModel: ModelFactory = () => {
  const root = new THREE.Group();
  place(root, cylinder(0.44, 0.44, 0.1, 10, 0x9b6539), 0, 0.65, 0);
  place(root, cylinder(0.09, 0.14, 0.6, 6, timber), 0, 0.3, 0);
  place(root, box(0.64, 0.1, 0.13, timber), 0, 0.05, 0);
  place(root, box(0.13, 0.1, 0.64, timber), 0, 0.05, 0);
  place(root, cylinder(0.095, 0.08, 0.15, 8, 0xd1b68b), 0.2, 0.77, 0.08);
  place(root, sphere(0.11, 8, 6, 0xb63c36), -0.13, 0.77, -0.12);
  return createModelInstance(root);
};

export const createBookcaseModel: ModelFactory = () => {
  const root = new THREE.Group();
  place(root, box(0.92, 1.35, 0.08, timber), 0, 0.675, -0.2);
  for (const x of [-0.44, 0.44]) place(root, box(0.08, 1.4, 0.46, timber), x, 0.7, 0);
  for (const y of [0.08, 0.48, 0.9, 1.36]) place(root, box(0.94, 0.07, 0.46, 0x9b6539), 0, y, 0);
  const colors = [0x486b77, 0x9b4548, 0x70864c, 0xb19156, 0x65517c];
  for (let shelf = 0; shelf < 3; shelf++) for (let book = 0; book < 5; book++) {
    const height = 0.23 + (book % 3) * 0.035;
    const x = -0.32 + book * 0.155;
    const y = [0.115, 0.515, 0.935][shelf] + height / 2;
    place(root, box(0.12, height, 0.28, colors[(book + shelf) % colors.length]), x, y, 0.04);
    place(root, box(0.09, 0.025, 0.012, brass), x, y + 0.055, 0.185);
  }
  return createModelInstance(root);
};

export const createShopCounterModel: ModelFactory = () => {
  const root = new THREE.Group();
  place(root, box(0.86, 0.65, 0.6, timber), 0, 0.325, 0);
  place(root, box(0.98, 0.1, 0.72, 0xbd8d55), 0, 0.7, 0);
  place(root, box(0.5, 0.015, 0.62, 0x466a6e), 0, 0.76, 0);
  place(root, cylinder(0.11, 0.08, 0.2, 8, brass), 0.2, 0.86, 0.02);
  place(root, box(0.24, 0.07, 0.23, 0xe0cd99), -0.17, 0.8, 0.06);
  return createModelInstance(root);
};

export const createArcheryTargetModel: ModelFactory = () => {
  const root = new THREE.Group();
  for (const x of [-0.26, 0.26]) place(root, box(0.09, 1.25, 0.12, timber), x, 0.625, 0);
  const rings: [number, number][] = [[0.43, 0xc2a970], [0.32, 0xe9dcc1], [0.22, 0x475e6a], [0.12, 0xc4513f], [0.05, 0xe9bd54]];
  rings.forEach(([radius, color], i) => {
    const disk = cylinder(radius, radius, 0.04, 16, color);
    disk.rotation.x = Math.PI / 2;
    place(root, disk, 0, 1, 0.08 + i * 0.025);
  });
  return createModelInstance(root);
};

export const createFountainModel: ModelFactory = () => {
  const root = new THREE.Group();
  place(root, cylinder(0.94, 0.98, 0.18, 12, 0x7c8589), 0, 0.09, 0);
  place(root, cylinder(0.84, 0.84, 0.09, 12, 0x469cac, { emissive: 0x123d46 }), 0, 0.22, 0);
  for (let i = 0; i < 12; i++) {
    const angle = i * Math.PI / 6;
    const rim = box(0.48, 0.3, 0.16, 0xa0a49c);
    place(root, rim, Math.sin(angle) * 0.9, 0.26, Math.cos(angle) * 0.9);
    rim.rotation.y = angle;
  }
  place(root, cylinder(0.12, 0.24, 0.9, 8, 0xa0a49c), 0, 0.65, 0);
  place(root, cylinder(0.52, 0.25, 0.18, 10, 0xa0a49c), 0, 1.06, 0);
  place(root, cylinder(0.44, 0.44, 0.03, 10, 0x66b9c9), 0, 1.16, 0);
  place(root, sphere(0.16, 8, 6, brass), 0, 1.33, 0);
  for (const angle of [0, Math.PI / 2, Math.PI, Math.PI * 1.5]) {
    place(root, cylinder(0.025, 0.04, 0.75, 5, 0x87d3df), Math.sin(angle) * 0.42, 0.7, Math.cos(angle) * 0.42);
  }
  return createModelInstance(root);
};
