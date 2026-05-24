import * as THREE from "three";
import Entity from "../entity/entity";
import PositionedEntityRenderer from "./positionedEntityRenderer";

export default class RendererDoor extends PositionedEntityRenderer {
  constructor(scene: THREE.Scene, entity: Entity) {
    super(scene, entity);

    const door = new THREE.Mesh(
      new THREE.BoxGeometry(0.18, 1.25, 0.82),
      new THREE.MeshPhongMaterial({ color: 0x8a5a34 })
    );
    door.position.set(0.5, 0.625, 0.5);
    this.mesh.add(door);

    const handle = new THREE.Mesh(
      new THREE.SphereGeometry(0.05, 8, 8),
      new THREE.MeshPhongMaterial({ color: 0xd2b46d })
    );
    handle.position.set(0.62, 0.62, 0.78);
    this.mesh.add(handle);

    this.addToScene();
  }
}
