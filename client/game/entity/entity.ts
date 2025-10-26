import * as THREE from "three";
import { createEntityNamecard } from "../ui/entityNamecard";
import { createEntityChat } from "../ui/entityChat";
import { CSS2DObject } from "three/examples/jsm/Addons.js";

class Entity {
  id: string;
  scene: THREE.Scene;
  interactionOptions: string[];
  mesh: THREE.Group;
  positionX: number;
  positionY: number;
  name?: string;
  chat?: CSS2DObject;

  constructor(id: string, scene: THREE.Scene) {
    this.id = id;
    this.scene = scene;
    this.interactionOptions = [];
    this.positionX = 0;
    this.positionY = 0;

    this.mesh = new THREE.Group();
    this.mesh.userData.entityId = id;

    const bodyGeometry = new THREE.CylinderGeometry(0.3, 0.3, 0.8, 16);
    const bodyMaterial = new THREE.MeshPhongMaterial({
      color: 0x00ff00,
    });
    const body = new THREE.Mesh(bodyGeometry, bodyMaterial);
    body.position.y = 0.4;
    this.mesh.add(body);

    const headGeometry = new THREE.SphereGeometry(0.25, 16, 16);
    const headMaterial = new THREE.MeshPhongMaterial({
      color: 0xffe0bd,
    });
    const head = new THREE.Mesh(headGeometry, headMaterial);
    head.position.y = 0.9;
    this.mesh.add(head);

    this.scene.add(this.mesh);
  }

  update() {
    this.mesh.position.x +=
      (this.positionX + 0.5 - this.mesh.position.x) * 0.05;
    this.mesh.position.z +=
      (this.positionY + 0.5 - this.mesh.position.z) * 0.05;
  }

  chatTimeoutId?: number;

  handleChat(text: string) {
    if (this.chatTimeoutId) {
      clearTimeout(this.chatTimeoutId);
    }
    if (this.chat) {
      this.mesh.remove(this.chat);
    }
    this.chat = createEntityChat(text);
    this.mesh.add(this.chat);
    this.chatTimeoutId = setTimeout(() => {
      if (this.chat) {
        this.mesh.remove(this.chat);
      }
      this.chat = undefined;
    }, 7000);
  }

  handleEntityUpdate(entityUpdate: any) {
    const positionComponent = entityUpdate.components.position;
    const metadataComponent = entityUpdate.components.metadata;
    const interactableComponent = entityUpdate.components.interactable;

    this.positionX = positionComponent.x;
    this.positionY = positionComponent.y;
    if (metadataComponent.name && metadataComponent.name !== this.name) {
      const name = metadataComponent.name as string;
      this.name = name;
      const namecard = createEntityNamecard(name, 0, 1.5, 0);
      this.mesh.add(namecard);
    }

    if (interactableComponent) {
      this.interactionOptions = interactableComponent.interactionOptions;
    }

    if (metadataComponent.color) {
      const body = this.mesh.children[0];
      if (body instanceof THREE.Mesh) {
        body.material.color.set(metadataComponent.color);
      }
    }
  }

  remove() {
    console.log("Removing entity from scene", this.id);
    this.scene.remove(this.mesh);
  }
}

export default Entity;
