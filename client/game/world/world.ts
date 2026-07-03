import * as THREE from "three";
import Input from "../../input";
import Camera from "../camera";
import { addWallGeometry, type WorldWall } from "../renderer/rendererWall";
import {
  createTerrainSurfaceGeometry,
  createTileHighlightGeometry,
  createWaterSurfaceGeometry,
  getTileHeight,
  sampleTerrainHeight,
  type TerrainHeightGrid,
} from "./terrainHeight";

const WATER_TEXTURE_SCROLL_X = 0.018;
const WATER_TEXTURE_SCROLL_Y = 0.032;

class World {
  sizeX: number;
  sizeY: number;
  terrain: string[];
  heights: number[];
  blockers: boolean[][];
  walls: WorldWall[];
  input: Input;
  mesh: THREE.Mesh;
  waterMesh: THREE.Mesh | undefined;
  waterTexture: THREE.Texture | undefined;
  waterAnimationTime: number;
  highlightMesh: THREE.Mesh;
  heightGrid: TerrainHeightGrid;
  highlightedTile: { x: number; y: number } | undefined;

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
    this.highlightedTile = undefined;
    this.waterAnimationTime = 0;

    this.mesh = new THREE.Mesh(
      createTerrainSurfaceGeometry(this.heightGrid, this.terrain, terrainColor),
      new THREE.MeshPhongMaterial({
        vertexColors: true,
        side: THREE.DoubleSide,
      })
    );
    scene.add(this.mesh);

    const waterGeometry = createWaterSurfaceGeometry(this.heightGrid, this.terrain);
    if ((waterGeometry.getAttribute("position")?.count ?? 0) > 0) {
      this.waterTexture = createWaterRippleTexture();
      this.waterMesh = new THREE.Mesh(
        waterGeometry,
        new THREE.MeshBasicMaterial({
          color: 0xd4f5ff,
          map: this.waterTexture,
          transparent: true,
          opacity: 0.34,
          depthWrite: false,
          side: THREE.DoubleSide,
        })
      );
      this.waterMesh.renderOrder = 1;
      scene.add(this.waterMesh);
    } else {
      waterGeometry.dispose();
    }

    this.highlightMesh = new THREE.Mesh(
      new THREE.PlaneGeometry(1, 1),
      new THREE.MeshBasicMaterial({
        color: 0xd8d8d8,
        transparent: true,
        opacity: 0.32,
        side: THREE.DoubleSide,
      })
    );
    this.highlightMesh.renderOrder = 2;
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

      if (gridX < 0 || gridY < 0 || gridX >= this.sizeX || gridY >= this.sizeY) {
        return undefined;
      }

      return { x: gridX, y: gridY };
    }
  }

  update(camera: Camera, deltaSeconds: number) {
    this.updateWaterAnimation(deltaSeconds);

    const hoveredTile = this.getHoveredTile(camera);
    if (hoveredTile) {
      this.highlightMesh.visible = true;
      if (
        !this.highlightedTile ||
        this.highlightedTile.x !== hoveredTile.x ||
        this.highlightedTile.y !== hoveredTile.y
      ) {
        this.highlightMesh.geometry.dispose();
        this.highlightMesh.geometry = createTileHighlightGeometry(this.heightGrid, hoveredTile.x, hoveredTile.y);
        this.highlightedTile = hoveredTile;
      }
    } else {
      this.highlightMesh.visible = false;
      this.highlightedTile = undefined;
    }
  }

  private updateWaterAnimation(deltaSeconds: number) {
    if (!this.waterTexture) {
      return;
    }

    this.waterAnimationTime += deltaSeconds;
    this.waterTexture.offset.set(
      (this.waterAnimationTime * WATER_TEXTURE_SCROLL_X) % 1,
      (this.waterAnimationTime * WATER_TEXTURE_SCROLL_Y) % 1
    );
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

function createWaterRippleTexture(): THREE.CanvasTexture {
  const size = 128;
  const canvas = document.createElement("canvas");
  canvas.width = size;
  canvas.height = size;

  const context = canvas.getContext("2d");
  if (!context) {
    return new THREE.CanvasTexture(canvas);
  }

  context.clearRect(0, 0, size, size);
  context.lineCap = "round";

  drawRippleLines(context, size, "rgba(255, 255, 255, 0.48)", 2.1, 18, 0);
  drawRippleLines(context, size, "rgba(135, 211, 232, 0.3)", 1.2, 24, 9);

  const texture = new THREE.CanvasTexture(canvas);
  texture.wrapS = THREE.RepeatWrapping;
  texture.wrapT = THREE.RepeatWrapping;
  texture.repeat.set(0.85, 0.85);
  return texture;
}

function drawRippleLines(
  context: CanvasRenderingContext2D,
  size: number,
  strokeStyle: string,
  lineWidth: number,
  spacing: number,
  phase: number
) {
  context.strokeStyle = strokeStyle;
  context.lineWidth = lineWidth;

  for (let y = -size; y < size * 2; y += spacing) {
    context.beginPath();
    for (let x = -16; x <= size + 16; x += 8) {
      const waveY = y + phase + x * 0.16 + Math.sin((x + y) * 0.09) * 3.2;
      if (x === -16) {
        context.moveTo(x, waveY);
      } else {
        context.lineTo(x, waveY);
      }
    }
    context.stroke();
  }
}

export default World;
