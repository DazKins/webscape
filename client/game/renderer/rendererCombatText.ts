import * as THREE from "three";
import {
  createReactCss2dObject,
  type ReactCss2dObject,
} from "../../util/reactCss2dObject";
import CombatText from "../../ui/components/combatText";

export default class RendererCombatText {
  private parent: THREE.Object3D;
  private combatText: ReactCss2dObject<{ text: string; kind: string }>;

  constructor(parent: THREE.Object3D, position: THREE.Vector3, text: string, kind: string) {
    this.parent = parent;
    const combatTextWrapper = createReactCss2dObject(CombatText, {
      text,
      kind,
    });
    this.combatText = combatTextWrapper;
    this.combatText.object.position.copy(position);
    this.parent.add(this.combatText.object);
  }

  dispose() {
    this.parent.remove(this.combatText.object);
    this.combatText.dispose();
  }

  reparent(parent: THREE.Object3D) {
    parent.attach(this.combatText.object);
    this.parent = parent;
  }
}
