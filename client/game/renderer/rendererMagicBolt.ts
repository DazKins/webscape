import * as THREE from "three";

export default class RendererMagicBolt {
  private readonly root = new THREE.Group();
  private readonly start: THREE.Vector3;

  constructor(parent: THREE.Object3D, start: THREE.Vector3) {
    this.start = start.clone();
    const core = new THREE.Mesh(
      new THREE.OctahedronGeometry(0.095, 0),
      new THREE.MeshStandardMaterial({
        color: 0xf2fdff,
        emissive: 0xbcefff,
        emissiveIntensity: 2.2,
        roughness: 0.15,
      }),
    );
    this.root.add(core);

    const aura = new THREE.Mesh(
      new THREE.OctahedronGeometry(0.17, 0),
      new THREE.MeshBasicMaterial({
        color: 0x69cfff,
        transparent: true,
        opacity: 0.32,
        depthWrite: false,
      }),
    );
    aura.rotation.y = Math.PI / 4;
    this.root.add(aura);

    const light = new THREE.PointLight(0x9ee8ff, 1.8, 2.5);
    this.root.add(light);
    this.root.position.copy(start);
    parent.add(this.root);
  }

  update(progress: number, target: THREE.Vector3) {
    const eased = THREE.MathUtils.smoothstep(progress, 0, 1);
    this.root.position.lerpVectors(this.start, target, eased);
    this.root.position.y += Math.sin(progress * Math.PI) * 0.2;
    this.root.rotation.x += 0.12;
    this.root.rotation.y += 0.18;
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
