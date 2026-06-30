import * as THREE from "three";
import Entity from "../entity/entity";
import PositionedEntityRenderer from "./positionedEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";

export default class RendererTree extends PositionedEntityRenderer {
  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler
  ) {
    super(scene, entity, terrainHeightSampler);

    const trunk = new THREE.Mesh(
      new THREE.CylinderGeometry(0.12, 0.16, 0.8, 8),
      new THREE.MeshPhongMaterial({ color: 0x7a4f2a })
    );
    trunk.position.set(0.5, 0.4, 0.5);
    this.mesh.add(trunk);

    const leaves = new THREE.Mesh(
      new THREE.ConeGeometry(0.48, 1.0, 10),
      new THREE.MeshPhongMaterial({ color: 0x2f6b3b })
    );
    leaves.position.set(0.5, 1.15, 0.5);
    this.mesh.add(leaves);

    this.addToScene();
  }
}
