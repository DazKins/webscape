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

class Game extends EventTarget implements InputReceiver {
  wsClient!: WebSocketClient;
  myPlayerId!: string;
  scene: THREE.Scene;
  camera: Camera;
  renderer: THREE.WebGLRenderer;
  cssRenderer2d: CSS2DRenderer;
  myLocationHighlightMesh: THREE.Mesh;
  entities: Entity[];
  entityRenderSystem: EntityRenderSystem;

  input: Input;
  world!: World;

  typedChatText: string;

  constructor(sceneLayerRoot: HTMLElement, hudLayerRoot: HTMLElement) {
    super();

    this.scene = new THREE.Scene();
    this.input = new Input();
    this.camera = new Camera(this.input);
    this.entityRenderSystem = new EntityRenderSystem(this.scene);

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

    this.myLocationHighlightMesh = new THREE.Mesh(
      new THREE.PlaneGeometry(1, 1),
      new THREE.MeshBasicMaterial({
        color: 0x572e05,
        transparent: true,
        opacity: 0.5,
        side: THREE.DoubleSide,
      })
    );
    this.myLocationHighlightMesh.rotation.x = -Math.PI / 2;
    this.myLocationHighlightMesh.position.y = 0.02;
    this.scene.add(this.myLocationHighlightMesh);

    this.entities = [];

    this.input.registerClickCallback(() => {
      const hoveredTile = this.world.getHoveredTile(this.camera);
      if (hoveredTile) {
        this.wsClient.sendMessage(
          createCommand("move", {
            x: hoveredTile.x,
            y: hoveredTile.y,
          })
        );
      }
    });

    this.input.registerRightClickCallback((event: MouseEvent) => {
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
          this.dispatchEvent(
            new InteractionMenuOpenEvent(
              entity.getId(),
              entity.getComponent("metadata")?.name || "Unnamed?!?!",
              entity.getComponent("interactable")?.interactionOptions || [],
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

  updateCamera() {
    const myEntity = this.getMyEntity();
    if (!myEntity) return;

    const myEntityRenderer = this.entityRenderSystem.getRenderers()[myEntity.getId()];
    if (!myEntityRenderer) {
      return;
    }

    const myObject3D = myEntityRenderer.getObject3D()!;

    this.camera.update(
      new THREE.Vector3(myObject3D.position.x + 0.5, 0, myObject3D.position.z + 0.5)
    );
  }

  handleGameUpdate(gameUpdate: any) {
    console.log("Game update");

    const entityComponentsUpdates = gameUpdate.entities;

    for (const entityComponentUpdate of entityComponentsUpdates) {
      const entityId = entityComponentUpdate.entityId;
      const componentId = entityComponentUpdate.componentId;
      const data = entityComponentUpdate.data;

      let localEntity = this.entities.find((e) => e.getId() === entityId);
      if (!localEntity) {
        console.log("Adding new entity", entityId);
        localEntity = new Entity(entityId);
        this.entities.push(localEntity);
      }

      if (componentId === "chatmessage" && data) {
        console.log("Chat message", data.message);

        const fromEntityId = data.fromEntityId;

        const fromEntity = this.entities.find((e) => e.getId() === fromEntityId);
        if (!fromEntity) {
          console.log("From entity not found", fromEntityId);
          continue;
        }

        const fromEntityMetadata = fromEntity.getComponent("metadata");
        if (!fromEntityMetadata) {
          console.log("From entity metadata not found", fromEntityId);
          continue;
        }

        this.dispatchEvent(
          new ChatMessageEvent(data.message, fromEntityMetadata.name ?? "Unknown")
        );
      }

      if (data === null) {
        console.log("Removing component", { entityId, componentId });
        localEntity.removeComponent(componentId);
        continue;
      }

      console.log(`Updating component`, { entityId, componentId, data });

      localEntity.updateComponent(componentId, data);

      // Dispatch inventory update event if this is the player's inventory
      if (componentId === "inventory" && entityId === this.myPlayerId) {
        this.dispatchEvent(new InventoryUpdateEvent());
      }
    }

    const emptyEntities = this.entities.filter((e) => e.isEmpty());
    if (emptyEntities.length > 0) {
      console.log("Removing empty entities", emptyEntities.map((e) => e.getId()));
      this.entities = this.entities.filter((e) => !e.isEmpty());
    }
  }

  handleEntityRemove(entityId: string) {
    console.log("Removing entity", entityId);
    const entity = this.entities.find((e) => e.getId() === entityId);
    if (entity) {
      this.entities = this.entities.filter((e) => e.getId() !== entityId);
    } else {
      console.log("Entity not found", entityId);
    }
  }

  update() {
    this.updateCamera();

    const myEntity = this.getMyEntity();
    if (myEntity) {
      const positionComponent = myEntity.getComponent("position");
      if (positionComponent) {
        this.myLocationHighlightMesh.position.x = positionComponent.x + 0.5;
        this.myLocationHighlightMesh.position.z = positionComponent.y + 0.5;
      }
    }

    this.entityRenderSystem.update(this.entities);

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
    this.world = new World(
      this.scene,
      worldUpdate.sizeX,
      worldUpdate.sizeY,
      worldUpdate.walls,
      this.input
    );
  }

  getMyEntity(): Entity | undefined {
    return this.entities.find((e) => e.getId() === this.myPlayerId);
  }

  handleInteractionOptionClick(entityId: string, option: string) {
    this.wsClient.sendMessage(
      createCommand("interact", {
        entityId,
        option,
      })
    );
  }
}

export default Game;
