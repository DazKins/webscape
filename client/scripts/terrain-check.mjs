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
  const page = await browser.newPage({ viewport: { width: 1200, height: 900 } });
  const browserErrors = [];
  page.on("pageerror", error => browserErrors.push(error.message));
  page.on("console", message => { if (message.type() === "error") browserErrors.push(message.text()); });
  await page.goto(`${origin}/model-lab.html?capture=1`);
  const result = await page.evaluate(async () => {
    const THREE = await import("/node_modules/three/build/three.module.js");
    const { createTerrainBlend } = await import("/game/world/terrainBlend.ts");
    const { variedTerrainColor } = await import("/game/world/terrainAppearance.ts");
    const { construct } = await import("/game/assets/construction.ts");
    const { default: World } = await import("/game/world/world.ts");
    const errors = [];
    const check = (ok, message) => { if (!ok) errors.push(message); };
    const size = 4;
    const terrainAt = (x, y) => x === -1 ? "road" : x >= 0 && y >= 0 ? "water" : y < -2 && x < -1 ? "stone" : y < 0 && x >= 0 ? "dirt" : "grassLong";
    const heightAt = (x, y) => x < -2 && y < -2 ? 1 : 0;
    const snapshot = (originX, originY) => {
      const terrain = [], terrainBorder = [], heights = [];
      for (let y = -1; y <= size; y++) for (let x = -1; x <= size; x++) {
        terrainBorder.push(terrainAt(originX + x, originY + y));
        heights.push(heightAt(originX + x, originY + y));
        if (x >= 0 && y >= 0 && x < size && y < size) terrain.push(terrainAt(originX + x, originY + y));
      }
      return { sizeX: size, sizeY: size, originX, originY, terrain, terrainBorder, heights, walls: [] };
    };
    const seen = new Map(), waterSeen = new Map();
    let shared = 0;
    for (const originY of [-4, 0]) for (const originX of [-4, 0]) {
      const chunk = snapshot(originX, originY);
      const blend = createTerrainBlend(size, size, chunk.terrain, chunk.terrainBorder, originX, originY);
      for (let y = 0; y < size; y++) for (let x = 0; x < size; x++) {
        const color = blend(x + 0.5, y + 0.5).color;
        const expected = variedTerrainColor(terrainAt(originX + x, originY + y), originX + x, originY + y);
        check(color.toArray().every((v, i) => Math.abs(v - expected.toArray()[i]) < 1e-9), "tile centre lost its terrain colour");
      }
      const { surfaces } = construct({ kind: "chunk", chunk });
      for (const [name, values] of Object.entries(surfaces.terrain.attributes)) {
        check([...values.array].every(Number.isFinite), `invalid terrain ${name}`);
      }
      const positions = surfaces.terrain.attributes.position.array, colors = surfaces.terrain.attributes.color.array;
      for (let i = 0; i < positions.length; i += 3) {
        const key = `${positions[i] + originX},${positions[i + 2] + originY}`;
        const color = [...colors.slice(i, i + 3)], previous = seen.get(key);
        if (previous) { shared++; check(color.every((v, j) => Math.abs(v - previous[j]) < 1e-6), `colour seam at ${key}`); }
        seen.set(key, color);
      }
      const water = surfaces.water.attributes;
      for (let i = 0; i < water.waterWeight.array.length; i++) {
        const x = water.position.array[i * 3] + originX, y = water.position.array[i * 3 + 2] + originY;
        const key = `${x},${y}`, weight = water.waterWeight.array[i];
        check(weight >= 0 && weight <= 1, "invalid shoreline opacity");
        check(water.uv.array[i * 2] === x && water.uv.array[i * 2 + 1] === y, "water pattern resets at chunk edge");
        check(!waterSeen.has(key) || Math.abs(waterSeen.get(key) - weight) < 1e-6, `water seam at ${key}`);
        waterSeen.set(key, weight);
      }
    }
    check(shared > 100, "fixture did not exercise shared edges");
    check([...waterSeen.values()].some(v => v > 0 && v < 1), "shoreline does not fade");
    const roadChunk = snapshot(-4, -4);
    const roadBlend = createTerrainBlend(size, size, roadChunk.terrain, roadChunk.terrainBorder, -4, -4);
    check(roadBlend(3.5, 2.5).grass === 0, "grass intrudes into road centre");
    check(roadBlend(2.95, 2.5).grass < roadBlend(2.5, 2.5).grass, "grass does not taper toward road");
    const isolated = createTerrainBlend(1, 1, ["dirt"], [], -10, -10);
    for (const [x, y] of [[0, 0], [1, 1], [0, 1], [1, 0]]) {
      check(isolated(x, y).color.toArray().every((v, i) => Math.abs(v - variedTerrainColor("dirt", -10, -10).toArray()[i]) < 1e-9), "missing neighbour darkens terrain");
    }

    // Exercise the actual worker, materials, and neighbour invalidation, then capture them.
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x263340);
    scene.add(new THREE.HemisphereLight(0xffffff, 0x667766, 2.2));
    const sun = new THREE.DirectionalLight(0xfff0d5, 2.5);
    sun.position.set(-3, 10, 5); scene.add(sun);
    const world = new World(scene, { x: size, y: size }, { isPointerBlocked: () => true });
    const loads = [];
    for (const y of [-1, 0]) for (const x of [-1, 0]) {
      const chunk = snapshot(x * size, y * size);
      loads.push({ coordinate: { x, y }, terrain: chunk.terrain, heights: chunk.terrain.map((_, i) => heightAt(x * size + i % size, y * size + Math.floor(i / size))), walls: [] });
    }
    const settle = async () => {
      for (let attempt = 0; attempt < 500; attempt++) {
        world.update(null, 0.016, { canHover: false, isCoarsePointer: true });
        if (!world.building && !world.dirty.size && !world.completed.size) return;
        await new Promise(resolve => setTimeout(resolve, 10));
      }
      throw new Error("chunk builds did not settle");
    };
    world.applyChunkUpdate({ load: loads }); await settle();
    const visual = world.chunks.get("-1,0");
    const before = [...visual.terrainMesh.geometry.attributes.color.array];
    world.applyChunkUpdate({ unload: [{ x: 0, y: 0 }] }); await settle();
    check(JSON.stringify(before) !== JSON.stringify([...visual.terrainMesh.geometry.attributes.color.array]), "unloading neighbour did not update colours");
    world.applyChunkUpdate({ load: [loads[3]] }); await settle();
    check(JSON.stringify(before) === JSON.stringify([...visual.terrainMesh.geometry.attributes.color.array]), "reloading neighbour changed colours");
    const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    renderer.setSize(1200, 900);
    renderer.setPixelRatio(1);
    document.body.replaceChildren(renderer.domElement);
    document.body.style.margin = "0";
    const camera = new THREE.PerspectiveCamera(40, 1200 / 900, 0.1, 100);
    camera.position.set(7, 11, 9); camera.lookAt(0, 0, 0);
    renderer.render(scene, camera);
    return { errors, shared, drawCalls: renderer.info.render.calls, triangles: renderer.info.render.triangles };
  });
  assert.deepEqual(result.errors, []);
  assert.deepEqual(browserErrors, []);
  await page.screenshot({ path: process.env.TERRAIN_SCREENSHOT ?? "/tmp/webscape-terrain.png" });
  console.log(`Checked terrain centres, ${result.shared} shared vertices, shoreline fade, grass taper, missing neighbours, and unload/reload. Preview: ${result.drawCalls} draw calls, ${result.triangles} triangles.`);
} finally {
  await browser?.close();
  server.kill("SIGTERM");
}
