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
  rock.position.y = 0.28;
  root.add(rock);
  return createModelInstance(root);
};
