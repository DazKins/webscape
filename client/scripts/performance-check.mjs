import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';
import ts from 'typescript';
import { spawn } from 'node:child_process';
import { chromium } from 'playwright';

const source = await readFile(new URL('../main.ts', import.meta.url), 'utf8');
const loop = ts.transpile(source.slice(source.indexOf('// Avoid rendering')));
for (const hz of [30, 60, 120, 144, 240]) {
  let callback, visibilityChanged;
  const deltas = [];
  const document = { hidden: false, addEventListener: (_, fn) => { visibilityChanged = fn; } };
  vm.runInNewContext(loop, {
    document, game: { update: delta => deltas.push(delta) },
    requestAnimationFrame: fn => { callback = fn; return 1; },
    cancelAnimationFrame: () => { callback = undefined; },
  });
  for (let i = 0; i < hz * 10; i++) callback(i * 1000 / hz);
  assert.ok(Math.abs(deltas.length - Math.min(hz, 60) * 10) <= 2);
  assert.ok(Math.abs(deltas.reduce((a, b) => a + b, 0) - 10) < 0.05);
  document.hidden = true;
  visibilityChanged();
  assert.equal(callback, undefined);
  document.hidden = false;
  visibilityChanged();
  callback(60000);
  assert.equal(deltas.at(-1), 0);
}

const server = spawn('node_modules/.bin/vite', ['--host', '127.0.0.1', '--port', '0'], {
  cwd: new URL('..', import.meta.url), stdio: ['ignore', 'pipe', 'pipe'],
});
let browser;
try {
  const url = await new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error('Vite startup timed out')), 15000);
    server.stdout.on('data', data => {
      const match = data.toString().match(/http:\/\/127\.0\.0\.1:\d+/);
      if (match) { clearTimeout(timer); resolve(match[0]); }
    });
    server.on('error', reject);
  });
  browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 800, height: 600 }, deviceScaleFactor: 2 });
  await page.route(`${url}/`, route => route.fulfill({ contentType: 'text/html', body: '<div id="scene" style="width:800px;height:600px"></div><div id="hud"></div>' }));
  await page.goto(url);
  const result = await page.evaluate(async () => {
    const { default: RefreshRuntime } = await import('/@react-refresh');
    RefreshRuntime.injectIntoGlobalHook(window);
    window.$RefreshReg$ = () => {};
    window.$RefreshSig$ = () => type => type;
    window.__vite_plugin_react_preamble_installed__ = true;
    const { default: Game } = await import('/game/game.ts');
    const { default: Camera } = await import('/game/camera.ts');
    const { default: World } = await import('/game/world/world.ts');
    const game = new Game(document.getElementById('scene'), document.getElementById('hud'));
    const input = { isPointerBlocked: () => false, getPointerPosition: () => ({ x: 400, y: 300 }), getKey: key => key === 'arrowleft' };
    const viewport = { width: 800, height: 600 };
    const positions = [30, 60, 144].map(fps => {
      const camera = new Camera(input, viewport);
      const target = camera.getPosition().clone().set(2, 0, 2);
      for (let i = 0; i < fps; i++) camera.update(target, {}, 1 / fps);
      return camera.getPosition().toArray();
    });
    const world = new World(game.scene, { x: 2, y: 2 }, input);
    world.applyChunkUpdate({ load: [{ coordinate: { x: 0, y: 0 }, terrain: Array(4).fill('grass'), heights: Array(4).fill(0), walls: [] }] });
    game.scene.updateMatrixWorld(true);
    const camera = game.camera.getInnerCamera();
    camera.position.set(0.5, 5, 0.5);
    camera.lookAt(0.5, 0, 0.5);
    let count = 0;
    const raycaster = world.pointerRaycaster;
    const intersect = raycaster.intersectObjects.bind(raycaster);
    raycaster.intersectObjects = (...args) => { count++; return intersect(...args); };
    const tile = world.getPointerTile(game.camera, viewport);
    for (let i = 0; i < 100; i++) world.getPointerTile(game.camera, viewport);
    const stationary = count;
    camera.position.x += 0.1;
    world.getPointerTile(game.camera, viewport);
    const moved = count;
    world.applyChunkUpdate({ unload: [{ x: 0, y: 0 }] });
    const unloaded = world.getPointerTile(game.camera, viewport);
    world.dispose();
    const pixelRatio = game.renderer.getPixelRatio();
    game.renderer.dispose();
    return { positions, tile, stationary, moved, count, unloaded, pixelRatio };
  });
  assert.equal(result.pixelRatio, 1.5);
  for (const position of result.positions) position.forEach((value, axis) => assert.ok(Math.abs(value - result.positions[0][axis]) < 1e-9));
  assert.deepEqual(result.tile, { x: 0, y: 0 });
  assert.equal(result.stationary, 1);
  assert.equal(result.moved, 2);
  assert.equal(result.count, 3);
  assert.equal(result.unloaded, undefined);
  console.log('Performance checks passed: frame cap, tab suspension, camera timing, pixel ratio, terrain picking cache.');
} finally {
  await browser?.close();
  server.kill();
}
