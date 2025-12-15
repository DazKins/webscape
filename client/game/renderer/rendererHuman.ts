import { lerp } from "../../math/lerp";
import Entity from "../entity/entity";
import EntityRenderer from "./renderer";
import * as THREE from "three";

export default class RendererHuman extends EntityRenderer {
  mesh: THREE.Group;

  constructor(scene: THREE.Scene, entity: Entity) {
    super(scene, entity);

    let metadataComponent = entity.getComponent("metadata");
    if (!metadataComponent) {
      metadataComponent = {}
    }
    const color = metadataComponent.color ?? "#00ff00";

    this.mesh = new THREE.Group();
    this.mesh.userData.entityId = entity.getId();

    const bodyGeometry = new THREE.CylinderGeometry(0.3, 0.3, 0.8, 16);
    const bodyMaterial = new THREE.MeshPhongMaterial({
      color: color,
    });
    const body = new THREE.Mesh(bodyGeometry, bodyMaterial);
    body.position.x = 0.5;
    body.position.y = 0.4;
    body.position.z = 0.5;
    this.mesh.add(body);

    const headGeometry = new THREE.SphereGeometry(0.25, 16, 16);
    const headMaterial = new THREE.MeshPhongMaterial({
      color: 0xffe0bd,
    });
    const head = new THREE.Mesh(headGeometry, headMaterial);
    head.position.x = 0.5;
    head.position.y = 0.9;
    head.position.z = 0.5;
    this.mesh.add(head);

    const positionComponent = this.entity.getComponent("position");
    this.mesh.position.set(positionComponent.x, 0, positionComponent.y);

    this.scene.add(this.mesh);
  }

  update() {
    const positionComponent = this.entity.getComponent("position");
    if (!positionComponent) {
      return;
    }

    this.mesh.position.set(
      lerp(this.mesh.position.x, positionComponent.x, 0.05),
      0,
      lerp(this.mesh.position.z, positionComponent.y, 0.05),
    );
  }

  getObject3D(): THREE.Object3D {
    return this.mesh;
  }

  onRemove() {
    this.scene.remove(this.mesh);
  }
}
