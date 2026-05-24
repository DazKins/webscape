import * as THREE from "three";
import Input from "../../input";
import Camera from "../camera";

class World {
  sizeX: number;
  sizeY: number;
  terrain: string[];
  walls: boolean[][];
  input: Input;
  mesh: THREE.Mesh;
  highlightMesh: THREE.Mesh;

  constructor(
    scene: THREE.Scene,
    sizeX: number,
    sizeY: number,
    terrain: string[],
    walls: boolean[][],
    input: Input
  ) {
    console.log("Creating world", sizeX, sizeY);

    this.sizeX = sizeX;
    this.sizeY = sizeY;
    this.terrain = terrain;
    this.walls = walls;
    this.input = input;

    this.mesh = new THREE.Mesh(
      new THREE.PlaneGeometry(sizeX, sizeY),
      new THREE.MeshPhongMaterial({
        color: 0xe77d11,
        side: THREE.DoubleSide,
      })
    );
    this.mesh.rotation.x = -Math.PI / 2;
    this.mesh.position.x = sizeX / 2;
    this.mesh.position.z = sizeY / 2;
    scene.add(this.mesh);

    this.highlightMesh = new THREE.Mesh(
      new THREE.PlaneGeometry(1, 1),
      new THREE.MeshBasicMaterial({
        color: 0x572e05,
        transparent: true,
        opacity: 0.5,
        side: THREE.DoubleSide,
      })
    );
    this.highlightMesh.rotation.x = -Math.PI / 2;
    this.highlightMesh.position.y = 0.02;
    scene.add(this.highlightMesh);

    this.terrain.forEach((terrainType, index) => {
      const x = index % sizeX;
      const y = Math.floor(index / sizeX);
      const terrainMesh = new THREE.Mesh(
        new THREE.PlaneGeometry(1, 1),
        new THREE.MeshPhongMaterial({
          color: terrainColor(terrainType),
          side: THREE.DoubleSide,
        })
      );
      terrainMesh.rotation.x = -Math.PI / 2;
      terrainMesh.position.x = x + 0.5;
      terrainMesh.position.z = y + 0.5;
      terrainMesh.position.y = 0.005;
      scene.add(terrainMesh);
    });

    this.walls.forEach((row, x) => {
      row.forEach((wall, y) => {
        if (wall === false) return;

        // Create the main wall mesh with pastel grey color
        const wallMesh = new THREE.Mesh(
          new THREE.BoxGeometry(1, 1, 1),
          new THREE.MeshPhongMaterial({ color: 0xd3d3d3 })
        );

        // Create wireframe for black edges
        // const wireframeGeometry = new THREE.EdgesGeometry(
        //   new THREE.BoxGeometry(1, 1, 1)
        // );
        // const wireframeMaterial = new THREE.LineBasicMaterial({
        //   color: 0x000000,
        //   linewidth: 2,
        //   transparent: true,
        //   opacity: 1.0,
        // });
        // const wireframe = new THREE.LineSegments(
        //   wireframeGeometry,
        //   wireframeMaterial
        // );

        // // Add wireframe to the wall mesh
        // wallMesh.add(wireframe);
        wallMesh.position.x = x + 0.5;
        wallMesh.position.z = y + 0.5;
        wallMesh.position.y = 0.5;
        scene.add(wallMesh);
      });
    });
  }

  getHoveredTile(camera: Camera) {
    const mouse = this.input.getMousePosition();
    const mouseX = (mouse.x / window.innerWidth) * 2 - 1;
    const mouseY = -(mouse.y / window.innerHeight) * 2 + 1;

    const raycaster = new THREE.Raycaster();
    raycaster.setFromCamera(new THREE.Vector2(mouseX, mouseY), camera.getInnerCamera());

    const intersects = raycaster.intersectObject(this.mesh);

    if (intersects.length > 0) {
      const point = intersects[0].point;

      const gridX = Math.floor(point.x);
      const gridY = Math.floor(point.z);

      return { x: gridX, y: gridY };
    }
  }

  update(camera: Camera) {
    const hoveredTile = this.getHoveredTile(camera);
    if (hoveredTile) {
      this.highlightMesh.position.x = hoveredTile.x + 0.5;
      this.highlightMesh.position.z = hoveredTile.y + 0.5;
    }
  }
}

function terrainColor(terrainType: string) {
  switch (terrainType) {
    case "grass":
      return 0x73964f;
    case "dirt":
      return 0x9a6b42;
    case "road":
      return 0xb8ab88;
    case "water":
      return 0x4f8fb8;
    case "stone":
      return 0x8b9296;
    default:
      return 0xe77d11;
  }
}

export default World;
