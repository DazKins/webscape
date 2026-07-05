import * as THREE from "three";
import Input from "../../input";
import Camera from "../camera";
import type { DeviceProfile, ViewportSize } from "../../responsive";
import { addWallGeometry, type WorldWall } from "../renderer/rendererWall";
import {
  createTerrainSurfaceGeometry,
  createTileHighlightGeometry,
  createWaterSurfaceGeometry,
  getTileHeight,
  sampleTerrainHeight,
  type TerrainHeightGrid,
} from "./terrainHeight";

const WATER_GLINT_SPEED = 1.15;

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
  waterMaterial: THREE.ShaderMaterial | undefined;
  waterAnimationTime: number;
  highlightMesh: THREE.Mesh;
  heightGrid: TerrainHeightGrid;
  highlightedTile: { x: number; y: number } | undefined;
  selectedTileSeconds: number;

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
    this.selectedTileSeconds = 0;

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
      this.waterMaterial = createWaterGlintMaterial();
      this.waterMesh = new THREE.Mesh(
        waterGeometry,
        this.waterMaterial
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

  getPointerTile(camera: Camera, viewport: ViewportSize) {
    if (this.input.isPointerBlocked()) {
      return undefined;
    }

    const pointer = this.input.getPointerPosition();
    const mouseX = (pointer.x / viewport.width) * 2 - 1;
    const mouseY = -(pointer.y / viewport.height) * 2 + 1;

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

  showTileIndicator(tile: { x: number; y: number }) {
    this.highlightMesh.visible = true;
    this.highlightMesh.geometry.dispose();
    this.highlightMesh.geometry = createTileHighlightGeometry(this.heightGrid, tile.x, tile.y);
    this.highlightedTile = tile;
    this.selectedTileSeconds = 0.34;
  }

  update(camera: Camera, deltaSeconds: number, profile: DeviceProfile) {
    this.updateWaterAnimation(deltaSeconds);

    if (profile.canHover && !profile.isCoarsePointer) {
      const hoveredTile = this.getPointerTile(camera, profile);
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
      return;
    }

    if (this.selectedTileSeconds > 0) {
      this.selectedTileSeconds = Math.max(0, this.selectedTileSeconds - deltaSeconds);
      this.highlightMesh.visible = true;
    } else {
      this.highlightMesh.visible = false;
      this.highlightedTile = undefined;
    }
  }

  private updateWaterAnimation(deltaSeconds: number) {
    if (!this.waterMaterial) {
      return;
    }

    this.waterAnimationTime += deltaSeconds;
    this.waterMaterial.uniforms.time.value = this.waterAnimationTime * WATER_GLINT_SPEED;
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

function createWaterGlintMaterial(): THREE.ShaderMaterial {
  return new THREE.ShaderMaterial({
    uniforms: {
      time: { value: 0 },
    },
    vertexShader: `
      varying vec2 vWaterUv;

      void main() {
        vWaterUv = uv;
        gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
      }
    `,
    fragmentShader: `
      uniform float time;
      varying vec2 vWaterUv;

      const float TAU = 6.28318530718;

      float hash(vec2 point) {
        return fract(sin(dot(point, vec2(127.1, 311.7))) * 43758.5453123);
      }

      float noise(vec2 point) {
        vec2 cell = floor(point);
        vec2 local = fract(point);
        local = local * local * (3.0 - 2.0 * local);

        float topLeft = hash(cell);
        float topRight = hash(cell + vec2(1.0, 0.0));
        float bottomLeft = hash(cell + vec2(0.0, 1.0));
        float bottomRight = hash(cell + vec2(1.0, 1.0));

        return mix(
          mix(topLeft, topRight, local.x),
          mix(bottomLeft, bottomRight, local.x),
          local.y
        );
      }

      void main() {
        vec2 p = vWaterUv * 3.4;
        float surface =
          noise(p) * 0.52 +
          noise(p * 2.17 + vec2(13.4, 8.2)) * 0.31 +
          noise(p * 4.61 + vec2(5.7, 21.9)) * 0.17;

        float glintShape = smoothstep(0.58, 0.88, surface);
        float phase = noise(vWaterUv * 0.75 + vec2(19.0, 3.0)) * TAU;
        float pulse = 0.62 + 0.38 * sin(time + phase);
        float glint = glintShape * pulse;

        glint = clamp(glint, 0.0, 1.0);
        vec3 color = mix(vec3(0.58, 0.82, 0.92), vec3(0.96, 0.99, 1.0), glint);
        float alpha = 0.035 + glint * 0.12;

        gl_FragColor = vec4(color, alpha);
      }
    `,
    transparent: true,
    depthWrite: false,
    side: THREE.DoubleSide,
  });
}

export default World;
