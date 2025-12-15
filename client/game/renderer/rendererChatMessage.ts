import EntityRenderer from "./renderer";
import * as THREE from "three";
import Entity from "../entity/entity";
import { createReactCss2dObject } from "../../util/reactCss2dObject";
import OverheadChat from "../../ui/components/overheadChat";
import { CSS2DObject } from "three/examples/jsm/renderers/CSS2DRenderer.js";

export default class RendererChatMessage extends EntityRenderer {
  private parentEntityRenderer: EntityRenderer;
  private overheadChat: CSS2DObject;

  constructor(scene: THREE.Scene, entity: Entity, parentEntityRenderer: EntityRenderer) {
    super(scene, entity);

    const chatMessageComponent = entity.getComponent("chatmessage");

    this.parentEntityRenderer = parentEntityRenderer;
    this.overheadChat = createReactCss2dObject(OverheadChat, { text: chatMessageComponent.message })
    this.overheadChat.position.x = 0.5;
    this.overheadChat.position.y = 1.5;
    this.overheadChat.position.z = 0.5;
    this.parentEntityRenderer.getObject3D()?.add(this.overheadChat);
  }

  update() { }

  onRemove() {
    this.parentEntityRenderer.getObject3D()?.remove(this.overheadChat);
  }
}
