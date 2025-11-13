import * as THREE from "three";
import RendererHuman from "./renderer/rendererMan";
import RendererChatMessage from "./renderer/rendererChatMessage";
import Renderer from "./renderer/renderer";

export default class RenderSystem {
  scene: THREE.Scene;
  renderers: Record<string, Renderer>;

  constructor(scene: THREE.Scene) {
    this.scene = scene;
    this.renderers = {};
  }

  createRenderer(componentSignature: string): Renderer | null {
    switch (componentSignature) {
      case "human":
        return new RendererHuman();
      case "chatmessage":
        return new RendererChatMessage();
    }
    console.error("unknown renderer type:", componentSignature);
    return null;
  }

  handleGameUpdate(gameUpdate: any) {
    const updatedEntities = gameUpdate.entities;
    for (const entity of updatedEntities) {
      const componentId = entity.componentId;
      if (componentId !== "renderable") {
        continue;
      }
      const renderer = this.createRenderer(entity.data.type);
      if (renderer) {
        this.renderers[entity.entityId] = renderer;
      }
    }

    for (const [entityId, renderer] of Object.entries(this.renderers)) {
      const updates = updatedEntities.filter(
        (e: any) => e.entityId === entityId
      );
      for (const update of updates) {
        renderer.handleComponentUpdate(entityId, update.componentId, update.data);
      }
    }
  }
}
