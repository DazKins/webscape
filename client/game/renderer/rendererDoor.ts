import * as THREE from "three";
import Entity from "../entity/entity";
import PositionedEntityRenderer from "./positionedEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";

export default class RendererDoor extends PositionedEntityRenderer {
  private orientationGroup: THREE.Group;
  private doorPivot: THREE.Group;

  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler
  ) {
    super(scene, entity, terrainHeightSampler);

    this.orientationGroup = new THREE.Group();
    this.orientationGroup.position.set(0.5, 0, 0.5);
    this.mesh.add(this.orientationGroup);

    this.doorPivot = new THREE.Group();
    this.doorPivot.position.set(-0.09, 0, -0.41);
    this.orientationGroup.add(this.doorPivot);

    const door = new THREE.Mesh(
      new THREE.BoxGeometry(0.18, 1.25, 0.82),
      new THREE.MeshPhongMaterial({ color: 0x8a5a34 })
    );
    door.position.set(0.09, 0.625, 0.41);
    this.doorPivot.add(door);

    const handle = new THREE.Mesh(
      new THREE.SphereGeometry(0.05, 8, 8),
      new THREE.MeshPhongMaterial({ color: 0xd2b46d })
    );
    handle.position.set(0.21, 0.62, 0.69);
    this.doorPivot.add(handle);

    this.addToScene();
  }

  update(deltaSeconds: number) {
    super.update(deltaSeconds);

    const openable = this.entity.getComponent("openable");
    const isOpen = Boolean(openable?.isOpen);
    const targetRotation = isOpen ? -Math.PI / 2 : 0;
    const rotationStep = Math.min(1, deltaSeconds * 12);
    this.orientationGroup.rotation.y = this.getOrientationRotation();
    this.doorPivot.rotation.y += (targetRotation - this.doorPivot.rotation.y) * rotationStep;
  }

  private getOrientationRotation() {
    const renderable = this.entity.getComponent("renderable");
    switch (renderable?.orientation) {
      case "east":
        return Math.PI / 2;
      case "south":
        return Math.PI;
      case "west":
        return -Math.PI / 2;
      case "north":
      default:
        return 0;
    }
  }
}
