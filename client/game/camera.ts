import * as THREE from "three";
import Input from "../input.ts";
import type { ViewportSize } from "../responsive.ts";

export default class Camera {
  private camera: THREE.PerspectiveCamera;
  private input: Input;

  private distance: number;
  private renderedDistance: number;
  private angle: number;
  private height: number;
  private renderedHeight: number;
  private orbitSpeed: number;
  private heightSpeed: number;
  private dragOrbitSpeed: number;
  private dragHeightSpeed: number;
  private minHeight: number;
  private maxHeight: number;
  private cameraTarget: THREE.Vector3;

  constructor(input: Input, viewport: ViewportSize) {
    this.camera = new THREE.PerspectiveCamera(
      75,
      viewport.width / viewport.height,
      0.1,
      1000
    );
    this.camera.position.z = 5;
    this.camera.position.y = 5;
    this.input = input;

    this.distance = 5;
    this.renderedDistance = this.distance;
    this.angle = 0;
    this.height = 5;
    this.renderedHeight = this.height;
    this.orbitSpeed = 0.05;
    this.heightSpeed = 0.1;
    this.dragOrbitSpeed = 0.008;
    this.dragHeightSpeed = 0.012;
    this.minHeight = 2;
    this.maxHeight = 10;
    this.cameraTarget = new THREE.Vector3(0, 0, 0);
  }

  getPosition(): THREE.Vector3 {
    return this.camera.position;
  }

  getInnerCamera(): THREE.PerspectiveCamera {
    return this.camera;
  }

  orbitByDrag(delta: { x: number; y: number }) {
    this.angle -= delta.x * this.dragOrbitSpeed;
    this.height = Math.max(
      this.minHeight,
      Math.min(this.maxHeight, this.height + delta.y * this.dragHeightSpeed)
    );
  }

  update(target: THREE.Vector3, options: { distance?: number; height?: number } = {}) {
    let heightChangedByInput = false;

    if (this.input.getKey("arrowleft")) {
      this.angle += this.orbitSpeed;
    }
    if (this.input.getKey("arrowright")) {
      this.angle -= this.orbitSpeed;
    }

    if (this.input.getKey("arrowup")) {
      this.height = Math.min(this.height + this.heightSpeed, this.maxHeight);
      heightChangedByInput = true;
    }
    if (this.input.getKey("arrowdown")) {
      this.height = Math.max(this.height - this.heightSpeed, this.minHeight);
      heightChangedByInput = true;
    }

    const desiredDistance = options.distance ?? this.distance;
    const desiredHeight = options.height ?? this.height;

    this.cameraTarget.lerp(target, 0.1);
    this.renderedDistance += (desiredDistance - this.renderedDistance) * 0.1;
    if (heightChangedByInput && options.height === undefined) {
      this.renderedHeight = desiredHeight;
    } else {
      this.renderedHeight += (desiredHeight - this.renderedHeight) * 0.1;
    }

    this.camera.position.x =
      this.cameraTarget.x + Math.cos(this.angle) * this.renderedDistance;
    this.camera.position.y = this.cameraTarget.y + this.renderedHeight;
    this.camera.position.z =
      this.cameraTarget.z + Math.sin(this.angle) * this.renderedDistance;

    this.camera.lookAt(this.cameraTarget);
  }

  onResize(viewport: ViewportSize) {
    this.camera.aspect = viewport.width / viewport.height;
    this.camera.updateProjectionMatrix();
  }
}
