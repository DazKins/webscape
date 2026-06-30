import * as THREE from "three";
import Entity from "../entity/entity";
import PositionedEntityRenderer from "./positionedEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";

export default class RendererChest extends PositionedEntityRenderer {
  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler
  ) {
    super(scene, entity, terrainHeightSampler);

    const body = new THREE.Mesh(
      new THREE.BoxGeometry(0.72, 0.42, 0.54),
      new THREE.MeshPhongMaterial({ color: 0x9a5d2e })
    );
    body.position.set(0.5, 0.21, 0.5);
    this.mesh.add(body);

    const lid = new THREE.Mesh(
      new THREE.BoxGeometry(0.78, 0.16, 0.6),
      new THREE.MeshPhongMaterial({ color: 0xb47b37 })
    );
    lid.position.set(0.5, 0.5, 0.5);
    this.mesh.add(lid);

    const latch = new THREE.Mesh(
      new THREE.BoxGeometry(0.12, 0.12, 0.04),
      new THREE.MeshPhongMaterial({ color: 0xd2b46d })
    );
    latch.position.set(0.5, 0.38, 0.79);
    this.mesh.add(latch);

    this.addToScene();
  }
}
