import * as THREE from "three";
import Entity from "../entity/entity";
import { createModel, isModelName, type ModelInstance } from "../models";
import { applyModelTransform, getEquipmentPresentation } from "../models/equipment";
import PositionedEntityRenderer from "./positionedEntityRenderer";
import type { TerrainHeightSampler } from "./renderer";

export default class RendererDroppedItem extends PositionedEntityRenderer {
  private readonly modelInstance: ModelInstance;
  private readonly hitTarget: THREE.Mesh<THREE.BoxGeometry, THREE.MeshBasicMaterial>;

  constructor(scene: THREE.Scene, entity: Entity, terrainHeightSampler?: TerrainHeightSampler) {
    super(scene, entity, terrainHeightSampler);
    const modelName = this.getMetadataValue("renderModel", "unknownItem");
    this.modelInstance = createModel(isModelName(modelName) ? modelName : "unknownItem");
    const root = this.modelInstance.root;
    applyModelTransform(root, getEquipmentPresentation(modelName)?.dropped ?? {});

    // Equipment models are authored around attachment pivots. Center the
    // transformed bounds on the tile and rest the lowest point on the terrain.
    const bounds = new THREE.Box3().setFromObject(root);
    const center = bounds.getCenter(new THREE.Vector3());
    root.position.add(new THREE.Vector3(0.5 - center.x, -bounds.min.y + 0.015, 0.5 - center.z));
    this.mesh.add(root);

    // Give even thin items (arrows, keys) a consistent clickable area. Material
    // visibility hides the box from rendering while retaining mesh raycasting.
    this.hitTarget = new THREE.Mesh(
      new THREE.BoxGeometry(0.8, 0.3, 0.8),
      new THREE.MeshBasicMaterial({ visible: false }),
    );
    this.hitTarget.name = "droppedItemHitTarget";
    this.hitTarget.position.set(0.5, 0.15, 0.5);
    this.mesh.add(this.hitTarget);
    this.addToScene();
  }

  update(deltaSeconds: number) {
    super.update(deltaSeconds);
    this.modelInstance.update(deltaSeconds);
  }

  onRemove() {
    this.hitTarget.geometry.dispose();
    this.hitTarget.material.dispose();
    this.modelInstance.dispose();
    super.onRemove();
  }
}
