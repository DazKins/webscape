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
};

export default class EquipmentAttachmentController {
  private readonly attachments = new Map<string, Attachment>();

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

    const modelInstance = createModel(item.renderModel);
    applyModelTransform(modelInstance.root, presentation.equipped);
    socket.add(modelInstance.root);
    this.attachments.set(slot, {
      itemId: item.id,
      renderModel: item.renderModel,
      modelInstance,
    });
  }

  private remove(slot: string): void {
    const attachment = this.attachments.get(slot);
    if (!attachment) {
      return;
    }
    attachment.modelInstance.dispose();
    this.attachments.delete(slot);
  }
}
