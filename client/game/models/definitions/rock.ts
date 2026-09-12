import * as THREE from "three";
import { dodecahedron } from "../primitives";
import { createModelInstance } from "../rig";
import type { ModelFactory } from "../types";

export const createRockModel: ModelFactory = () => {
  const root = new THREE.Group();
  root.name = "rock";
  const rock = dodecahedron(0.42, 0x8b9296);
  rock.scale.set(1.1, 0.65, 0.9);
  rock.rotation.set(0.08, 0.31, -0.05);
  rock.position.y = 0.25;
  // Flatten the base so this reads as an embedded boulder, not a hovering gem.
  const positions = rock.geometry.getAttribute("position");
  for (let index = 0; index < positions.count; index++) {
    const x = positions.getX(index);
    const z = positions.getZ(index);
    positions.setXYZ(index, x * (1 + z * 0.2), Math.max(-0.32, positions.getY(index)), z);
  }
  rock.geometry.computeVertexNormals();
  root.add(rock);
  const bounds = new THREE.Box3().setFromObject(rock);
  rock.position.y -= bounds.min.y;
  return createModelInstance(root);
};
