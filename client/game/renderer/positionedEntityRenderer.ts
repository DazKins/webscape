import * as THREE from "three";
import Entity from "../entity/entity";
import EntityRenderer, { type TerrainHeightSampler } from "./renderer";

export default class PositionedEntityRenderer extends EntityRenderer {
  mesh: THREE.Group;

  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler?: TerrainHeightSampler
  ) {
    super(scene, entity, terrainHeightSampler);

    this.mesh = new THREE.Group();
    this.mesh.userData.entityId = entity.getId();
    this.updatePosition();
  }

  update(_deltaSeconds: number) {
    this.updatePosition();
  }

  getObject3D(): THREE.Object3D {
    return this.mesh;
  }

  onRemove() {
    this.scene.remove(this.mesh);
  }

  protected addToScene() {
    this.scene.add(this.mesh);
  }

  protected updatePosition() {
    const positionComponent = this.entity.getComponent("position");
    if (!positionComponent) {
      return;
    }
    const footprintSize = this.getFootprintSize();
    const visualHeight = this.sampleVisualHeight(
      positionComponent.x + footprintSize.width / 2,
      positionComponent.y + footprintSize.height / 2
    );
    this.mesh.position.set(positionComponent.x, visualHeight, positionComponent.y);
  }

  protected getMetadataValue<T>(key: string, fallback: T): T {
    const metadataComponent = this.entity.getComponent("metadata") ?? {};
    return metadataComponent[key] ?? fallback;
  }

  protected getFootprintSize() {
    return {
      width: this.getMetadataValue("width", 1),
      height: this.getMetadataValue("height", 1),
    };
  }
}
