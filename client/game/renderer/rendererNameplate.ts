import type * as THREE from "three";
import { createReactCss2dObject, type ReactCss2dObject } from "../../util/reactCss2dObject";
import Nameplate from "../../ui/components/nameplate";

export default class RendererNameplate {
  private label: ReactCss2dObject<{ text: string }>;

  constructor(parent: THREE.Object3D, private text: string) {
    this.label = createReactCss2dObject(Nameplate, { text });
    this.label.object.position.set(0.5, 1.85, 0.5);
    parent.add(this.label.object);
  }

  update(text: string) {
    if (this.text === text) return;
    this.text = text;
    this.label.updateProps({ text });
  }

  dispose() {
    this.label.dispose();
  }
}
