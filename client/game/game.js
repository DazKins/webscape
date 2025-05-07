import Entity from "./entity/entity.js";
import World from "./world/world.js";
import Input from "../input.js";
import addReferenceGeometry from "./referenceGeometry.js";
import { createMessage } from "../message/message.js";

class Game {
  constructor(wsClient) {
    this.wsClient = wsClient;
    this.scene = new THREE.Scene();
    this.camera = new THREE.PerspectiveCamera(
      75,
      window.innerWidth / window.innerHeight,
      0.1,
      1000
    );
    this.camera.position.z = 5;
    this.camera.position.y = 5;

    this.renderer = new THREE.WebGLRenderer();
    this.renderer.setSize(window.innerWidth, window.innerHeight);
    this.renderer.setClearColor(0x87ceeb);
    document.body.appendChild(this.renderer.domElement);

    this.scene.add(new THREE.AmbientLight(0xffffff, 0.5));
    this.scene.add(new THREE.DirectionalLight(0xffffff, 0.8));

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

    this.input = new Input();

    this.input.registerClickCallback(() => {
      const hoveredTile = this.world.getHoveredTile(this.camera);
      if (hoveredTile) {
        this.wsClient.sendMessage(
          createMessage("move", {
            x: hoveredTile.x,
            y: hoveredTile.y,
          })
        );
      }
    });

    this.entities = [];
  }

  registerWsClient(wsClient) {
    this.wsClient = wsClient;
  }

  cameraTargetX = 0.0;
  cameraTargetY = 0.0;

  updateCamera() {
    const myEntity = this.getMyEntity();
    if (!myEntity) return;

    if (this.input.getKey("ArrowLeft")) {
      this.cameraAngle += this.orbitSpeed;
    }
    if (this.input.getKey("ArrowRight")) {
      this.cameraAngle -= this.orbitSpeed;
    }

    if (this.input.getKey("ArrowUp")) {
      this.cameraHeight = Math.min(
        this.cameraHeight + this.heightSpeed,
        this.maxHeight
      );
    }
    if (this.input.getKey("ArrowDown")) {
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

  addEntity(entity) {
    this.entities.push(entity);
  }

  handleEntityUpdate(entityUpdate) {
    let entity = this.entities.find((e) => e.id === entityUpdate.entityId);

    if (!entity) {
      entity = new Entity(entityUpdate.entityId, this.scene);
      this.addEntity(entity);
    }

    entity.update(entityUpdate, this.world);
  }

  handleEntityRemove(entityId) {
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
    this.renderer.render(this.scene, this.camera);
    if (this.world) {
      this.world.update(this.camera);
    }
  }

  onWindowResize() {
    this.camera.aspect = window.innerWidth / window.innerHeight;
    this.camera.updateProjectionMatrix();
    this.renderer.setSize(window.innerWidth, window.innerHeight);
  }

  registerWorld(worldUpdate) {
    this.world = new World(
      this.scene,
      worldUpdate.sizeX,
      worldUpdate.sizeY,
      this.input
    );
  }

  registerMyId(entityId) {
    this.myId = entityId;
  }

  getMyEntity() {
    return this.entities.find((e) => e.id === this.myId);
  }
}

export default Game;
