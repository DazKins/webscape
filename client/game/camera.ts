import * as THREE from "three";
import Input from "../input.ts";

export default class Camera {
  private camera: THREE.PerspectiveCamera;
  private input: Input;

  private distance: number;
  private angle: number;
  private height: number;
  private orbitSpeed: number;
  private heightSpeed: number;
  private minHeight: number;
  private maxHeight: number;
  private cameraTarget: THREE.Vector3;

  constructor(input: Input) {
    this.camera = new THREE.PerspectiveCamera(
      75,
      window.innerWidth / window.innerHeight,
      0.1,
      1000
    );
    this.camera.position.z = 5;
    this.camera.position.y = 5;
    this.input = input;

    this.distance = 5;
    this.angle = 0;
    this.height = 5;
    this.orbitSpeed = 0.05;
    this.heightSpeed = 0.1;
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

  update(target: THREE.Vector3) {
    if (this.input.getKey("arrowleft")) {
      this.angle += this.orbitSpeed;
    }
    if (this.input.getKey("arrowright")) {
      this.angle -= this.orbitSpeed;
    }

    if (this.input.getKey("arrowup")) {
      this.height = Math.min(this.height + this.heightSpeed, this.maxHeight);
    }
    if (this.input.getKey("arrowdown")) {
      this.height = Math.max(this.height - this.heightSpeed, this.minHeight);
    }

    this.cameraTarget.lerp(target, 0.1);

    this.camera.position.x =
      this.cameraTarget.x + Math.cos(this.angle) * this.distance;
    this.camera.position.y = this.height;
    this.camera.position.z =
      this.cameraTarget.z + Math.sin(this.angle) * this.distance;

    this.camera.lookAt(this.cameraTarget);
  }

  onWindowResize() {
    this.camera.aspect = window.innerWidth / window.innerHeight;
    this.camera.updateProjectionMatrix();
  }
}
