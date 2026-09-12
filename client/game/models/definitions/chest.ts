import * as THREE from "three";
import { box, mesh, sphere } from "../primitives";
import { animation, createModelInstance, joint } from "../rig";
import type { ModelFactory } from "../types";

export const CHEST_OPEN_SECONDS = 0.4;

export const createChestModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "chest";
  const iron = 0x4c5358;
  const floor = box(0.72, 0.07, 0.54, 0x50331f);
  floor.position.y = 0.035;
  root.add(floor);
  // Separate walls leave an actual cavity when the lid opens.
  for (const z of [-0.245, 0.245]) {
    for (let row = 0; row < 3; row++) {
      const plank = box(0.72, 0.103, 0.05, row % 2 ? 0x996335 : 0xa77340);
      plank.position.set(0, 0.13 + row * 0.11, z);
      root.add(plank);
    }
  }
  for (const x of [-0.335, 0.335]) {
    const side = box(0.05, 0.34, 0.46, 0x8b592f);
    side.position.set(x, 0.24, 0);
    root.add(side);
  }
  const lidHinge = joint("lidHinge", root, [0, 0.41, -0.27]);
  // Extrude a semicircle along x: a barrel lid with closed end faces.
  const profile = new THREE.Shape();
  profile.moveTo(-0.285, 0);
  for (let i = 0; i <= 6; i++) {
    const angle = Math.PI - i * Math.PI / 6;
    profile.lineTo(Math.cos(angle) * 0.285, Math.sin(angle) * 0.285);
  }
  profile.closePath();
  const geometry = new THREE.ExtrudeGeometry(profile, { depth: 0.75, bevelEnabled: false, steps: 1 });
  geometry.translate(0, 0, -0.375);
  geometry.rotateY(Math.PI / 2);
  const lid = mesh(geometry, 0xb58046);
  lid.position.z = 0.27;
  lidHinge.add(lid);
  for (const x of [-0.24, 0.24]) {
    for (const z of [-0.277, 0.277]) {
      const band = box(0.065, 0.4, 0.018, iron, { metalness: 0.45 });
      band.position.set(x, 0.22, z);
      root.add(band);
      for (const y of [0.09, 0.34]) {
        const rivet = sphere(0.018, 6, 4, 0xb9aa80, { metalness: 0.5 });
        rivet.position.set(x, y, z);
        root.add(rivet);
      }
    }
    for (let segment = 0; segment < 6; segment++) {
      const angle = (segment + 0.5) * Math.PI / 6;
      const band = box(0.068, 0.023, 0.155, iron, { metalness: 0.45 });
      band.position.set(x, Math.sin(angle) * 0.28, 0.27 + Math.cos(angle) * 0.28);
      band.rotation.x = Math.PI / 2 - angle;
      lidHinge.add(band);
    }
  }
  const latch = box(0.095, 0.14, 0.035, 0xc6a352, { metalness: 0.55 });
  latch.position.set(0, -0.025, 0.563);
  lidHinge.add(latch);
  const keyhole = box(0.023, 0.045, 0.008, 0x343331);
  keyhole.position.set(0, -0.035, 0.584);
  lidHinge.add(keyhole);
  const open = animation("open", CHEST_OPEN_SECONDS, false, (phase) => ({
    lidHinge: { rotation: [-THREE.MathUtils.smoothstep(phase, 0, 1) * 1.9, 0, 0] },
  }));
  return createModelInstance(root, { lidHinge }, [open]);
};
