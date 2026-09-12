import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import net from "node:net";
import { chromium } from "playwright";

const port = await new Promise((resolve, reject) => {
  const server = net.createServer();
  server.on("error", reject);
  server.listen(0, "127.0.0.1", () => {
    const port = server.address().port;
    server.close(() => resolve(port));
  });
});
const origin = `http://127.0.0.1:${port}`;
const server = spawn("node_modules/.bin/vite", ["--host", "127.0.0.1", "--port", String(port), "--strictPort"], { stdio: "ignore" });
let browser;
try {
  for (let attempt = 0; ; attempt++) {
    try {
      if ((await fetch(origin)).ok) break;
    } catch { /* Vite is starting. */ }
    assert(attempt < 100, "Vite failed to start");
    await new Promise(resolve => setTimeout(resolve, 100));
  }
  browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  await page.goto(`${origin}/model-lab.html?capture=1`);
  const result = await page.evaluate(async () => {
    const THREE = await import("/node_modules/three/build/three.module.js");
    const { createModel, modelNames } = await import("/game/models/registry.ts");
    const { applyModelTransform, getEquipmentPresentation } = await import("/game/models/equipment.ts");
    const { default: RendererArrow } = await import("/game/renderer/rendererArrow.ts");
    const { default: RendererMagicBolt } = await import("/game/renderer/rendererMagicBolt.ts");
    const { default: RendererChest } = await import("/game/renderer/rendererChest.ts");
    const { default: Entity } = await import("/game/entity/entity.ts");
    const errors = [];
    let animations = 0;
    let resources = 0;
    function check(condition, message) { if (!condition) errors.push(message); }
    function transforms(root) {
      root.updateMatrixWorld(true);
      const values = [];
      root.traverse(object => values.push(...object.matrixWorld.elements));
      return values;
    }
    function same(a, b) { return a.length === b.length && a.every((value, index) => Math.abs(value - b[index]) < 1e-6); }
    for (const name of modelNames) {
      const model = createModel(name);
      const owned = new Set();
      model.root.traverse(object => {
        if (!object.isMesh) return;
        owned.add(object.geometry);
        for (const material of Array.isArray(object.material) ? object.material : [object.material]) owned.add(material);
        const positions = object.geometry.getAttribute("position");
        check(Array.from(positions.array).every(Number.isFinite), `${name}: invalid geometry`);
      });
      for (const clip of model.animations) {
        animations++;
        model.seek(clip.name, 0);
        const start = transforms(model.root);
        for (let step = 0; step <= 32; step++) {
          const phase = step / 32;
          for (const joint of Object.keys(clip.sample(phase))) {
            check(Boolean(model.root.getObjectByName(joint)), `${name}/${clip.name}: missing joint ${joint}`);
          }
          model.seek(clip.name, phase);
          check(transforms(model.root).every(Number.isFinite), `${name}/${clip.name}: invalid pose at ${phase}`);
          if (name === "human" && ["run", "bowRun", "staffRun"].includes(clip.name)) {
            for (const side of ["left", "right"]) {
              const ankle = model.root.getObjectByName(`${side}Ankle`);
              const foot = ankle.children[0];
              const bounds = new THREE.Box3().setFromObject(foot);
              check(bounds.min.y >= -0.015, `${clip.name}/${side}: foot below ground at ${phase}: ${bounds.min.y}`);
            }
          }
        }
        if (clip.loop) check(same(start, transforms(model.root)), `${name}/${clip.name}: loop seam`);
        model.seek(clip.name, 0.37);
        const sought = transforms(model.root);
        model.seek(clip.name, 0);
        model.update(clip.duration * 0.37);
        check(same(sought, transforms(model.root)), `${name}/${clip.name}: seek and playback disagree`);
      }
      for (const resource of owned) {
        let disposed = 0;
        resource.addEventListener("dispose", () => disposed++);
        resource.userData ??= {};
        resource.userData.disposeCount = () => disposed;
      }
      model.dispose();
      model.dispose();
      for (const resource of owned) check(resource.userData.disposeCount() === 1, `${name}: resource not disposed exactly once`);
      resources += owned.size;
    }
    const human = createModel("human");
    for (const name of modelNames) {
      const presentation = getEquipmentPresentation(name);
      if (!presentation) continue;
      const equipment = createModel(name);
      const socket = human.getSocket(presentation.equipped.socket);
      check(Boolean(socket), `${name}: missing socket`);
      applyModelTransform(equipment.root, presentation.equipped);
      socket.add(equipment.root);
      for (const clip of human.animations) {
        human.seek(clip.name, 0.5);
        check(transforms(equipment.root).every(Number.isFinite), `${name}/${clip.name}: invalid attachment`);
      }
      equipment.dispose();
    }
    for (const clip of ["attack", "pickup", "chop"]) {
      human.seek(clip, 0);
      const start = transforms(human.root);
      human.seek(clip, 1);
      check(same(start, transforms(human.root)), `${clip}: does not recover to starting pose`);
    }
    human.seek("fishWait", 0);
    const waiting = transforms(human.root);
    for (const phase of [0, 1]) {
      human.seek("fishAction", phase);
      check(same(waiting, transforms(human.root)), `fishAction: does not join waiting pose at ${phase}`);
    }
    human.seek("pickup", 0.5);
    check(human.getSocket("rightHand").getWorldPosition(new THREE.Vector3()).y < 0.16, "pickup: hand does not reach the ground");
    human.dispose();
    // A newly interested client must immediately see the replicated open lid.
    const chestEntity = new Entity("chest-check");
    chestEntity.updateComponent("position", { x: 0, y: 0 });
    chestEntity.updateComponent("openable", { isOpen: true });
    const chestScene = new THREE.Scene();
    const chestRenderer = new RendererChest(chestScene, chestEntity);
    const hinge = chestRenderer.getObject3D().getObjectByName("lidHinge");
    check(hinge.rotation.x < -1.5, "chest: initial open state not rendered");
    chestEntity.updateComponent("openable", { isOpen: false });
    chestRenderer.update(0.2);
    check(hinge.rotation.x > -1.9 && hinge.rotation.x < 0, "chest: closing does not interpolate");
    chestRenderer.update(1);
    check(Math.abs(hinge.rotation.x) < 1e-6, "chest: does not close");
    chestRenderer.onRemove();
    check(chestScene.children.length === 0, "chest: not removed");
    for (const Renderer of [RendererArrow, RendererMagicBolt]) {
      const parent = new THREE.Group();
      const projectile = new Renderer(parent, new THREE.Vector3(0, 1, 0));
      const target = new THREE.Vector3(2, 0.5, 3);
      projectile.update(0.5, target);
      const first = transforms(parent);
      projectile.update(0.5, target);
      check(same(first, transforms(parent)), `${Renderer.name}: depends on frame count`);
      projectile.update(1, target);
      check(transforms(parent).every(Number.isFinite), `${Renderer.name}: invalid impact pose`);
      projectile.dispose();
      check(parent.children.length === 0, `${Renderer.name}: not removed`);
    }
    return { models: modelNames.length, animations, resources, errors };
  });
  assert.deepEqual(result.errors, []);
  console.log(`Checked ${result.models} models, ${result.animations} animations, ${result.resources} owned resources, equipment sockets and both projectile effects.`);
} finally {
  await browser?.close();
  server.kill("SIGTERM");
}
