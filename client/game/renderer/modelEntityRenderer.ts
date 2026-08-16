import * as THREE from "three";
import Entity from "../entity/entity";
import { createModel, type ModelInstance, type ModelName, type ModelOptions } from "../models";
import PositionedEntityRenderer from "./positionedEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";

export default class ModelEntityRenderer extends PositionedEntityRenderer {
  protected readonly modelInstance: ModelInstance;

  constructor(
    scene: THREE.Scene,
    entity: Entity,
    terrainHeightSampler: TerrainHeightSampler | undefined,
    modelName: ModelName,
    modelOptions: ModelOptions = {},
  ) {
    super(scene, entity, terrainHeightSampler);
    this.modelInstance = createModel(modelName, modelOptions);
    const footprint = this.getFootprintSize();
    this.modelInstance.root.position.set(footprint.width / 2, 0, footprint.height / 2);
    this.mesh.add(this.modelInstance.root);
    this.addToScene();
  }

  update(deltaSeconds: number) {
    super.update(deltaSeconds);
    this.modelInstance.update(deltaSeconds);
  }

  onRemove() {
    this.modelInstance.dispose();
    super.onRemove();
  }
}
