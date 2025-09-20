import Entity from "./entity/entity.js";
import World from "./world/world.js";
import Input from "../input.js";
import addReferenceGeometry from "./referenceGeometry.js";
import { createMessage } from "../message/message.js";

class Game {
  constructor(myPlayerId) {
    this.wsClient = null;

    this.myPlayerId = myPlayerId;
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

    this.myLocationHighlightMesh = new THREE.Mesh(
      new THREE.PlaneGeometry(1, 1),
      new THREE.MeshBasicMaterial({
        color: 0x388e3c,
        transparent: true,
        opacity: 0.5,
        side: THREE.DoubleSide,
      })
    );
    this.myLocationHighlightMesh.rotation.x = -Math.PI / 2;
    this.myLocationHighlightMesh.position.y = 0.02;
    this.scene.add(this.myLocationHighlightMesh);

    this.entities = [];

    const uiBox = document.createElement("div");
    uiBox.style.position = "absolute";
    uiBox.style.background = "rgba(0, 0, 0, 1.0)";
    uiBox.style.color = "white";
    uiBox.style.padding = "10px";
    uiBox.style.display = "none"; // Hidden by default
    uiBox.style.border = "1px solid white";
    uiBox.style.userSelect = "none"; // Disable text selection
    uiBox.style.webkitUserSelect = "none"; // For Safari
    uiBox.style.mozUserSelect = "none"; // For Firefox

    uiBox.addEventListener("click", (event) => {
      event.stopPropagation();
    });

    document.body.appendChild(uiBox);

    this.input = new Input();

    this.input.registerClickCallback(() => {
      uiBox.style.display = "none";
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

    this.input.registerRightClickCallback((event) => {
      uiBox.style.display = "none";
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
        let hitMesh = intersects[0].object;

        while (hitMesh && !hitMesh.userData.entityId) {
          hitMesh = hitMesh.parent;
        }

        const entityId = hitMesh.userData.entityId;
        const entity = this.entities.find((e) => e.id === entityId);
        if (entity) {
          uiBox.style.left = `${event.clientX}px`;
          uiBox.style.top = `${event.clientY}px`;
          uiBox.style.display = "block";
          uiBox.textContent = `Entity: ${entity.name || "Unnamed"}`; // Customize content
          console.log(`Hit entity: ${entity.name}`);
        }
      }
    });
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

    entity.handleEntityUpdate(entityUpdate, this.world);
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
    this.myLocationHighlightMesh.position.x =
      this.getMyEntity().positionX + 0.5;
    this.myLocationHighlightMesh.position.z =
      this.getMyEntity().positionY + 0.5;
    this.entities.forEach((e) => e.update());
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

  getMyEntity() {
    return this.entities.find((e) => e.id === this.myPlayerId);
  }
}

export default Game;
