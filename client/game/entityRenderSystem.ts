import * as THREE from "three";
import RendererHuman from "./renderer/rendererHuman";
import RendererChatMessage from "./renderer/rendererChatMessage";
import RendererCombatText from "./renderer/rendererCombatText";
import EntityRenderer from "./renderer/renderer";
import Entity from "./entity/entity";
import RendererBuilding from "./renderer/rendererBuilding";
import RendererChest from "./renderer/rendererChest";
import RendererDoor from "./renderer/rendererDoor";
import RendererError from "./renderer/rendererError";
import RendererRock from "./renderer/rendererRock";
import RendererTree from "./renderer/rendererTree";
import RendererRewardDrop from "./renderer/rendererRewardDrop";
import type { TerrainHeightSampler } from "./renderer/renderer";

type VisualHeightWorld = {
  getVisualHeightAtWorldPosition(worldX: number, worldZ: number): number;
};

export default class EntityRenderSystem {
  scene: THREE.Scene;
  renderers: Record<string, EntityRenderer | null>;
  private sampleVisualHeight: TerrainHeightSampler;

  constructor(scene: THREE.Scene, getWorld?: () => VisualHeightWorld | undefined) {
    this.scene = scene;
    this.renderers = {};
    this.sampleVisualHeight = (worldX: number, worldZ: number) =>
      getWorld?.()?.getVisualHeightAtWorldPosition(worldX, worldZ) ?? 0;
  }

  createRenderer(renderableType: string, entity: Entity): EntityRenderer | null {
    switch (renderableType) {
      case "human":
        return new RendererHuman(this.scene, entity, this.sampleVisualHeight);
      case "tree":
        return new RendererTree(this.scene, entity, this.sampleVisualHeight);
      case "door":
        return new RendererDoor(this.scene, entity, this.sampleVisualHeight);
      case "chest":
        return new RendererChest(this.scene, entity, this.sampleVisualHeight);
      case "rock":
        return new RendererRock(this.scene, entity, this.sampleVisualHeight);
      case "building":
        return new RendererBuilding(this.scene, entity, this.sampleVisualHeight);
      case "rewarddrop":
        return new RendererRewardDrop(this.scene, entity, this.sampleVisualHeight);
      case "chatmessage":
        const parentRenderer = this.renderers[entity.getComponent("chatmessage").fromEntityId];
        if (!parentRenderer) {
          console.error("parent renderer not found for chat message");
          return null;
        }
        return new RendererChatMessage(this.scene, entity, parentRenderer);
      case "combattext":
        const combatTextParent = this.renderers[entity.getComponent("combattext").fromEntityId];
        if (!combatTextParent) {
          console.error("parent renderer not found for combat text");
          return null;
        }
        return new RendererCombatText(this.scene, entity, combatTextParent);
    }
    console.error("unknown renderer type:", renderableType);
    return new RendererError(this.scene, entity, this.sampleVisualHeight);
  }

  update(entities: Entity[], deltaSeconds: number) {
    for (const entity of entities) {
      const renderableComponent = entity.getComponent("renderable");
      if (!renderableComponent) {
        continue;
      }

      let renderer = this.renderers[entity.getId()];
      if (!renderer) {
        renderer = this.createRenderer(renderableComponent.type, entity);
        if (renderer) {
          this.renderers[entity.getId()] = renderer;
        } else {
          continue;
        }
      }

      renderer!.update(deltaSeconds);
    }

    for (const entityId of Object.keys(this.renderers)) {
      const renderer = this.renderers[entityId];
      if (!renderer) {
        continue;
      }

      if (!entities.find((e) => e.getId() === entityId)) {
        renderer.onRemove();
        delete this.renderers[entityId];
      }
    }
  }

  getRenderers(): Record<string, EntityRenderer | null> {
    return this.renderers;
  }
}
