import EntityRenderer from "./renderer";
import * as THREE from "three";
import Entity from "../entity/entity";
import { createReactCss2dObject } from "../../util/reactCss2dObject";
import CombatText from "../../ui/components/combatText";

export default class RendererCombatText extends EntityRenderer {
  private parentEntityRenderer: EntityRenderer;
  private combatText: THREE.Object3D;

  constructor(scene: THREE.Scene, entity: Entity, parentEntityRenderer: EntityRenderer) {
    super(scene, entity);

    const combatTextComponent = entity.getComponent("combattext");

    this.parentEntityRenderer = parentEntityRenderer;
    const combatTextWrapper = createReactCss2dObject(CombatText, {
      text: combatTextComponent.text,
      kind: combatTextComponent.kind,
    });
    this.combatText = combatTextWrapper.object;
    this.combatText.position.x = 0.5;
    this.combatText.position.y = 1.8;
    this.combatText.position.z = 0.5;
    this.parentEntityRenderer.getObject3D()?.add(this.combatText);
  }

  update(_deltaSeconds: number) {}

  onRemove() {
    this.parentEntityRenderer.getObject3D()?.remove(this.combatText);
  }
}
