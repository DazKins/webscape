import * as THREE from "three";
import Entity from "../entity/entity";

export default class EntityRenderer {
  scene: THREE.Scene;
  entity: Entity;

  constructor(scene: THREE.Scene, entity: Entity) {
    this.scene = scene;
    this.entity = entity;
  }

  update() {
    throw Error("update not implemented");
  }

  getObject3D(): THREE.Object3D | null {
    return null;
  }

  onRemove() {
    throw Error("onRemove not implemented");
  }
}
