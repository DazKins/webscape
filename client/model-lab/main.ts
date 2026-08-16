import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import { createModel, isModelName, modelNames, type ModelInstance, type ModelName } from "../game/models";
import { applyModelTransform, getEquipmentPresentation } from "../game/models/equipment";
import "./style.css";

type CameraView = "front" | "back" | "side" | "top" | "three-quarter";

const CAPTURE_SIZE = 800;
const params = new URLSearchParams(window.location.search);
const captureMode = params.get("capture") === "1";
const requestedModel = params.get("model") ?? "human";
const requestedEquipment = params.get("equipment");
let modelName: ModelName = isModelName(requestedModel) ? requestedModel : "human";
let animationName = params.get("animation") ?? "";
let phase = clampPhase(Number(params.get("phase") ?? 0));
let cameraView = parseView(params.get("view"));
let playing = !captureMode && params.get("play") !== "0";
let modelInstance: ModelInstance;
let equipmentInstance: ModelInstance | undefined;
let previousFrame: number | null = null;

const viewport = requireElement<HTMLElement>("viewport");
const modelSelect = requireElement<HTMLSelectElement>("model");
const animationSelect = requireElement<HTMLSelectElement>("animation");
const playPause = requireElement<HTMLButtonElement>("playPause");
const phaseInput = requireElement<HTMLInputElement>("phase");
const phaseValue = requireElement<HTMLOutputElement>("phaseValue");

if (captureMode) {
  document.body.classList.add("capture");
}

const scene = new THREE.Scene();
scene.background = new THREE.Color(0xdce4e8);
scene.add(new THREE.HemisphereLight(0xf7fbff, 0x6f6758, 1.55));
const keyLight = new THREE.DirectionalLight(0xffffff, 2.1);
keyLight.position.set(3.5, 5.5, 4.5);
scene.add(keyLight);
const fillLight = new THREE.DirectionalLight(0xc8dcff, 0.75);
fillLight.position.set(-4, 2.5, -2);
scene.add(fillLight);

const grid = new THREE.GridHelper(8, 16, 0x65727d, 0xa7b0b7);
scene.add(grid);
const axes = new THREE.AxesHelper(0.75);
axes.position.y = 0.004;
scene.add(axes);

const camera = new THREE.PerspectiveCamera(34, 1, 0.05, 100);
const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: captureMode });
renderer.outputColorSpace = THREE.SRGBColorSpace;
renderer.setPixelRatio(captureMode ? 1 : Math.min(window.devicePixelRatio || 1, 2));
viewport.appendChild(renderer.domElement);

const orbit = new OrbitControls(camera, renderer.domElement);
orbit.enableDamping = !captureMode;
orbit.enabled = !captureMode;
orbit.minPolarAngle = 0.08;
orbit.maxPolarAngle = Math.PI / 2 - 0.02;

for (const name of modelNames) {
  modelSelect.add(new Option(name, name));
}
modelSelect.value = modelName;

loadModel();
resize();

modelSelect.addEventListener("change", () => {
  const value = modelSelect.value;
  if (!isModelName(value)) {
    return;
  }
  modelName = value;
  animationName = "";
  phase = 0;
  loadModel();
  updateUrl();
});

animationSelect.addEventListener("change", () => {
  animationName = animationSelect.value;
  phase = 0;
  if (animationName) {
    modelInstance.seek(animationName, phase);
  }
  updateControls();
  updateUrl();
});

playPause.addEventListener("click", () => {
  playing = !playing;
  updateControls();
  updateUrl();
});

phaseInput.addEventListener("input", () => {
  phase = clampPhase(Number(phaseInput.value));
  playing = false;
  if (animationName) {
    modelInstance.seek(animationName, phase);
  }
  updateControls();
  updateUrl();
});

for (const button of document.querySelectorAll<HTMLButtonElement>("[data-view]")) {
  button.addEventListener("click", () => {
    cameraView = parseView(button.dataset.view ?? null);
    frameCamera();
    updateUrl();
  });
}

window.addEventListener("resize", resize);
window.__MODEL_LAB_MODELS__ = [...modelNames];
requestAnimationFrame(animate);

function loadModel() {
  equipmentInstance?.dispose();
  equipmentInstance = undefined;
  modelInstance?.dispose();
  modelInstance = createModel(modelName, modelName === "building" ? { width: 2, height: 2 } : {});
  scene.add(modelInstance.root);
  attachRequestedEquipment();

  animationSelect.replaceChildren(new Option("None", ""));
  for (const animation of modelInstance.animations) {
    animationSelect.add(new Option(animation.name, animation.name));
  }
  if (!modelInstance.animations.some((candidate) => candidate.name === animationName)) {
    animationName = modelInstance.animations[0]?.name ?? "";
  }
  animationSelect.value = animationName;
  frameCamera();
  if (animationName) {
    modelInstance.seek(animationName, phase);
  }
  updateControls();
}

function animate(frameTime: number) {
  requestAnimationFrame(animate);
  const deltaSeconds = previousFrame === null ? 0 : Math.min((frameTime - previousFrame) / 1000, 0.1);
  previousFrame = frameTime;

  const animation = modelInstance.animations.find((candidate) => candidate.name === animationName);
  if (playing && animation) {
    modelInstance.update(deltaSeconds);
    phase += deltaSeconds / animation.duration;
    phase = animation.loop ? phase % 1 : Math.min(1, phase);
    updateControls();
  }
  equipmentInstance?.update(deltaSeconds);

  orbit.update();
  renderer.render(scene, camera);
  if (!window.__MODEL_LAB_READY__) {
    requestAnimationFrame(() => {
      renderer.render(scene, camera);
      window.__MODEL_LAB_READY__ = true;
      document.documentElement.dataset.modelLabReady = "true";
    });
  }
}

function frameCamera() {
  const bounds = new THREE.Box3().setFromObject(modelInstance.root);
  const size = bounds.getSize(new THREE.Vector3());
  const center = bounds.getCenter(new THREE.Vector3());
  const extent = Math.max(size.x, size.y, size.z, 1);
  const distance = extent * 2.8;
  orbit.target.set(center.x, Math.max(size.y * 0.42, center.y), center.z);

  const positions: Record<CameraView, THREE.Vector3> = {
    front: new THREE.Vector3(0, size.y * 0.58, distance),
    back: new THREE.Vector3(0, size.y * 0.58, -distance),
    side: new THREE.Vector3(distance, size.y * 0.58, 0),
    top: new THREE.Vector3(0.01, distance, 0.01),
    "three-quarter": new THREE.Vector3(distance * 0.72, distance * 0.58, distance * 0.72),
  };
  camera.position.copy(orbit.target).add(positions[cameraView]);
  camera.near = Math.max(0.01, distance / 100);
  camera.far = distance * 20;
  camera.updateProjectionMatrix();
  orbit.update();
}

function resize() {
  const width = captureMode ? CAPTURE_SIZE : Math.max(1, viewport.clientWidth);
  const height = captureMode ? CAPTURE_SIZE : Math.max(1, viewport.clientHeight);
  renderer.setSize(width, height, false);
  camera.aspect = width / height;
  camera.updateProjectionMatrix();
}

function updateControls() {
  modelSelect.value = modelName;
  animationSelect.value = animationName;
  phaseInput.disabled = animationName === "";
  phaseInput.value = phase.toFixed(3);
  phaseValue.value = phase.toFixed(3);
  playPause.disabled = animationName === "";
  playPause.textContent = playing ? "Pause" : "Play";
}

function updateUrl() {
  const next = new URLSearchParams();
  next.set("model", modelName);
  if (animationName) {
    next.set("animation", animationName);
    next.set("phase", phase.toFixed(3));
  }
  next.set("view", cameraView);
  if (requestedEquipment) {
    next.set("equipment", requestedEquipment);
  }
  if (!playing) {
    next.set("play", "0");
  }
  history.replaceState(null, "", `${window.location.pathname}?${next}`);
}

function attachRequestedEquipment() {
  if (!requestedEquipment || modelName !== "human" || !isModelName(requestedEquipment)) {
    return;
  }
  const presentation = getEquipmentPresentation(requestedEquipment);
  if (!presentation) {
    return;
  }
  const socket = modelInstance.getSocket(presentation.equipped.socket);
  if (!socket) {
    throw new Error(`model does not define equipment socket: ${presentation.equipped.socket}`);
  }
  equipmentInstance = createModel(requestedEquipment);
  applyModelTransform(equipmentInstance.root, presentation.equipped);
  socket.add(equipmentInstance.root);
}

function requireElement<T extends HTMLElement>(id: string): T {
  const element = document.getElementById(id);
  if (!element) {
    throw new Error(`missing model lab element: ${id}`);
  }
  return element as T;
}

function parseView(value: string | null): CameraView {
  switch (value) {
    case "front":
    case "back":
    case "side":
    case "top":
    case "three-quarter":
      return value;
    default:
      return "three-quarter";
  }
}

function clampPhase(value: number): number {
  return Number.isFinite(value) ? THREE.MathUtils.clamp(value, 0, 1) : 0;
}
