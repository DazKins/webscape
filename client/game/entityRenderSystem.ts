import * as THREE from "three";
import RendererHuman from "./renderer/rendererHuman";
import RendererChatMessage from "./renderer/rendererChatMessage";
import EntityRenderer from "./renderer/renderer";
import Entity from "./entity/entity";

export default class EntityRenderSystem {
  scene: THREE.Scene;
  renderers: Record<string, EntityRenderer | null>;

  constructor(scene: THREE.Scene) {
    this.scene = scene;
    this.renderers = {};
  }

  createRenderer(renderableType: string, entity: Entity): EntityRenderer | null {
    switch (renderableType) {
      case "human":
        return new RendererHuman(this.scene, entity);
      case "chatmessage":
        const parentRenderer = this.renderers[entity.getComponent("chatmessage").fromEntityId];
        if (!parentRenderer) {
          console.error("parent renderer not found for chat message");
          return null;
        }
        return new RendererChatMessage(this.scene, entity, parentRenderer);
    }
    console.error("unknown renderer type:", renderableType);
    return null;
  }

  update(entities: Entity[]) {
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

      renderer!.update();
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
