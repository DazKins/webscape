import * as THREE from "three";
import Entity from "../entity/entity";
import PositionedEntityRenderer from "./positionedEntityRenderer";

export default class RendererError extends PositionedEntityRenderer {
  constructor(scene: THREE.Scene, entity: Entity) {
    super(scene, entity);

    const body = new THREE.Mesh(
      new THREE.BoxGeometry(0.8, 0.8, 0.8),
      new THREE.MeshPhongMaterial({ color: 0xff00aa })
    );
    body.position.set(0.5, 0.4, 0.5);
    this.mesh.add(body);

    const marker = new THREE.Mesh(
      new THREE.BoxGeometry(0.14, 0.55, 0.14),
      new THREE.MeshBasicMaterial({ color: 0xffffff })
    );
    marker.position.set(0.5, 0.92, 0.5);
    this.mesh.add(marker);

    const dot = new THREE.Mesh(
      new THREE.BoxGeometry(0.14, 0.14, 0.14),
      new THREE.MeshBasicMaterial({ color: 0xffffff })
    );
    dot.position.set(0.5, 0.18, 0.5);
    this.mesh.add(dot);

    this.addToScene();
  }
}
