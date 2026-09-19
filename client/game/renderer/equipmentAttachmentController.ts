import type * as THREE from "three";
import {
  createModel,
  isModelName,
  type ModelInstance,
} from "../models";
import {
  applyModelTransform,
  getEquipmentPresentation,
} from "../models/equipment";

type EquippedItem = {
  id: string;
  renderModel?: string;
};

export type EquippedComponent = {
  slots?: Record<string, EquippedItem | null>;
};

type Attachment = {
  itemId: string;
  renderModel: string;
  modelInstance: ModelInstance;
  parts: THREE.Object3D[];
  hiddenParts: THREE.Object3D[];
};

export default class EquipmentAttachmentController {
  private readonly attachments = new Map<string, Attachment>();
  private readonly hiddenParts = new Map<THREE.Object3D, { count: number; wasVisible: boolean }>();

  constructor(private readonly hostModel: ModelInstance) {}

  update(equipped: EquippedComponent | undefined, deltaSeconds: number): void {
    const slots = equipped?.slots ?? {};

    for (const slot of this.attachments.keys()) {
      if (!slots[slot]) {
        this.remove(slot);
      }
    }

    for (const [slot, item] of Object.entries(slots)) {
      this.sync(slot, item);
    }

    for (const attachment of this.attachments.values()) {
      attachment.modelInstance.update(deltaSeconds);
    }
  }

  dispose(): void {
    for (const slot of [...this.attachments.keys()]) {
      this.remove(slot);
    }
  }

  getAttachmentObject(slot: string): THREE.Object3D | undefined {
    return this.attachments.get(slot)?.modelInstance.root;
  }

  seekWeaponAnimation(name: string, phase: number): void {
    const model = this.attachments.get("weapon")?.modelInstance;
    if (model?.animations.some(animation => animation.name === name)) {
      model.seek(name, phase);
    }
  }

  private sync(slot: string, item: EquippedItem | null): void {
    const current = this.attachments.get(slot);
    if (!item?.renderModel) {
      this.remove(slot);
      return;
    }
    if (current?.itemId === item.id && current.renderModel === item.renderModel) {
      return;
    }

    this.remove(slot);
    const presentation = getEquipmentPresentation(item.renderModel);
    if (!presentation || !isModelName(item.renderModel)) {
      return;
    }

    const socket = this.hostModel.getSocket(presentation.equipped.socket);
    if (!socket) {
      console.error(
        `model does not define equipment socket: ${presentation.equipped.socket}`,
      );
      return;
    }

    const modelInstance = createModel(item.renderModel, { equipped: true });
    const parts: THREE.Object3D[] = [];
    for (const [partName, socketName] of Object.entries(presentation.equipped.parts ?? {})) {
      const part = modelInstance.root.getObjectByName(partName);
      const partSocket = this.hostModel.getSocket(socketName);
      if (!part || !partSocket) {
        for (const attached of parts) modelInstance.root.add(attached);
        modelInstance.dispose();
        console.error(`cannot attach equipment part: ${partName} to ${socketName}`);
        return;
      }
      partSocket.add(part);
      parts.push(part);
    }
    applyModelTransform(modelInstance.root, presentation.equipped);
    socket.add(modelInstance.root);
    const hiddenParts: THREE.Object3D[] = [];
    for (const name of presentation.equipped.hideParts ?? []) {
      const part = this.hostModel.root.getObjectByName(name);
      if (!part) continue;
      const hidden = this.hiddenParts.get(part) ?? { count: 0, wasVisible: part.visible };
      hidden.count++;
      this.hiddenParts.set(part, hidden);
      part.visible = false;
      hiddenParts.push(part);
    }
    this.attachments.set(slot, {
      itemId: item.id,
      renderModel: item.renderModel,
      modelInstance,
      parts,
      hiddenParts,
    });
  }

  private remove(slot: string): void {
    const attachment = this.attachments.get(slot);
    if (!attachment) {
      return;
    }
    // Reunite detached groups before disposing, leaving no armour on the rig.
    for (const part of attachment.parts) attachment.modelInstance.root.add(part);
    attachment.modelInstance.dispose();
    for (const part of attachment.hiddenParts) {
      const hidden = this.hiddenParts.get(part)!;
      if (--hidden.count === 0) {
        part.visible = hidden.wasVisible;
        this.hiddenParts.delete(part);
      }
    }
    this.attachments.delete(slot);
  }
}
