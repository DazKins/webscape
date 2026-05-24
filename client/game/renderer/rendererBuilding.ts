import * as THREE from "three";
import Entity from "../entity/entity";
import PositionedEntityRenderer from "./positionedEntityRenderer";

export default class RendererBuilding extends PositionedEntityRenderer {
  constructor(scene: THREE.Scene, entity: Entity) {
    super(scene, entity);

    const size = this.getFootprintSize();

    const body = new THREE.Mesh(
      new THREE.BoxGeometry(Math.max(0.8, size.width * 0.9), 1.2, Math.max(0.8, size.height * 0.9)),
      new THREE.MeshPhongMaterial({ color: 0xb8ab88 })
    );
    body.position.set(size.width / 2, 0.6, size.height / 2);
    this.mesh.add(body);

    const roof = new THREE.Mesh(
      new THREE.ConeGeometry(Math.max(size.width, size.height) * 0.7, 0.55, 4),
      new THREE.MeshPhongMaterial({ color: 0x7f4231 })
    );
    roof.rotation.y = Math.PI / 4;
    roof.position.set(size.width / 2, 1.475, size.height / 2);
    this.mesh.add(roof);

    this.addToScene();
  }
}
