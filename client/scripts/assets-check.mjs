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
    const { createModel, modelAssets } = await import("/game/models/registry.ts");
    const { clearSharedAssets } = await import("/game/models/assetCache.ts");
    const { constructionClient, prepareModelAssets } = await import("/game/assets/constructionClient.ts");
    const { createTerrainSurfaceGeometry } = await import("/game/world/terrainHeight.ts");
    const { terrainColor } = await import("/game/assets/construction.ts");
    const { construct } = await import("/game/assets/construction.ts");
    const { default: World } = await import("/game/world/world.ts");
    const errors = [];
    const check = (ok, message) => { if (!ok) errors.push(message); };
    await prepareModelAssets();
    check(constructionClient.worker instanceof Worker, "construction did not run in a worker");
    const a = createModel("tree"), b = createModel("tree");
    const trunkA = a.root.getObjectByName("treeTrunk"), trunkB = b.root.getObjectByName("treeTrunk");
    check(trunkA !== trunkB && trunkA.geometry === trunkB.geometry && trunkA.material === trunkB.material, "trees must share resources, not objects");
    a.seek("hit", 0.2);
    check(!a.root.getObjectByName("treeShake").quaternion.equals(b.root.getObjectByName("treeShake").quaternion), "tree animation leaked");
    const damaged = createModel("tree", { damageStage: 3 });
    check(damaged.root.getObjectByName("canopyDamage3").visible && !a.root.getObjectByName("canopyDamage3").visible, "tree damage leaked");
    let disposed = 0; trunkA.geometry.addEventListener("dispose", () => disposed++);
    a.dispose(); a.dispose(); modelAssets.clear(); clearSharedAssets();
    check(disposed === 0, "cache clear disposed live tree geometry");
    b.update(0.1); b.dispose();
    check(disposed === 0, "damaged tree lost its geometry");
    damaged.dispose(); check(disposed === 1, "last owner must dispose geometry once");

    const humanA = createModel("human", { color: "red" }), humanB = createModel("human", { color: "blue" });
    const torsoA = humanA.root.getObjectByName("tunic").children.find(o => o.isMesh);
    const torsoB = humanB.root.getObjectByName("tunic").children.find(o => o.isMesh);
    check(torsoA.geometry === torsoB.geometry && torsoA.material !== torsoB.material, "human geometry/appearance sharing incorrect");
    const clone = humanA.clone();
    check(clone.getSocket("rightHand") !== humanA.getSocket("rightHand"), "clone socket points at original");
    clone.seek("run", 0.25);
    check(!clone.getSocket("rightHand").parent.quaternion.equals(humanA.getSocket("rightHand").parent.quaternion), "cloned rig animation leaked");
    humanA.dispose(); humanB.dispose(); clone.dispose();

    const size = 4;
    const chunk = { terrainBorder: [], sizeX: size, sizeY: size, heights: Array.from({length: 36}, (_, i) => i % 4), terrain: Array(16).fill("grass"), walls: [{id:"a",type:"stone",x:0,y:0},{id:"b",type:"stone",x:1,y:0}] };
    chunk.terrain[0] = "water";
    const local = construct({kind:"chunk", chunk});
    const remote = await constructionClient.run({kind:"chunk", chunk});
    for (const name of ["position", "normal", "color"]) {
      check(JSON.stringify([...local.surfaces.terrain.attributes[name].array]) === JSON.stringify([...remote.surfaces.terrain.attributes[name].array]), `worker ${name} mismatch`);
    }
    for (const name of ["position", "normal", "uv", "waterWeight"]) {
      check(JSON.stringify([...local.surfaces.water.attributes[name].array]) === JSON.stringify([...remote.surfaces.water.attributes[name].array]), `worker water ${name} mismatch`);
    }
    for (const name of ["position", "normal", "color"]) {
      check(JSON.stringify([...local.surfaces.details.attributes[name].array]) === JSON.stringify([...remote.surfaces.details.attributes[name].array]), `worker detail ${name} mismatch`);
    }
    const serialize = value => JSON.stringify(value);
    const flat = { terrainBorder: [], sizeX: 4, sizeY: 4, heights: Array(36).fill(0), terrain: Array(16).fill("grassLong"), walls: [], originX: -32, originY: 64 };
    const detailBuild = input => construct({ kind: "chunk", chunk: input }).surfaces;
    const first = detailBuild(flat), repeated = detailBuild(flat);
    check(serialize(first) === serialize(repeated), "terrain changed on regeneration");
    const shifted = detailBuild({ ...flat, originX: 0 });
    check(serialize(first.details) !== serialize(shifted.details), "chunks repeat decoration patterns");
    check(serialize(first.terrain.attributes.color) !== serialize(shifted.terrain.attributes.color), "chunks repeat colour patterns");
    check(first.details.attributes.position.array.length / 9 >= 16 * 12, "long grass has bare tiles below its density floor");
    const short = detailBuild({ ...flat, terrain: Array(16).fill("grassShort") });
    const maxHeight = surface => Math.max(...surface.attributes.position.array.filter((_, i) => i % 3 === 1));
    check(maxHeight(first.details) > maxHeight(short.details) * 1.5, "grass length variants indistinguishable");
    for (const [type, limit] of [["grassLong", 36], ["grassShort", 12], ["stone", 15], ["water", 0], ["dirt", 0], ["unknown", 0]]) {
      const surface = detailBuild({ ...flat, terrain: Array(16).fill(type) }).details;
      check(surface.attributes.position.array.length / 9 <= 16 * limit, `${type} exceeded triangle budget`);
      check([...surface.attributes.position.array].every(Number.isFinite), `${type} has invalid vertices`);
    }
    const { variedTerrainColor, terrainColor: baseColor } = await import("/game/world/terrainAppearance.ts");
    for (let x = -20; x < 20; x++) {
      const color = variedTerrainColor("grass", x, 3), base = new THREE.Color(baseColor("grass"));
      check(Math.abs(color.r / base.r - 1) <= 0.0451, "tile colour outside subtle variation range");
    }
    check(JSON.stringify(local.surfaces.walls) === JSON.stringify(remote.surfaces.walls), "worker wall placement mismatch");
    // Duplicate vertices must agree across tile colors, water, and chunk seams.
    const normalsByPosition = new Map();
    for (let chunkY = 0; chunkY < 2; chunkY++) {
      for (let chunkX = 0; chunkX < 2; chunkX++) {
        const heights = [];
        for (let y = -1; y <= size; y++) {
          for (let x = -1; x <= size; x++) {
            heights.push((x + chunkX * size + 2 * (y + chunkY * size) + 30) % 7);
          }
        }
        const surfaces = construct({ kind: "chunk", chunk: {
          sizeX: size, sizeY: size, heights, terrainBorder: [],
          terrain: Array.from({ length: size * size }, (_, i) => i % 2 ? "water" : "grass"), walls: [],
        } }).surfaces;
        for (const surface of [surfaces.terrain, surfaces.water]) {
          const positions = surface.attributes.position.array;
          const normals = surface.attributes.normal.array;
          for (let i = 0; i < positions.length; i += 3) {
            const key = `${positions[i] + chunkX * size},${positions[i + 2] + chunkY * size}`;
            const normal = [...normals.slice(i, i + 3)];
            const previous = normalsByPosition.get(key);
            check(!previous || normal.every((value, axis) => Math.abs(value - previous[axis]) < 1e-6), `terrain normal seam at ${key}`);
            check(Math.abs(Math.hypot(...normal) - 1) < 1e-6 && normal[1] > 0, `invalid terrain normal at ${key}`);
            normalsByPosition.set(key, normal);
          }
        }
      }
    }
    const again = await constructionClient.run({kind:"chunk", chunk});
    check(again.surfaces.wallGeometries[0][1].attributes.position.array.length > 0, "worker transferred cached source buffers");

    const input = { isPointerBlocked: () => true };
    const world = new World(new THREE.Scene(), {x:4,y:4}, input);
    const data = {coordinate:{x:0,y:0},terrain:Array(16).fill("grass"),heights:Array(16).fill(1),walls:chunk.walls};
    world.applyChunkUpdate({load:[data]});
    world.applyChunkUpdate({unload:[data.coordinate]});
    world.applyChunkUpdate({load:[{...data,heights:Array(16).fill(3)}]});
    const visual = world.chunks.get("0,0");
    for(let i=0;i<500;i++) {
      world.update(null,0.016,{canHover:false,isCoarsePointer:true});
      if (visual.terrainMesh.geometry.attributes.position) break;
      await new Promise(resolve=>setTimeout(resolve,10));
    }
    const values = visual.terrainMesh.geometry.attributes.position?.array;
    check(Boolean(values), "async chunk never attached");
    if(values) {
      const center = Array.from(values).findIndex((value, i) => i % 3 === 0 && value === 0.5 && values[i + 2] === 0.5);
      check(center >= 0 && Math.abs(values[center + 1]-1.8)<1e-5, "stale chunk result overwrote replacement");
    }
    // A neighbor arriving during a rebuild must invalidate the previous border snapshot.
    world.applyChunkUpdate({load:[{...data,coordinate:{x:1,y:0},heights:Array(16).fill(6)}]});
    world.applyChunkUpdate({load:[{...data,coordinate:{x:1,y:1},heights:Array(16).fill(9)}]});
    for(let i=0;i<500;i++) {
      world.update(null,0.016,{canHover:false,isCoarsePointer:true});
      if (!world.building && !world.dirty.size && !world.completed.size) break;
      await new Promise(resolve=>setTimeout(resolve,10));
    }
    const expected = createTerrainSurfaceGeometry(visual.grid, visual.data.terrain, terrainColor);
    check(JSON.stringify([...expected.attributes.position.array]) === JSON.stringify([...visual.terrainMesh.geometry.attributes.position.array]), "neighbor border snapshot became stale");
    check(JSON.stringify([...expected.attributes.normal.array]) === JSON.stringify([...visual.terrainMesh.geometry.attributes.normal.array]), "neighbor normal snapshot became stale");
    expected.dispose();
    const liveWall = visual.root.getObjectByName("chunkWalls").children[0];
    let wallDisposed = 0;
    liveWall.geometry.addEventListener("dispose", () => wallDisposed++);
    clearSharedAssets();
    check(wallDisposed === 0, "cache clear disposed live wall");
    const liveDetails = visual.root.getObjectByName("chunkDetails");
    check(Boolean(liveDetails), "chunk decorations never attached");
    let detailsDisposed = 0;
    liveDetails?.geometry.addEventListener("dispose", () => detailsDisposed++);
    world.applyChunkUpdate({load:[data]}); world.dispose();
    check(detailsDisposed === 1, "chunk detail geometry was not disposed once");
    await new Promise(resolve=>setTimeout(resolve,50));
    check(world.chunks.size === 0, "disposed world resurrected chunks");
    // Worker failures must keep the client usable through the local fallback.
    constructionClient.worker.dispatchEvent(new Event("error"));
    const fallback = await constructionClient.run({kind:"chunk", chunk});
    check(fallback.surfaces.terrain.attributes.position.array.length > 0, "worker fallback failed");
    modelAssets.clear(); clearSharedAssets();
    return { errors };
  });
  assert.deepEqual(result.errors, []);
  console.log("Checked shared resource lifetime, independent poses/appearances, worker parity, transferable reuse, and stale chunk results.");
} finally {
  await browser?.close();
  server.kill("SIGTERM");
}
