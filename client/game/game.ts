import InventoryPrediction, { type InventoryLayoutItem } from "./inventoryPrediction.ts";
import { BankUpdateEventName, BankResultEvent, type BankResultPayload } from "../events/bank";
import { ShopUpdateEventName, TradeResultEvent, type TradeResultPayload } from "../events/shop";
import Entity from "./entity/entity.ts";
import DayCycleClock from "./dayCycle";
import RenderTiming from "./renderTiming";
import EnvironmentLighting from "./environmentLighting";
import World, { type ChunkUpdate } from "./world/world.ts";
import Input from "../input.ts";
import addReferenceGeometry from "./referenceGeometry.ts";
import { createCommand } from "../command/command.ts";
import * as THREE from "three";
import { WebSocketClient } from "../ws.ts";
import { CSS2DRenderer } from "three/examples/jsm/Addons.js";
import { InteractionMenuOpenEvent } from "../events/interactionMenu.ts";
import Camera from "./camera.ts";
import EntityRenderSystem, { type CombatProjectileLaunchedPayload } from "./entityRenderSystem.ts";
import { ChatMessageEvent } from "../events/chat.ts";
import { InventoryUpdateEvent } from "../events/inventory.ts";
import { PlayerVitalsUpdateEventName } from "../events/playerVitals";
import { CombatLogUpdateEvent } from "../events/combatlog.ts";
import { ConversationCloseEvent, ConversationEvent, type ConversationPayload } from "../events/conversation.ts";
import { QuestLogUpdateEvent } from "../events/questlog.ts";
import { QuestCompletedEvent, type QuestCompletedPayload } from "../events/questCompleted.ts";
import { QuestStartedEvent, type QuestStartedPayload } from "../events/questStarted.ts";
import {
  getDeviceProfile,
  getElementSize,
  type DeviceProfile,
  type ViewportSize,
} from "../responsive.ts";

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

export type ChatMessagePayload = {
  fromEntityId: string;
  message: string;
};

export type CombatResolvedPayload = {
  attackerEntityId: string;
  targetEntityId: string;
  didHit: boolean;
  damage: number;
  isCritical: boolean;
  attackMethod?: string;
};

const SERVER_TICK_MILLISECONDS = 500;

class Game extends EventTarget {
  wsClient!: WebSocketClient;
  private inventoryPrediction = new InventoryPrediction();
  myPlayerId!: string;
  scene: THREE.Scene;
  camera: Camera;
  renderer: THREE.WebGLRenderer;
  renderTiming: RenderTiming | null = null;
  cssRenderer2d: CSS2DRenderer;
  sceneLayerRoot: HTMLElement;
  viewport: ViewportSize;
  deviceProfile: DeviceProfile;
  resizeObserver: ResizeObserver;
  entities: Entity[];
  entityRenderSystem: EntityRenderSystem;
  quests: QuestDefinition[];
  activeConversation: ConversationPayload | null;
  observerFocus: { x: number; y: number };
  readonly dayCycle = new DayCycleClock();
  private readonly lighting: EnvironmentLighting;
  private readonly lightingFocus = new THREE.Vector3();
  private latestServerTick = 0;
  private serverTickMilliseconds = SERVER_TICK_MILLISECONDS;
  private serverTickReceivedAtMilliseconds = performance.now();
  private dismissedBankTargetId: string | null = null;

  input: Input;
  world?: World;

  typedChatText: string;

  constructor(sceneLayerRoot: HTMLElement, hudLayerRoot: HTMLElement) {
    super();

    this.sceneLayerRoot = sceneLayerRoot;
    this.viewport = getElementSize(sceneLayerRoot);
    this.deviceProfile = getDeviceProfile(this.viewport);
    this.scene = new THREE.Scene();
    this.lighting = new EnvironmentLighting(this.scene);
    this.input = new Input();
    this.camera = new Camera(this.input, this.viewport);
    this.entityRenderSystem = new EntityRenderSystem(
      this.scene,
      () => this.world,
      () => this.estimatedServerTick(),
      () => this.serverTickMilliseconds / 1000,
      this.lighting,
    );
    this.quests = [];
    this.activeConversation = null;
    this.observerFocus = { x: 0, y: 0 };

    this.renderer = new THREE.WebGLRenderer({ antialias: true, powerPreference: "low-power" });
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 1.5));
    this.renderer.setSize(this.viewport.width, this.viewport.height);
    this.renderer.setClearColor(0x87ceeb);
    sceneLayerRoot.appendChild(this.renderer.domElement);

    this.cssRenderer2d = new CSS2DRenderer();
    this.cssRenderer2d.setSize(this.viewport.width, this.viewport.height);
    this.cssRenderer2d.domElement.style.position = "absolute";
    this.cssRenderer2d.domElement.style.top = "0";
    this.cssRenderer2d.domElement.style.pointerEvents = "none";
    hudLayerRoot.appendChild(this.cssRenderer2d.domElement);

    addReferenceGeometry(this.scene);

    this.onViewportResize = this.onViewportResize.bind(this);
    this.resizeObserver = new ResizeObserver(this.onViewportResize);
    this.resizeObserver.observe(sceneLayerRoot);
    window.addEventListener("resize", this.onViewportResize, false);
    window.addEventListener("orientationchange", this.onViewportResize, false);

    this.entities = [];

    this.input.registerPointerCallbacks({
      onTap: (event) => {
        this.handleSceneTap(event.clientX, event.clientY);
      },
      onLongPress: (event) => {
        this.handleSceneLongPress(event.clientX, event.clientY);
      },
      onDrag: (event, delta) => {
        this.handleSceneDrag(event, delta);
      },
    });

    this.input.registerRightClickCallback((event: MouseEvent) => {
      if (this.input.isPointerBlocked()) {
        return;
      }
      this.openInteractionMenuAt(event.clientX, event.clientY);
    });

    this.typedChatText = "";
  }

  registerWsClient(wsClient: WebSocketClient) {
    this.wsClient = wsClient;
  }

  setPointerOverUi(isPointerOverUi: boolean) {
    this.input.setPointerBlocked(isPointerOverUi);
  }

  setRegistrationBlocked(blocked: boolean) {
    this.input.setWorldBlocked(blocked);
  }

  setMenuBlocked(blocked: boolean) {
    this.input.setMenuBlocked(blocked);
  }

  setRenderTimingEnabled(enabled: boolean) {
    this.renderTiming?.reset();
    // Three.js's WebGLRenderer uses WebGL 2 exclusively.
    this.renderTiming = enabled ? new RenderTiming(this.renderer.getContext() as WebGL2RenderingContext) : null;
  }

  getDeviceProfile(): DeviceProfile {
    return this.deviceProfile;
  }

  setTypedChatText(text: string) {
    if (text === this.typedChatText) {
      return;
    }
    this.typedChatText = text;
    this.dispatchTypedChatTextChanged();
  }

  sendTypedChatText() {
    const message = this.typedChatText.trim();
    if (message.length > 0) {
      this.wsClient.sendMessage(createCommand("chat", { message }));
    }
    this.setTypedChatText("");
  }

  private dispatchTypedChatTextChanged() {
    this.dispatchEvent(
      new CustomEvent<string>("typedChatTextChanged", {
        detail: this.typedChatText,
      })
    );
  }

  private handleSceneTap(clientX: number, clientY: number) {
    if (!this.world || this.input.isPointerBlocked()) {
      return;
    }

    if (this.openInteractionMenuAt(clientX, clientY)) {
      return;
    }

    const tile = this.world.getPointerTile(this.camera, this.viewport);
    if (tile) {
      this.world.showTileIndicator(tile);
      this.handleMoveClick(tile.x, tile.y);
    }
  }

  private handleSceneLongPress(clientX: number, clientY: number) {
    if (this.input.isPointerBlocked()) {
      return;
    }
    this.openInteractionMenuAt(clientX, clientY);
  }

  private handleSceneDrag(event: PointerEvent, delta: { x: number; y: number }) {
    if (this.input.isPointerBlocked() || event.buttons !== 1) {
      return;
    }
    this.camera.orbitByDrag(delta);
  }

  private openInteractionMenuAt(clientX: number, clientY: number) {
    const entity = this.getEntityAtScreenPoint(clientX, clientY);
    if (!entity) {
      return false;
    }

    const interactionOptions = entity.getAvailableInteractions();
    if (interactionOptions.length === 0) {
      return false;
    }

    this.dispatchEvent(
      new InteractionMenuOpenEvent(
        entity.getId(),
        entity.getComponent("metadata")?.name || "Unnamed?!?!",
        interactionOptions,
        clientX,
        clientY
      )
    );
    return true;
  }

  private getEntityAtScreenPoint(clientX: number, clientY: number): Entity | null {
    const mouseX = (clientX / this.viewport.width) * 2 - 1;
    const mouseY = -(clientY / this.viewport.height) * 2 + 1;

    const raycaster = new THREE.Raycaster();
    raycaster.setFromCamera(
      new THREE.Vector2(mouseX, mouseY),
      this.camera.getInnerCamera()
    );
    const object3Ds = Object.values(this.entityRenderSystem.getRenderers())
      // The local player must not obscure selectable items beneath their feet.
      .filter((renderer) => renderer?.entity.getId() !== this.myPlayerId)
      .map((renderer) => renderer?.getObject3D() ?? null)
      .filter((object3d): object3d is THREE.Object3D => object3d !== null);
    const intersects = raycaster.intersectObjects(object3Ds, true);
    if (intersects.length === 0) {
      return null;
    }

    let hitMesh: THREE.Object3D | null = intersects[0].object;
    while (hitMesh && !hitMesh.userData.entityId) {
      hitMesh = hitMesh.parent;
    }

    if (!hitMesh) {
      return null;
    }

    return this.entities.find((entity) => entity.getId() === hitMesh.userData.entityId) ?? null;
  }

  updateCamera(deltaSeconds: number) {
    const myEntity = this.getMyEntity();
    if (!myEntity) {
      const visualHeight = this.world
        ? this.world.getVisualHeightAtTile(this.observerFocus.x, this.observerFocus.y)
        : 0;
      this.camera.update(
        new THREE.Vector3(
          this.observerFocus.x + 0.5,
          visualHeight,
          this.observerFocus.y + 0.5
        ),
        this.deviceProfile.isMobileLayout
          ? {
              distance: this.deviceProfile.isPortrait ? 7.1 : 6.4,
              height: this.deviceProfile.isPortrait ? 6.6 : 4.9,
            }
          : {},
        deltaSeconds,
      );
      return;
    }

    const myFocusPoint = this.getEntityFocusPoint(myEntity.getId());
    if (!myFocusPoint) {
      return;
    }

    const conversationTarget = this.getConversationTargetFocusPoint();
    if (conversationTarget) {
      const conversationDistance = this.deviceProfile.isMobileLayout ? 4.5 : 3;
      const conversationHeight = this.deviceProfile.isMobileLayout ? 4.1 : 3;
      this.camera.update(myFocusPoint.clone().add(conversationTarget).multiplyScalar(0.5), {
        distance: conversationDistance,
        height: conversationHeight,
      }, deltaSeconds);
      return;
    }

    if (this.deviceProfile.isMobileLayout) {
      this.camera.update(myFocusPoint, {
        distance: this.deviceProfile.isPortrait ? 7.1 : 6.4,
        height: this.deviceProfile.isPortrait ? 6.6 : 4.9,
      }, deltaSeconds);
      return;
    }

    this.camera.update(myFocusPoint, {}, deltaSeconds);
  }

  handleGameUpdate(gameUpdate: any) {
    if (
      typeof gameUpdate.serverTick === "number" &&
      Number.isFinite(gameUpdate.serverTick) &&
      gameUpdate.serverTick >= 0
    ) {
      this.latestServerTick = gameUpdate.serverTick;
      this.serverTickReceivedAtMilliseconds = performance.now();
      this.dayCycle.receive(gameUpdate.serverTick, this.serverTickReceivedAtMilliseconds);
    }
    const entityComponentsUpdates = gameUpdate.entities;
    let inventoryChanged = false;
    let vitalsChanged = false;

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

      if ((componentId === "inventory" || componentId === "equipped") && entityId === this.myPlayerId) {
        inventoryChanged = true;
      }
      if ((componentId === "health" || componentId === "mana") && entityId === this.myPlayerId) vitalsChanged = true;
      if (data === null) {
        localEntity.removeComponent(componentId);
        continue;
      }

      localEntity.updateComponent(componentId, data);


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
    // UI observers read a complete snapshot, including removed components.
    const inventoryAcknowledged = this.inventoryPrediction.acknowledge(gameUpdate.inventoryMoveSequence);
    if (inventoryChanged || inventoryAcknowledged) this.dispatchEvent(new InventoryUpdateEvent());
    if (vitalsChanged) this.dispatchEvent(new Event(PlayerVitalsUpdateEventName));
    if (this.getMyEntity()?.getComponent("trading") || this.getMyEntity()?.getComponent("banking")) this.closeActiveConversation();
    this.dispatchEvent(new Event(ShopUpdateEventName));
    this.dispatchEvent(new Event(BankUpdateEventName));
  }

  update(deltaSeconds: number) {
    this.updateCamera(deltaSeconds);
    if (this.world) {
      this.world.update(this.camera, deltaSeconds, this.deviceProfile);
    }

    this.entityRenderSystem.update(this.entities, deltaSeconds, this.getMyEntity()?.getId());

    this.lightingFocus.copy(this.getEntityFocusPoint(this.myPlayerId) ??
      this.lightingFocus.set(this.observerFocus.x, 0, this.observerFocus.y));
    this.lighting.update(this.dayCycle.read()?.cycleProgress ?? 0, this.lightingFocus);

    this.renderTiming?.begin();
    try {
      this.renderer.render(this.scene, this.camera.getInnerCamera());
    } finally {
      this.renderTiming?.end();
    }
    this.cssRenderer2d.render(this.scene, this.camera.getInnerCamera());
    this.dispatchEvent(new Event("frameRendered"));
  }

  private estimatedServerTick(): number {
    const elapsedMilliseconds = Math.max(
      0,
      performance.now() - this.serverTickReceivedAtMilliseconds,
    );
    return this.latestServerTick + elapsedMilliseconds / this.serverTickMilliseconds;
  }

  private resetServerClock() {
    this.dayCycle.reset();
    this.latestServerTick = 0;
    this.serverTickReceivedAtMilliseconds = performance.now();
  }

  onViewportResize() {
    const nextViewport = getElementSize(this.sceneLayerRoot);
    if (
      nextViewport.width === this.viewport.width &&
      nextViewport.height === this.viewport.height
    ) {
      this.deviceProfile = getDeviceProfile(nextViewport);
      return;
    }

    this.viewport = nextViewport;
    this.deviceProfile = getDeviceProfile(nextViewport);
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 1.5));
    this.camera.onResize(nextViewport);
    this.renderer.setSize(nextViewport.width, nextViewport.height);
    this.cssRenderer2d.setSize(nextViewport.width, nextViewport.height);
  }

  registerMyPlayerId(myPlayerId: string) {
    this.myPlayerId = myPlayerId;
    this.dispatchEvent(new Event(PlayerVitalsUpdateEventName));
  }

  clearSession() {
    this.prepareForReconnect();
    this.world?.dispose();
    this.world = undefined;
    this.entities = [];
    this.quests = [];
    this.typedChatText = "";
    this.entityRenderSystem.update(this.entities, 0);
  }

  prepareForReconnect() {
    this.entityRenderSystem.clearTransientEffects();
    this.myPlayerId = "";
    this.inventoryPrediction.reset();
    this.dispatchEvent(new InventoryUpdateEvent());
    this.dismissedBankTargetId = null;
    this.dispatchEvent(new Event(PlayerVitalsUpdateEventName));
    this.dispatchEvent(new Event(ShopUpdateEventName));
    this.dispatchEvent(new Event(BankUpdateEventName));
    this.activeConversation = null;
    this.resetServerClock();
  }

  registerWorld(worldUpdate: any) {
    this.serverTickMilliseconds =
      typeof worldUpdate.tickIntervalMs === "number" &&
      Number.isFinite(worldUpdate.tickIntervalMs) && worldUpdate.tickIntervalMs > 0
        ? worldUpdate.tickIntervalMs : SERVER_TICK_MILLISECONDS;
    this.entityRenderSystem.clearTransientEffects();
    this.resetServerClock();
    this.world?.dispose();
    this.entities = [];
    this.entityRenderSystem.update(this.entities, 0);
    this.quests = worldUpdate.quests ?? [];
    const chunkSize = worldUpdate.chunkSize ?? { x: 32, y: 32 };
    this.observerFocus = worldUpdate.playerSpawn ?? { x: 0, y: 0 };
    this.world = new World(this.scene, chunkSize, this.input);
    this.dayCycle.configure(worldUpdate.dayCycle, this.serverTickMilliseconds);
  }

  handleChunkUpdate(update: ChunkUpdate) {
    this.world?.applyChunkUpdate(update);
  }

  handleAdminCommandResult(payload: { command: string; success: boolean; message: string }) {
    if (payload.success && payload.command === "/reset") this.closeActiveConversation();
    this.dispatchEvent(new ChatMessageEvent(payload.message, "Admin"));
  }

  handleChatMessage(payload: ChatMessagePayload) {
    this.entityRenderSystem.showChatMessage(payload.fromEntityId, payload.message);
    this.dispatchEvent(
      new ChatMessageEvent(payload.message, this.getEntityName(payload.fromEntityId)),
    );
  }

  handleCombatResolved(payload: CombatResolvedPayload) {
    this.entityRenderSystem.showCombatResult(
      payload.attackerEntityId,
      payload.targetEntityId,
      payload.didHit,
      payload.damage,
      payload.isCritical,
      payload.attackMethod ?? "melee",
    );
  }

  handleItemPickedUp(payload: { playerEntityId: string }) {
    this.entityRenderSystem.showItemPickup(payload.playerEntityId);
  }

  handleCombatProjectileLaunched(payload: CombatProjectileLaunchedPayload) {
    this.entityRenderSystem.showCombatProjectile(payload);
  }

  getMyEntity(): Entity | undefined {
    return this.entities.find((e) => e.getId() === this.myPlayerId);
  }

  getQuestDefinitions(): QuestDefinition[] {
    return this.quests;
  }

  getEntity(entityId: string): Entity | undefined {
    return this.entities.find((entity) => entity.getId() === entityId);
  }

  handleBankTransfer(targetEntityId: string, action: "deposit" | "withdraw", itemId: string, quantity: number) {
    this.wsClient.sendMessage(createCommand(action === "deposit" ? "bankDeposit" : "bankWithdraw", { targetEntityId, itemId, quantity }));
  }

  handleBankClose(targetEntityId: string) {
    // Dismiss presentation immediately; replicated banking state stays authoritative.
    this.dismissedBankTargetId = targetEntityId;
    this.dispatchEvent(new Event(BankUpdateEventName));
    this.wsClient.sendMessage(createCommand("bankClose", { targetEntityId }));
  }

  getBankPanelTargetId(): string | undefined {
    const targetId = this.getMyEntity()?.getComponent("banking")?.targetEntityId;
    return targetId === this.dismissedBankTargetId ? undefined : targetId;
  }

  handleBankResult(payload: BankResultPayload) {
    this.dispatchEvent(new BankResultEvent(payload));
  }

  handleTrade(targetEntityId: string, action: "buy" | "sell", itemId: string) {
    this.wsClient.sendMessage(createCommand("trade", { targetEntityId, action, itemId }));
  }

  handleTradeClose(targetEntityId: string) {
    this.wsClient.sendMessage(createCommand("tradeClose", { targetEntityId }));
  }

  handleTradeResult(payload: TradeResultPayload) {
    this.dispatchEvent(new TradeResultEvent(payload));
  }

  getEntityName(entityId: string, fallback = "Unknown"): string {
    const entity = this.entities.find((e) => e.getId() === entityId);
    const name = entity?.getComponent("metadata")?.name;
    return typeof name === "string" && name.length > 0 ? name : fallback;
  }

  handleInteractionOptionClick(entityId: string, option: string) {
    // Only an explicit new Bank interaction may reopen a locally dismissed panel.
    if (option === "bank") this.dismissedBankTargetId = null;
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

  handleQuestStarted(payload: QuestStartedPayload) {
    this.dispatchEvent(new QuestStartedEvent(payload));
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

  getEntityFocusPoint(entityId: string): THREE.Vector3 | null {
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

  projectInventoryItems<T extends InventoryLayoutItem>(items: T[], capacity: number): T[] {
    return this.inventoryPrediction.project(items, capacity);
  }

  handleInventoryMove(itemId: string, slot: number) {
    const inventory = this.getMyEntity()?.getComponent("inventory");
    if (!this.wsClient.isConnected || !inventory || !Number.isInteger(slot) || slot < 0 ||
        slot >= inventory.width * inventory.height || !inventory.items.some((item: InventoryLayoutItem) => item.id === itemId)) return;
    const move = this.inventoryPrediction.enqueue(itemId, slot);
    this.wsClient.sendMessage(createCommand("inventoryMove", move));
    this.dispatchEvent(new InventoryUpdateEvent());
  }

  handleDropItem(itemId: string) {
    this.wsClient.sendMessage(createCommand("drop", { itemId }));
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
