import * as THREE from "three";
import Entity from "../entity/entity";
import PositionedEntityRenderer from "./positionedEntityRenderer";

export default class RendererRewardDrop extends PositionedEntityRenderer {
  constructor(scene: THREE.Scene, entity: Entity) {
    super(scene, entity);

    const bag = new THREE.Mesh(
      new THREE.SphereGeometry(0.28, 16, 12),
      new THREE.MeshPhongMaterial({ color: 0xc58a3a })
    );
    bag.scale.set(1, 0.75, 0.85);
    bag.position.set(0.5, 0.28, 0.5);
    this.mesh.add(bag);

    const tie = new THREE.Mesh(
      new THREE.TorusGeometry(0.15, 0.025, 8, 16),
      new THREE.MeshPhongMaterial({ color: 0xf0d38a })
    );
    tie.rotation.x = Math.PI / 2;
    tie.position.set(0.5, 0.52, 0.5);
    this.mesh.add(tie);

    const glint = new THREE.Mesh(
      new THREE.BoxGeometry(0.12, 0.12, 0.04),
      new THREE.MeshBasicMaterial({ color: 0xfff0a8 })
    );
    glint.position.set(0.36, 0.38, 0.77);
    this.mesh.add(glint);

    this.addToScene();
  }
}
