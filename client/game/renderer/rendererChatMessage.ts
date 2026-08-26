import * as THREE from "three";
import {
  createReactCss2dObject,
  type ReactCss2dObject,
} from "../../util/reactCss2dObject";
import OverheadChat from "../../ui/components/overheadChat";

export default class RendererChatMessage {
  private parent: THREE.Object3D;
  private overheadChat: ReactCss2dObject<{ text: string }>;

  constructor(parent: THREE.Object3D, text: string) {
    this.parent = parent;
    const overheadChatWrapper = createReactCss2dObject(OverheadChat, { text });
    this.overheadChat = overheadChatWrapper;
    this.overheadChat.object.position.x = 0.5;
    this.overheadChat.object.position.y = 1.5;
    this.overheadChat.object.position.z = 0.5;
    this.parent.add(this.overheadChat.object);
  }

  dispose() {
    this.parent.remove(this.overheadChat.object);
    this.overheadChat.dispose();
  }
}
