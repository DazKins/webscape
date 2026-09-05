import * as THREE from "three";

export default class RendererArrow {
  private readonly root = new THREE.Group();
  private readonly start: THREE.Vector3;

  constructor(parent: THREE.Object3D, start: THREE.Vector3) {
    this.start = start.clone();

    const shaft = new THREE.Mesh(
      new THREE.CylinderGeometry(0.014, 0.014, 0.65, 7),
      new THREE.MeshStandardMaterial({ color: 0x8a5a32, roughness: 0.9 }),
    );
    shaft.rotation.x = Math.PI / 2;
    this.root.add(shaft);

    const point = new THREE.Mesh(
      new THREE.ConeGeometry(0.045, 0.13, 5),
      new THREE.MeshStandardMaterial({
        color: 0xcbd1d4,
        roughness: 0.35,
        metalness: 0.45,
      }),
    );
    point.rotation.x = Math.PI / 2;
    point.position.z = 0.39;
    this.root.add(point);

    for (const side of [-1, 1]) {
      const fletching = new THREE.Mesh(
        new THREE.BoxGeometry(0.075, 0.01, 0.13),
        new THREE.MeshStandardMaterial({ color: 0xc44b3d, roughness: 0.85 }),
      );
      fletching.rotation.z = side * 0.55;
      fletching.position.z = -0.27;
      this.root.add(fletching);
    }

    this.root.position.copy(start);
    parent.add(this.root);
  }

  update(progress: number, target: THREE.Vector3) {
    const position = new THREE.Vector3().lerpVectors(this.start, target, progress);
    position.y += Math.sin(progress * Math.PI) * 0.12;
    this.root.position.copy(position);
    this.root.lookAt(target);
  }

  dispose() {
    this.root.traverse((object) => {
      if (!(object instanceof THREE.Mesh)) {
        return;
      }
      object.geometry.dispose();
      const materials = Array.isArray(object.material) ? object.material : [object.material];
      for (const material of materials) {
        material.dispose();
      }
    });
    this.root.removeFromParent();
  }
}
