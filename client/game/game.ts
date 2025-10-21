import Entity from "./entity/entity.ts";
import World from "./world/world.ts";
import Input from "../input.ts";
import addReferenceGeometry from "./referenceGeometry.ts";
import { createCommand } from "../command/command.ts";
import EntityInteractionBox from "./ui/entityInteractionBox";
import * as THREE from "three";
import { WebSocketClient } from "../ws.ts";
import { CSS2DRenderer } from "three/examples/jsm/Addons.js";
import ChatBox from "./ui/chatBox.ts";

class Game {
  wsClient!: WebSocketClient;
  myPlayerId!: string;
  scene: THREE.Scene;
  camera: THREE.PerspectiveCamera;
  renderer: THREE.WebGLRenderer;
  cssRenderer2d: CSS2DRenderer;
  myLocationHighlightMesh: THREE.Mesh;
  entities: Entity[];

  cameraDistance: number;
  cameraAngle: number;
  cameraHeight: number;
  orbitSpeed: number;
  heightSpeed: number;
  minHeight: number;
  maxHeight: number;

  entityInteractionBox: EntityInteractionBox;
  input: Input;
  world!: World;
  chatBox: ChatBox;

  constructor() {
    this.scene = new THREE.Scene();
    this.camera = new THREE.PerspectiveCamera(
      75,
      window.innerWidth / window.innerHeight,
      0.1,
      1000
    );
    this.camera.position.z = 5;
    this.camera.position.y = 5;

    this.renderer = new THREE.WebGLRenderer({ antialias: true });
    this.renderer.setSize(window.innerWidth, window.innerHeight);
    this.renderer.setClearColor(0x87ceeb);
    document.body.appendChild(this.renderer.domElement);

    this.cssRenderer2d = new CSS2DRenderer();
    this.cssRenderer2d.setSize(window.innerWidth, window.innerHeight);
    this.cssRenderer2d.domElement.style.position = "absolute";
    this.cssRenderer2d.domElement.style.top = "0";
    this.cssRenderer2d.domElement.style.pointerEvents = "none";
    document.body.appendChild(this.cssRenderer2d.domElement);

    this.scene.add(new THREE.AmbientLight(0xffffff, 1.0));
    const light = new THREE.DirectionalLight(0xffffff, 0.8);
    light.position.set(10, 10, 10);
    this.scene.add(light);

    addReferenceGeometry(this.scene);

    this.onWindowResize = this.onWindowResize.bind(this);
    window.addEventListener("resize", this.onWindowResize, false);

    this.cameraDistance = 5;
    this.cameraAngle = 0;
    this.cameraHeight = 5;
    this.orbitSpeed = 0.05;
    this.heightSpeed = 0.1;
    this.minHeight = 2;
    this.maxHeight = 10;

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

    this.input = new Input();

    this.entityInteractionBox = new EntityInteractionBox();

    this.input.registerClickCallback(() => {
      this.entityInteractionBox.hide();
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

    this.input.registerRightClickCallback((event) => {
      this.entityInteractionBox.hide();
      const mouse = this.input.getMousePosition();
      const mouseX = (mouse.x / window.innerWidth) * 2 - 1;
      const mouseY = -(mouse.y / window.innerHeight) * 2 + 1;

      const raycaster = new THREE.Raycaster();
      raycaster.setFromCamera(new THREE.Vector2(mouseX, mouseY), this.camera);
      const intersects = raycaster.intersectObjects(
        this.entities.map((e) => e.mesh),
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
        const entity = this.entities.find((e) => e.id === entityId);
        if (entity) {
          this.entityInteractionBox.show(
            entity,
            event.clientX,
            event.clientY,
            entity.name || "Unnamed?!?!"
          );
        }
      }
    });

    this.chatBox = new ChatBox();
    this.chatBox.show();
  }

  registerWsClient(wsClient: WebSocketClient) {
    this.wsClient = wsClient;
  }

  cameraTargetX = 0.0;
  cameraTargetY = 0.0;

  updateCamera() {
    const myEntity = this.getMyEntity();
    if (!myEntity) return;

    if (this.input.getKey("arrowleft")) {
      this.cameraAngle += this.orbitSpeed;
    }
    if (this.input.getKey("arrowright")) {
      this.cameraAngle -= this.orbitSpeed;
    }

    if (this.input.getKey("arrowup")) {
      this.cameraHeight = Math.min(
        this.cameraHeight + this.heightSpeed,
        this.maxHeight
      );
    }
    if (this.input.getKey("arrowdown")) {
      this.cameraHeight = Math.max(
        this.cameraHeight - this.heightSpeed,
        this.minHeight
      );
    }

    const targetX = myEntity.mesh.position.x;
    const targetY = myEntity.mesh.position.z;

    this.cameraTargetX += (targetX - this.cameraTargetX) * 0.1;
    this.cameraTargetY += (targetY - this.cameraTargetY) * 0.1;

    this.camera.position.x =
      this.cameraTargetX + Math.cos(this.cameraAngle) * this.cameraDistance;
    this.camera.position.z =
      this.cameraTargetY + Math.sin(this.cameraAngle) * this.cameraDistance;
    this.camera.position.y = this.cameraHeight;

    this.camera.lookAt(this.cameraTargetX, 0, this.cameraTargetY);
  }

  addEntity(entity: Entity) {
    this.entities.push(entity);
  }

  handleEntityUpdate(entityUpdate: any) {
    let entity = this.entities.find((e) => e.id === entityUpdate.id);

    if (!entity) {
      entity = new Entity(entityUpdate.id, this.scene);
      this.addEntity(entity);
    }

    entity.handleEntityUpdate(entityUpdate);
  }

  handleEntityRemove(entityId: string) {
    console.log("Removing entity", entityId);
    const entity = this.entities.find((e) => e.id === entityId);
    if (entity) {
      entity.remove();
      this.entities = this.entities.filter((e) => e.id !== entityId);
    } else {
      console.log("Entity not found", entityId);
    }
  }

  update() {
    this.updateCamera();

    const myEntity = this.getMyEntity();
    if (myEntity) {
      this.myLocationHighlightMesh.position.x = myEntity.positionX + 0.5;
      this.myLocationHighlightMesh.position.z = myEntity.positionY + 0.5;
    }

    const typedChatText = this.input.getTypedCharsBuffer();
    this.chatBox.update(typedChatText);

    if (this.input.getKeyJustDown("enter") && typedChatText.length > 0) {
      this.wsClient.sendMessage(
        createCommand("chat", {
          message: typedChatText,
        })
      );
      this.input.clearTypedCharsBuffer();
    }

    this.entities.forEach((e) => e.update());
    this.renderer.render(this.scene, this.camera);
    this.cssRenderer2d.render(this.scene, this.camera);
    if (this.world) {
      this.world.update(this.camera);
    }

    this.input.flush();
  }

  onWindowResize() {
    this.camera.aspect = window.innerWidth / window.innerHeight;
    this.camera.updateProjectionMatrix();
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
    return this.entities.find((e) => e.id === this.myPlayerId);
  }
}

export default Game;
