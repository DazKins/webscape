import * as THREE from "three";
import { createModel, type ModelInstance } from "../models";

export default class RendererArrow {
  private readonly root = new THREE.Group();
  private readonly start: THREE.Vector3;
  private readonly model: ModelInstance;

  constructor(parent: THREE.Object3D, start: THREE.Vector3) {
    this.start = start.clone();
    // Use the inventory arrow's silhouette and fletching in flight as well.
    this.model = createModel("arrow");
    const center = new THREE.Box3().setFromObject(this.model.root).getCenter(new THREE.Vector3());
    this.model.root.position.sub(center);
    this.root.add(this.model.root);
    this.root.position.copy(start);
    parent.add(this.root);
  }

  update(progress: number, target: THREE.Vector3) {
    const phase = THREE.MathUtils.clamp(progress, 0, 1);
    const position = new THREE.Vector3().lerpVectors(this.start, target, phase);
    position.y += Math.sin(phase * Math.PI) * 0.12;
    this.root.position.copy(position);
    // Follow the tangent of the arc, including at impact where position=target.
    const direction = target.clone().sub(this.start);
    direction.y += Math.cos(phase * Math.PI) * Math.PI * 0.12;
    if (direction.lengthSq() > 0.000001) this.root.lookAt(position.clone().add(direction));
  }

  dispose() {
    this.model.dispose();
    this.root.removeFromParent();
  }
}
