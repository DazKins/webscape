import { useEffect, useRef } from "react";
import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import {
  entityPosition,
  entitySize,
  tileIndex,
  type WorldEntity,
  type WorldFormat,
  type WorldWall,
} from "./worldFormat.ts";
import { terrainColor } from "./terrainPalette.ts";

type MapTool = "terrain" | "blocker" | "height" | "wall" | "entity" | "select";
type MapSelection =
  | { type: "wall"; id: string }
  | { type: "entity"; id: string }
  | { type: "tile"; x: number; y: number }
  | null;

type TilePointerHandler = (x: number, y: number) => void;
type TilePointerEnterHandler = (x: number, y: number, buttons: number) => void;

type ViewportState = {
  scene: THREE.Scene;
  camera: THREE.PerspectiveCamera;
  controls: OrbitControls;
  renderer: THREE.WebGLRenderer;
  raycaster: THREE.Raycaster;
  pointer: THREE.Vector2;
  root: THREE.Group;
  tileMeshes: THREE.Mesh[];
  hoveredTile: { x: number; y: number } | null;
  hoverOverlay: THREE.Mesh<THREE.PlaneGeometry, THREE.MeshBasicMaterial>;
  framedSize: string | null;
  frameId: number | null;
  resizeObserver: ResizeObserver;
};

const HEIGHT_SCALE = 0.2;
const TILE_BASE_DEPTH = 0.04;
const MARKER_OFFSET = 0.035;

export function MapViewport3D({
  world,
  selection,
  tool,
  onTilePointerDown,
  onTilePointerEnter,
}: {
  world: WorldFormat;
  selection: MapSelection;
  tool: MapTool;
  onTilePointerDown: TilePointerHandler;
  onTilePointerEnter: TilePointerEnterHandler;
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const stateRef = useRef<ViewportState | null>(null);
  const callbacksRef = useRef({ onTilePointerDown, onTilePointerEnter });
  const worldRef = useRef(world);
  const selectionRef = useRef(selection);

  callbacksRef.current = { onTilePointerDown, onTilePointerEnter };
  worldRef.current = world;
  selectionRef.current = selection;

  useEffect(() => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0xc7c2b3);

    const camera = new THREE.PerspectiveCamera(42, 1, 0.1, 1000);
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    renderer.shadowMap.enabled = false;
    container.appendChild(renderer.domElement);

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.enablePan = false;
    controls.minDistance = 3;
    controls.maxDistance = 60;
    controls.minPolarAngle = 0.18;
    controls.maxPolarAngle = Math.PI * 0.46;
    controls.mouseButtons.LEFT = undefined;
    controls.mouseButtons.MIDDLE = THREE.MOUSE.DOLLY;
    controls.mouseButtons.RIGHT = THREE.MOUSE.ROTATE;

    const root = new THREE.Group();
    scene.add(root);

    const hoverOverlay = createTileOverlayMesh(0xfff2a8, 0.28);
    hoverOverlay.visible = false;
    scene.add(hoverOverlay);

    const ambient = new THREE.HemisphereLight(0xffffff, 0x7c7464, 1.45);
    scene.add(ambient);

    const keyLight = new THREE.DirectionalLight(0xffffff, 1.9);
    keyLight.position.set(-4, 8, 6);
    scene.add(keyLight);

    const state: ViewportState = {
      scene,
      camera,
      controls,
      renderer,
      raycaster: new THREE.Raycaster(),
      pointer: new THREE.Vector2(),
      root,
      tileMeshes: [],
      hoveredTile: null,
      hoverOverlay,
      framedSize: null,
      frameId: null,
      resizeObserver: new ResizeObserver(() => resizeRenderer(container, camera, renderer)),
    };
    stateRef.current = state;

    resizeRenderer(container, camera, renderer);
    state.resizeObserver.observe(container);

    const animate = () => {
      controls.update();
      renderer.render(scene, camera);
      state.frameId = window.requestAnimationFrame(animate);
    };
    animate();

    const canvas = renderer.domElement;
    const handlePointerDown = (event: PointerEvent) => {
      if (event.button !== 0) {
        return;
      }

      const tile = pickTile(event, state);
      if (!tile) {
        return;
      }

      event.preventDefault();
      canvas.setPointerCapture(event.pointerId);
      callbacksRef.current.onTilePointerDown(tile.x, tile.y);
    };

    const handlePointerMove = (event: PointerEvent) => {
      const tile = pickTile(event, state);
      if (!sameTile(tile, state.hoveredTile)) {
        state.hoveredTile = tile;
        updateHoverOverlay(state, worldRef.current);
        if (tile) {
          callbacksRef.current.onTilePointerEnter(tile.x, tile.y, event.buttons);
        }
      }
    };

    const handlePointerLeave = () => {
      if (state.hoveredTile) {
        state.hoveredTile = null;
        updateHoverOverlay(state, worldRef.current);
      }
    };

    const handleContextMenu = (event: MouseEvent) => {
      event.preventDefault();
    };

    canvas.addEventListener("pointerdown", handlePointerDown);
    canvas.addEventListener("pointermove", handlePointerMove);
    canvas.addEventListener("pointerleave", handlePointerLeave);
    canvas.addEventListener("contextmenu", handleContextMenu);

    return () => {
      canvas.removeEventListener("pointerdown", handlePointerDown);
      canvas.removeEventListener("pointermove", handlePointerMove);
      canvas.removeEventListener("pointerleave", handlePointerLeave);
      canvas.removeEventListener("contextmenu", handleContextMenu);
      state.resizeObserver.disconnect();
      if (state.frameId !== null) {
        window.cancelAnimationFrame(state.frameId);
      }
      disposeGroup(root);
      disposeObject(hoverOverlay);
      scene.clear();
      controls.dispose();
      renderer.dispose();
      renderer.domElement.remove();
      stateRef.current = null;
    };
  }, []);

  useEffect(() => {
    const state = stateRef.current;
    if (!state) {
      return;
    }

    const frameKey = `${world.size.x}:${world.size.y}`;
    if (state.framedSize !== frameKey) {
      frameCamera(state.camera, state.controls, world);
      state.framedSize = frameKey;
    }
    renderWorld(state, world, selection);
    updateHoverOverlay(state, world);
  }, [world, selection]);

  return (
    <div
      ref={containerRef}
      className={`mapViewport3D tool-${tool}`}
      aria-label="3D map viewport"
    />
  );
}

function renderWorld(state: ViewportState, world: WorldFormat, selection: MapSelection) {
  disposeGroup(state.root);
  state.root.clear();
  state.tileMeshes = [];

  const terrainMaterialCache = new Map<string, THREE.MeshLambertMaterial[]>();
  for (let y = 0; y < world.size.y; y += 1) {
    for (let x = 0; x < world.size.x; x += 1) {
      const index = tileIndex(world.size, x, y);
      const height = tileVisualHeight(world, x, y);
      const depth = height + TILE_BASE_DEPTH;
      const geometry = new THREE.BoxGeometry(1, depth, 1);
      const terrain = world.terrain[index] ?? "grass";
      const materials = tileMaterials(terrain, terrainMaterialCache);
      const tile = new THREE.Mesh(geometry, materials);
      tile.position.set(x + 0.5, (height - TILE_BASE_DEPTH) / 2, y + 0.5);
      tile.userData = { tileX: x, tileY: y };
      state.root.add(tile);
      state.tileMeshes.push(tile);

      addTileGridLines(state.root, x, y, height);

      if (world.blockers?.[index]) {
        addBlockerMarker(state.root, x, y, height);
      }
    }
  }

  for (const wall of world.walls) {
    addWallMarker(state.root, world, wall, selection?.type === "wall" && selection.id === wall.id);
  }

  for (const entity of world.entities) {
    addEntityMarker(state.root, world, entity, selection?.type === "entity" && selection.id === entity.id);
  }

  if (selection?.type === "tile") {
    addTileOverlay(state.root, world, selection.x, selection.y, 0x36c7d4, 0.36);
    addTileOutline(state.root, world, selection.x, selection.y, 0x0d7180);
  }
}

function tileMaterials(terrain: string, cache: Map<string, THREE.MeshLambertMaterial[]>) {
  const cached = cache.get(terrain);
  if (cached) {
    return cached;
  }

  const color = new THREE.Color(terrainColor(terrain));
  const sideColor = color.clone().multiplyScalar(0.72);
  const topMaterial = new THREE.MeshLambertMaterial({ color });
  const sideMaterial = new THREE.MeshLambertMaterial({ color: sideColor });
  const materials = [
    sideMaterial,
    sideMaterial,
    topMaterial,
    sideMaterial,
    sideMaterial,
    sideMaterial,
  ];
  cache.set(terrain, materials);
  return materials;
}

function addTileGridLines(root: THREE.Group, x: number, y: number, height: number) {
  const geometry = new THREE.BufferGeometry().setFromPoints([
    new THREE.Vector3(x, height + MARKER_OFFSET, y),
    new THREE.Vector3(x + 1, height + MARKER_OFFSET, y),
    new THREE.Vector3(x + 1, height + MARKER_OFFSET, y + 1),
    new THREE.Vector3(x, height + MARKER_OFFSET, y + 1),
    new THREE.Vector3(x, height + MARKER_OFFSET, y),
  ]);
  const line = new THREE.Line(
    geometry,
    new THREE.LineBasicMaterial({ color: 0x222222, transparent: true, opacity: 0.22 })
  );
  root.add(line);
}

function addBlockerMarker(root: THREE.Group, x: number, y: number, height: number) {
  const geometry = new THREE.BoxGeometry(0.56, 0.035, 0.56);
  const material = new THREE.MeshBasicMaterial({ color: 0x202124, transparent: true, opacity: 0.72 });
  const mesh = new THREE.Mesh(geometry, material);
  mesh.position.set(x + 0.5, height + 0.045, y + 0.5);
  root.add(mesh);
}

function addWallMarker(root: THREE.Group, world: WorldFormat, wall: WorldWall, selected: boolean) {
  const height = tileVisualHeight(world, wall.x, wall.y);
  const geometry = new THREE.BoxGeometry(0.86, 0.56, 0.16);
  const material = new THREE.MeshLambertMaterial({ color: wall.type === "wood" ? 0x8a5a34 : 0x6b6f76 });
  const mesh = new THREE.Mesh(geometry, material);
  mesh.position.set(wall.x + 0.5, height + 0.28, wall.y + 0.5);
  mesh.userData = { wallId: wall.id };
  root.add(mesh);

  if (selected) {
    addBoxOutline(root, mesh, 0x0d7180);
  }
}

function addEntityMarker(root: THREE.Group, world: WorldFormat, entity: WorldEntity, selected: boolean) {
  const position = entityPosition(entity);
  if (!position) {
    return;
  }

  const size = entitySize(entity);
  const height = tileVisualHeight(world, position.x, position.y);
  const geometry = new THREE.BoxGeometry(
    Math.max(0.42, size.width * 0.58),
    0.5,
    Math.max(0.42, size.height * 0.58)
  );
  const material = new THREE.MeshLambertMaterial({ color: entityMarkerColor(entity) });
  const mesh = new THREE.Mesh(geometry, material);
  mesh.position.set(position.x + size.width / 2, height + 0.25, position.y + size.height / 2);
  mesh.userData = { entityId: entity.id };
  root.add(mesh);

  if (selected) {
    addBoxOutline(root, mesh, 0x0d7180);
  }
}

function addTileOverlay(
  root: THREE.Group,
  world: WorldFormat,
  x: number,
  y: number,
  color: THREE.ColorRepresentation,
  opacity: number
) {
  const height = tileVisualHeight(world, x, y);
  const geometry = new THREE.PlaneGeometry(1, 1);
  const material = new THREE.MeshBasicMaterial({
    color,
    transparent: true,
    opacity,
    side: THREE.DoubleSide,
    depthWrite: false,
  });
  const mesh = new THREE.Mesh(geometry, material);
  mesh.rotation.x = -Math.PI / 2;
  mesh.position.set(x + 0.5, height + MARKER_OFFSET + 0.006, y + 0.5);
  root.add(mesh);
}

function createTileOverlayMesh(color: THREE.ColorRepresentation, opacity: number) {
  const geometry = new THREE.PlaneGeometry(1, 1);
  const material = new THREE.MeshBasicMaterial({
    color,
    transparent: true,
    opacity,
    side: THREE.DoubleSide,
    depthWrite: false,
  });
  const mesh = new THREE.Mesh(geometry, material);
  mesh.rotation.x = -Math.PI / 2;
  return mesh;
}

function updateHoverOverlay(state: ViewportState, world: WorldFormat) {
  const tile = state.hoveredTile;
  if (!tile || tile.x < 0 || tile.y < 0 || tile.x >= world.size.x || tile.y >= world.size.y) {
    state.hoverOverlay.visible = false;
    return;
  }

  const height = tileVisualHeight(world, tile.x, tile.y);
  state.hoverOverlay.position.set(tile.x + 0.5, height + MARKER_OFFSET + 0.006, tile.y + 0.5);
  state.hoverOverlay.visible = true;
}

function addTileOutline(root: THREE.Group, world: WorldFormat, x: number, y: number, color: THREE.ColorRepresentation) {
  const height = tileVisualHeight(world, x, y) + MARKER_OFFSET + 0.014;
  const points = [
    new THREE.Vector3(x, height, y),
    new THREE.Vector3(x + 1, height, y),
    new THREE.Vector3(x + 1, height, y + 1),
    new THREE.Vector3(x, height, y + 1),
    new THREE.Vector3(x, height, y),
  ];
  const geometry = new THREE.BufferGeometry().setFromPoints(points);
  root.add(new THREE.Line(geometry, new THREE.LineBasicMaterial({ color, linewidth: 2 })));
}

function addBoxOutline(root: THREE.Group, mesh: THREE.Mesh, color: THREE.ColorRepresentation) {
  const edges = new THREE.EdgesGeometry(mesh.geometry);
  const line = new THREE.LineSegments(edges, new THREE.LineBasicMaterial({ color }));
  line.position.copy(mesh.position);
  line.rotation.copy(mesh.rotation);
  line.scale.copy(mesh.scale).multiplyScalar(1.04);
  root.add(line);
}

function pickTile(event: PointerEvent, state: ViewportState) {
  const rect = state.renderer.domElement.getBoundingClientRect();
  state.pointer.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
  state.pointer.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
  state.raycaster.setFromCamera(state.pointer, state.camera);
  const hits = state.raycaster.intersectObjects(state.tileMeshes, false);
  const hit = hits[0]?.object;
  const x = hit?.userData.tileX;
  const y = hit?.userData.tileY;

  if (Number.isInteger(x) && Number.isInteger(y)) {
    return { x: Number(x), y: Number(y) };
  }

  return null;
}

function frameCamera(camera: THREE.PerspectiveCamera, controls: OrbitControls, world: WorldFormat) {
  const centerX = world.size.x / 2;
  const centerZ = world.size.y / 2;
  const span = Math.max(world.size.x, world.size.y, 4);
  camera.position.set(centerX + span * 0.72, span * 0.95, centerZ + span * 1.18);
  controls.target.set(centerX, 0, centerZ);
  controls.update();
}

function resizeRenderer(
  container: HTMLDivElement,
  camera: THREE.PerspectiveCamera,
  renderer: THREE.WebGLRenderer
) {
  const width = Math.max(1, container.clientWidth);
  const height = Math.max(1, container.clientHeight);
  camera.aspect = width / height;
  camera.updateProjectionMatrix();
  renderer.setSize(width, height, false);
}

function tileVisualHeight(world: WorldFormat, x: number, y: number) {
  const index = tileIndex(world.size, x, y);
  const height = world.heights[index] ?? 0;
  if (!Number.isInteger(height) || height < 0 || height > 10) {
    return 0;
  }
  return height * HEIGHT_SCALE;
}

function entityMarkerColor(entity: WorldEntity) {
  const label = entity.id;
  let hash = 0;
  for (let i = 0; i < label.length; i += 1) {
    hash = label.charCodeAt(i) + ((hash << 5) - hash);
  }
  return new THREE.Color(`hsl(${Math.abs(hash) % 360} 42% 43%)`);
}

function sameTile(a: { x: number; y: number } | null, b: { x: number; y: number } | null) {
  return a?.x === b?.x && a?.y === b?.y;
}

function disposeGroup(group: THREE.Group) {
  for (const child of group.children) {
    child.traverse((object) => {
      disposeObject(object);
    });
  }
}

function disposeObject(object: THREE.Object3D) {
  if (object instanceof THREE.Mesh || object instanceof THREE.Line || object instanceof THREE.LineSegments) {
    object.geometry.dispose();
    const material = object.material;
    if (Array.isArray(material)) {
      for (const item of material) {
        item.dispose();
      }
    } else {
      material.dispose();
    }
  }
}
