import Entity from "./entity/entity.ts";
import World from "./world/world.ts";
import Input, { InputReceiver } from "../input.ts";
import addReferenceGeometry from "./referenceGeometry.ts";
import { createCommand } from "../command/command.ts";
import * as THREE from "three";
import { WebSocketClient } from "../ws.ts";
import { CSS2DRenderer } from "three/examples/jsm/Addons.js";
import { InteractionMenuOpenEvent } from "../events/interactionMenu.ts";
import Camera from "./camera.ts";
import EntityRenderSystem from "./entityRenderSystem.ts";
import { ChatMessageEvent } from "../events/chat.ts";
import { InventoryUpdateEvent } from "../events/inventory.ts";
import { CombatLogUpdateEvent } from "../events/combatlog.ts";
import { ConversationCloseEvent, ConversationEvent, type ConversationPayload } from "../events/conversation.ts";
import { QuestLogUpdateEvent } from "../events/questlog.ts";
import { QuestCompletedEvent, type QuestCompletedPayload } from "../events/questCompleted.ts";

export type QuestDefinition = {
  id: string;
  displayName?: string;
  description?: string;
  startEventId?: string;
  steps: QuestStepDefinition[];
  rewards: {
    items: QuestRewardDefinition[];
  };
};

export type QuestStepDefinition = {
  id: string;
  description: string;
  requirement: {
    eventId: string;
    count: number;
  };
};

export type QuestRewardDefinition = {
  name: string;
  type: string;
  count: number;
};

class Game extends EventTarget implements InputReceiver {
  wsClient!: WebSocketClient;
  myPlayerId!: string;
  scene: THREE.Scene;
  camera: Camera;
  renderer: THREE.WebGLRenderer;
  cssRenderer2d: CSS2DRenderer;
  entities: Entity[];
  entityRenderSystem: EntityRenderSystem;
  quests: QuestDefinition[];
  activeConversation: ConversationPayload | null;

  input: Input;
  world!: World;

  typedChatText: string;

  constructor(sceneLayerRoot: HTMLElement, hudLayerRoot: HTMLElement) {
    super();

    this.scene = new THREE.Scene();
    this.input = new Input();
    this.camera = new Camera(this.input);
    this.entityRenderSystem = new EntityRenderSystem(this.scene, () => this.world);
    this.quests = [];
    this.activeConversation = null;

    this.renderer = new THREE.WebGLRenderer({ antialias: true });
    this.renderer.setSize(window.innerWidth, window.innerHeight);
    this.renderer.setClearColor(0x87ceeb);
    sceneLayerRoot.appendChild(this.renderer.domElement);

    this.cssRenderer2d = new CSS2DRenderer();
    this.cssRenderer2d.setSize(window.innerWidth, window.innerHeight);
    this.cssRenderer2d.domElement.style.position = "absolute";
    this.cssRenderer2d.domElement.style.top = "0";
    this.cssRenderer2d.domElement.style.pointerEvents = "none";
    hudLayerRoot.appendChild(this.cssRenderer2d.domElement);

    this.scene.add(new THREE.AmbientLight(0xffffff, 1.0));
    const light = new THREE.DirectionalLight(0xffffff, 0.8);
    light.position.set(10, 10, 10);
    this.scene.add(light);

    addReferenceGeometry(this.scene);

    this.onWindowResize = this.onWindowResize.bind(this);
    window.addEventListener("resize", this.onWindowResize, false);

    this.entities = [];

    this.input.registerClickCallback(() => {
      if (this.input.isPointerBlocked()) {
        return;
      }
      const hoveredTile = this.world.getHoveredTile(this.camera);
      if (hoveredTile) {
        this.handleMoveClick(hoveredTile.x, hoveredTile.y);
      }
    });

    this.input.registerRightClickCallback((event: MouseEvent) => {
      if (this.input.isPointerBlocked()) {
        return;
      }
      const mouse = this.input.getMousePosition();
      const mouseX = (mouse.x / window.innerWidth) * 2 - 1;
      const mouseY = -(mouse.y / window.innerHeight) * 2 + 1;

      const raycaster = new THREE.Raycaster();
      raycaster.setFromCamera(
        new THREE.Vector2(mouseX, mouseY),
        this.camera.getInnerCamera()
      );
      const object3Ds = Object.values(this.entityRenderSystem.getRenderers()).map((r) => r?.getObject3D() ?? null).filter((o) => o !== null);
      const intersects = raycaster.intersectObjects(
        object3Ds,
        true
      );
      if (intersects.length > 0) {
        let hitMesh: THREE.Object3D | null = intersects[0].object;

        while (hitMesh && !hitMesh.userData.entityId) {
          hitMesh = hitMesh.parent;
        }

        if (!hitMesh) {
          return;
        }

        const entityId = hitMesh.userData.entityId;
        const entity = this.entities.find((e) => e.getId() === entityId);
        if (entity) {
          const interactionOptions = entity.getAvailableInteractions();
          if (interactionOptions.length === 0) {
            return;
          }
          this.dispatchEvent(
            new InteractionMenuOpenEvent(
              entity.getId(),
              entity.getComponent("metadata")?.name || "Unnamed?!?!",
              interactionOptions,
              event.clientX,
              event.clientY
            )
          );
        }
      }
    });

    this.input.setActiveReceiver(this);

    this.typedChatText = "";
  }

  onKeyDown(key: string): void {
    const beforeTypedChatText = this.typedChatText;

    if (key === "Enter") {
      this.wsClient.sendMessage(
        createCommand("chat", { message: this.typedChatText })
      );
      this.typedChatText = "";
    } else if (key === "Escape") {
      this.typedChatText = "";
    } else if (key === "Backspace") {
      this.typedChatText = this.typedChatText.slice(0, -1);
    } else if (key.length === 1) {
      this.typedChatText += key;
    }

    if (beforeTypedChatText !== this.typedChatText) {
      this.dispatchEvent(
        new CustomEvent<string>("typedChatTextChanged", {
          detail: this.typedChatText,
        })
      );
    }
  }

  registerWsClient(wsClient: WebSocketClient) {
    this.wsClient = wsClient;
  }

  setPointerOverUi(isPointerOverUi: boolean) {
    this.input.setPointerBlocked(isPointerOverUi);
  }

  updateCamera() {
    const myEntity = this.getMyEntity();
    if (!myEntity) return;

    const myFocusPoint = this.getEntityFocusPoint(myEntity.getId());
    if (!myFocusPoint) {
      return;
    }

    const conversationTarget = this.getConversationTargetFocusPoint();
    if (conversationTarget) {
      this.camera.update(myFocusPoint.clone().add(conversationTarget).multiplyScalar(0.5), {
        distance: 3,
        height: 3,
      });
      return;
    }

    this.camera.update(myFocusPoint);
  }

  handleGameUpdate(gameUpdate: any) {
    const entityComponentsUpdates = gameUpdate.entities;

    for (const entityComponentUpdate of entityComponentsUpdates) {
      const entityId = entityComponentUpdate.entityId;
      const componentId = entityComponentUpdate.componentId;
      const data = entityComponentUpdate.data;

      let localEntity = this.entities.find((e) => e.getId() === entityId);
      if (!localEntity) {
        localEntity = new Entity(entityId);
        this.entities.push(localEntity);
      }

      localEntity.setAvailableInteractions(
        Array.isArray(entityComponentUpdate.availableInteractions)
          ? entityComponentUpdate.availableInteractions
          : []
      );

      if (componentId === "chatmessage" && data) {
        const fromEntityId = data.fromEntityId;

        const fromEntity = this.entities.find((e) => e.getId() === fromEntityId);
        if (!fromEntity) {
          continue;
        }

        const fromEntityMetadata = fromEntity.getComponent("metadata");
        if (!fromEntityMetadata) {
          continue;
        }

        this.dispatchEvent(new ChatMessageEvent(data.message, this.getEntityName(fromEntityId)));
      }

      if (data === null) {
        localEntity.removeComponent(componentId);
        continue;
      }

      localEntity.updateComponent(componentId, data);

      // Dispatch inventory update event if this is the player's inventory
      if ((componentId === "inventory" || componentId === "equipped") && entityId === this.myPlayerId) {
        this.dispatchEvent(new InventoryUpdateEvent());
      }

      if (componentId === "combatlog" && entityId === this.myPlayerId) {
        this.dispatchEvent(new CombatLogUpdateEvent());
      }

      if (componentId === "questlog" && entityId === this.myPlayerId) {
        this.dispatchEvent(new QuestLogUpdateEvent());
      }
    }

    const emptyEntities = this.entities.filter((e) => e.isEmpty());
    if (emptyEntities.length > 0) {
      this.entities = this.entities.filter((e) => !e.isEmpty());
    }
  }

  handleEntityRemove(entityId: string) {
    const entity = this.entities.find((e) => e.getId() === entityId);
    if (entity) {
      this.entities = this.entities.filter((e) => e.getId() !== entityId);
    }
  }

  update(deltaSeconds: number) {
    this.updateCamera();

    this.entityRenderSystem.update(this.entities, deltaSeconds);

    this.renderer.render(this.scene, this.camera.getInnerCamera());
    this.cssRenderer2d.render(this.scene, this.camera.getInnerCamera());
    if (this.world) {
      this.world.update(this.camera);
    }
  }

  onWindowResize() {
    this.camera.onWindowResize();
    this.renderer.setSize(window.innerWidth, window.innerHeight);
    this.cssRenderer2d.setSize(window.innerWidth, window.innerHeight);
  }

  registerMyPlayerId(myPlayerId: string) {
    this.myPlayerId = myPlayerId;
  }

  registerWorld(worldUpdate: any) {
    this.quests = worldUpdate.quests ?? [];
    this.world = new World(
      this.scene,
      worldUpdate.sizeX,
      worldUpdate.sizeY,
      worldUpdate.terrain ?? [],
      worldUpdate.heights ?? [],
      worldUpdate.blockers ?? [],
      worldUpdate.walls ?? [],
      this.input
    );
  }

  getMyEntity(): Entity | undefined {
    return this.entities.find((e) => e.getId() === this.myPlayerId);
  }

  getQuestDefinitions(): QuestDefinition[] {
    return this.quests;
  }

  getEntityName(entityId: string, fallback = "Unknown"): string {
    const entity = this.entities.find((e) => e.getId() === entityId);
    const name = entity?.getComponent("metadata")?.name;
    return typeof name === "string" && name.length > 0 ? name : fallback;
  }

  handleInteractionOptionClick(entityId: string, option: string) {
    this.wsClient.sendMessage(
      createCommand("interact", {
        entityId,
        option,
      })
    );
  }

  handleMoveClick(x: number, y: number) {
    this.closeActiveConversation();
    this.wsClient.sendMessage(
      createCommand("move", {
        x,
        y,
      })
    );
  }

  handleConversation(conversation: ConversationPayload) {
    this.activeConversation = conversation;
    this.dispatchEvent(new ConversationEvent(conversation));
  }

  handleQuestCompleted(payload: QuestCompletedPayload) {
    this.dispatchEvent(new QuestCompletedEvent(payload));
  }

  handleConversationClose(conversationId: string, nodeId: string) {
    if (
      this.activeConversation?.conversationId === conversationId &&
      this.activeConversation.nodeId === nodeId
    ) {
      this.closeActiveConversation();
    }
  }

  getActiveConversation(): ConversationPayload | null {
    return this.activeConversation;
  }

  isInConversation(): boolean {
    return this.activeConversation !== null;
  }

  getConversationTargetEntityId(): string | null {
    return this.activeConversation?.targetEntityId ?? null;
  }

  private closeActiveConversation() {
    if (!this.activeConversation) {
      return;
    }

    this.activeConversation = null;
    this.dispatchEvent(new ConversationCloseEvent());
  }

  private getConversationTargetFocusPoint(): THREE.Vector3 | null {
    const targetEntityId = this.getConversationTargetEntityId();
    return targetEntityId ? this.getEntityFocusPoint(targetEntityId) : null;
  }

  private getEntityFocusPoint(entityId: string): THREE.Vector3 | null {
    const renderer = this.entityRenderSystem.getRenderers()[entityId];
    const object3D = renderer?.getObject3D();
    if (object3D) {
      return new THREE.Vector3(
        object3D.position.x + 0.5,
        object3D.position.y,
        object3D.position.z + 0.5
      );
    }

    const entity = this.entities.find((candidate) => candidate.getId() === entityId);
    const position = entity?.getComponent("position");
    if (position && typeof position.x === "number" && typeof position.y === "number") {
      const visualHeight = this.world
        ? this.world.getVisualHeightAtTile(position.x, position.y)
        : 0;
      return new THREE.Vector3(position.x + 0.5, visualHeight, position.y + 0.5);
    }

    return null;
  }

  handleConversationOptionClick(
    conversationId: string,
    nodeId: string,
    optionId: string
  ) {
    this.wsClient.sendMessage(
      createCommand("conversationOption", {
        conversationId,
        nodeId,
        optionId,
      })
    );
  }

  handleEquipItem(itemId: string) {
    this.wsClient.sendMessage(
      createCommand("equip", {
        itemId,
      })
    );
  }

  handleUnequipSlot(slot: string) {
    this.wsClient.sendMessage(
      createCommand("unequip", {
        slot,
      })
    );
  }
}

export default Game;
