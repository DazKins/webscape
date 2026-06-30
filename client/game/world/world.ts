import * as THREE from "three";
import Input from "../../input";
import Camera from "../camera";
import { addWallGeometry, type WorldWall } from "../renderer/rendererWall";
import {
  createTerrainSurfaceGeometry,
  getTileHeight,
  sampleTerrainHeight,
  type TerrainHeightGrid,
} from "./terrainHeight";

class World {
  sizeX: number;
  sizeY: number;
  terrain: string[];
  heights: number[];
  blockers: boolean[][];
  walls: WorldWall[];
  input: Input;
  mesh: THREE.Mesh;
  highlightMesh: THREE.Mesh;
  heightGrid: TerrainHeightGrid;

  constructor(
    scene: THREE.Scene,
    sizeX: number,
    sizeY: number,
    terrain: string[],
    heights: number[],
    blockers: boolean[][],
    walls: WorldWall[],
    input: Input
  ) {
    this.sizeX = sizeX;
    this.sizeY = sizeY;
    this.terrain = terrain;
    this.heights = heights;
    this.blockers = blockers;
    this.walls = walls;
    this.input = input;
    this.heightGrid = { sizeX, sizeY, heights };

    this.mesh = new THREE.Mesh(
      createTerrainSurfaceGeometry(this.heightGrid, this.terrain, terrainColor),
      new THREE.MeshPhongMaterial({
        vertexColors: true,
        side: THREE.DoubleSide,
      })
    );
    scene.add(this.mesh);

    this.highlightMesh = new THREE.Mesh(
      new THREE.PlaneGeometry(1, 1),
      new THREE.MeshBasicMaterial({
        color: 0xd8d8d8,
        transparent: true,
        opacity: 0.32,
        side: THREE.DoubleSide,
      })
    );
    this.highlightMesh.rotation.x = -Math.PI / 2;
    this.highlightMesh.position.y = 0.02;
    scene.add(this.highlightMesh);

    addWallGeometry(scene, this.walls);
  }

  getVisualHeightAtWorldPosition(worldX: number, worldZ: number) {
    return sampleTerrainHeight(this.heightGrid, worldX, worldZ);
  }

  getVisualHeightAtTile(tileX: number, tileY: number) {
    return getTileHeight(this.heightGrid, tileX, tileY);
  }

  getHoveredTile(camera: Camera) {
    if (this.input.isPointerBlocked()) {
      return undefined;
    }

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
      this.highlightMesh.visible = true;
      this.highlightMesh.position.x = hoveredTile.x + 0.5;
      this.highlightMesh.position.z = hoveredTile.y + 0.5;
    } else {
      this.highlightMesh.visible = false;
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
