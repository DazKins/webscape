import * as THREE from "three";

const palettes = [
  { phase: 0, sky: 0xd5a18b, ambient: 0xffddc6, ambientPower: 0.65, sun: 0xffbc86, sunPower: 0.5, night: 0.35 },
  { phase: 0.04, sky: 0x87ceeb, ambient: 0xffffff, ambientPower: 1, sun: 0xfff4df, sunPower: 0.8, night: 0 },
  { phase: 0.44, sky: 0x87ceeb, ambient: 0xffffff, ambientPower: 1, sun: 0xfff4df, sunPower: 0.8, night: 0 },
  { phase: 0.48, sky: 0xd18b79, ambient: 0xffc6a5, ambientPower: 0.55, sun: 0xff9966, sunPower: 0.55, night: 0.25 },
  { phase: 0.5, sky: 0x555878, ambient: 0xb9b8e2, ambientPower: 0.35, sun: 0xa0b8ed, sunPower: 0.25, night: 0.75 },
  { phase: 0.54, sky: 0x111c35, ambient: 0x94acd9, ambientPower: 0.28, sun: 0xa4bff0, sunPower: 0.3, night: 1 },
  { phase: 0.96, sky: 0x111c35, ambient: 0x94acd9, ambientPower: 0.28, sun: 0xa4bff0, sunPower: 0.3, night: 1 },
  { phase: 1, sky: 0xd5a18b, ambient: 0xffddc6, ambientPower: 0.65, sun: 0xffbc86, sunPower: 0.5, night: 0.35 },
].map(palette => ({ ...palette,
  sky: new THREE.Color(palette.sky), ambient: new THREE.Color(palette.ambient), sun: new THREE.Color(palette.sun),
}));

type LightSource = {
  anchor: THREE.Object3D;
  position: THREE.Vector3;
  material: THREE.MeshStandardMaterial;
  glow: THREE.Sprite;
  distance: number;
};

// A fixed pool avoids compiling new material shaders as lanterns load/unload.
// All authored sources glow; the nearest eight cast local, unshadowed light.
export default class EnvironmentLighting {
  private readonly ambient = new THREE.AmbientLight();
  private readonly sun = new THREE.DirectionalLight();
  private readonly sky = new THREE.Color();
  private readonly lights = Array.from({ length: 8 }, () => new THREE.PointLight(0xffbd70, 0, 7, 2));
  private readonly sources = new Set<LightSource>();
  private readonly glowTexture = createGlowTexture();

  constructor(scene: THREE.Scene) {
    scene.background = this.sky;
    this.sun.position.set(10, 10, 10);
    scene.add(this.ambient, this.sun, ...this.lights);
    this.update(0, new THREE.Vector3());
  }

  // Models opt in using a lightSource socket on their luminous mesh. Materials
  // in the model cache remain immutable; each live light owns its cloned glass.
  registerSource(anchor: THREE.Object3D): () => void {
    if (!(anchor instanceof THREE.Mesh) || !(anchor.material instanceof THREE.MeshStandardMaterial)) {
      return () => {};
    }
    const originalMaterial = anchor.material;
    const material = originalMaterial.clone();
    anchor.material = material;
    const glow = new THREE.Sprite(new THREE.SpriteMaterial({
      map: this.glowTexture, color: 0xffc17e, transparent: true,
      blending: THREE.AdditiveBlending, depthWrite: false, opacity: 0,
    }));
    glow.scale.setScalar(1.15);
    glow.raycast = () => {};
    anchor.add(glow);
    const source = { anchor, position: new THREE.Vector3(), material, glow, distance: Infinity };
    this.sources.add(source);
    return () => {
      this.sources.delete(source);
      anchor.material = originalMaterial;
      glow.removeFromParent();
      glow.material.dispose();
      material.dispose();
    };
  }

  update(cycleProgress: number, focus: THREE.Vector3) {
    const upper = palettes.findIndex(palette => palette.phase > cycleProgress);
    const next = palettes[upper < 0 ? palettes.length - 1 : Math.max(1, upper)];
    const previous = palettes[upper < 0 ? palettes.length - 2 : Math.max(0, upper - 1)];
    const blend = THREE.MathUtils.smoothstep(cycleProgress, previous.phase, next.phase);
    this.sky.lerpColors(previous.sky, next.sky, blend);
    this.ambient.color.lerpColors(previous.ambient, next.ambient, blend);
    this.sun.color.lerpColors(previous.sun, next.sun, blend);
    this.ambient.intensity = THREE.MathUtils.lerp(previous.ambientPower, next.ambientPower, blend);
    this.sun.intensity = THREE.MathUtils.lerp(previous.sunPower, next.sunPower, blend);
    const night = THREE.MathUtils.lerp(previous.night, next.night, blend);

    for (const source of this.sources) {
      source.anchor.getWorldPosition(source.position);
      source.distance = Math.hypot(source.position.x - focus.x, source.position.z - focus.z);
      source.material.emissiveIntensity = 0.15 + night * 2.8;
      source.glow.material.opacity = night * 0.65;
    }
    const nearest = [...this.sources].filter(source => source.distance < 24)
      .sort((a, b) => a.distance - b.distance).slice(0, this.lights.length);
    this.lights.forEach((light, index) => {
      const source = nearest[index];
      light.intensity = source ? night * 13 * (1 - THREE.MathUtils.smoothstep(source.distance, 18, 24)) : 0;
      if (source) light.position.copy(source.position);
    });
  }
}

function createGlowTexture() {
  const canvas = document.createElement("canvas");
  canvas.width = canvas.height = 64;
  const context = canvas.getContext("2d")!;
  const gradient = context.createRadialGradient(32, 32, 0, 32, 32, 32);
  gradient.addColorStop(0, "rgba(255,255,255,0.9)");
  gradient.addColorStop(0.18, "rgba(255,255,255,0.45)");
  gradient.addColorStop(0.5, "rgba(255,255,255,0.1)");
  gradient.addColorStop(1, "rgba(255,255,255,0)");
  context.fillStyle = gradient;
  context.fillRect(0, 0, 64, 64);
  return new THREE.CanvasTexture(canvas);
}
