import assert from "node:assert/strict";
import test from "node:test";
import Input from "../input.ts";

function setup() {
  const timers = new Map();
  const listeners = new Map();
  let nextTimer = 0;
  globalThis.window = {
    addEventListener: (name, handler) => listeners.set(name, handler),
    setTimeout: (handler) => { timers.set(++nextTimer, handler); return nextTimer; },
    clearTimeout: (id) => timers.delete(id),
  };
  const input = new Input();
  input.setWorldBlocked(false);
  const calls = [];
  input.registerPointerCallbacks({
    onTap: () => calls.push("tap"),
    onLongPress: () => calls.push("longPress"),
    onDrag: () => calls.push("drag"),
  });
  input.registerRightClickCallback(() => calls.push("contextMenu"));
  return { input, calls, listeners, hold: () => { for (const callback of timers.values()) callback(); } };
}

const pointer = (pointerType, button = 0, clientX = 100) => ({
  pointerId: 1, pointerType, button, buttons: 1, clientX, clientY: 100,
});

test("mouse holds never trigger long-press; only primary releases tap", () => {
  for (const button of [0, 1, 2]) {
    const { input, calls, hold } = setup();
    input.onPointerDown(pointer("mouse", button));
    hold();
    assert.deepEqual(calls, []);
    input.onPointerUp(pointer("mouse", button));
    assert.deepEqual(calls, button === 0 ? ["tap"] : []);
  }
});

test("right-click uses the context menu callback without a tap or long-press", () => {
  const { input, calls, listeners, hold } = setup();
  input.onPointerDown(pointer("mouse", 2));
  let prevented = false;
  listeners.get("contextmenu")({ preventDefault: () => { prevented = true; } });
  hold();
  input.onPointerUp(pointer("mouse", 2));
  assert.equal(prevented, true);
  assert.deepEqual(calls, ["contextMenu"]);
});

test("touch long-press remains available and suppresses the release tap", () => {
  const { input, calls, hold } = setup();
  input.onPointerDown(pointer("touch"));
  hold();
  input.onPointerUp(pointer("touch"));
  assert.deepEqual(calls, ["longPress"]);
});

test("dragging or cancellation cancels the touch menu gesture", () => {
  for (const cancel of [false, true]) {
    const { input, calls, hold } = setup();
    input.onPointerDown(pointer("touch"));
    if (cancel) input.onPointerCancel(pointer("touch"));
    else input.onPointerMove(pointer("touch", 0, 150));
    hold();
    input.onPointerUp(pointer("touch", 0, 150));
    assert.deepEqual(calls, cancel ? [] : ["drag"]);
  }
});

test("touch target capture runs immediately, before the long-press timeout", () => {
  const { input, hold } = setup();
  let underFinger = "moving-rat";
  let target = null;
  const opened = [];
  input.registerPointerCallbacks({
    onTouchStart: () => { target = underFinger; },
    onLongPress: () => opened.push(target),
  });
  input.onPointerDown(pointer("touch"));
  assert.equal(target, "moving-rat");
  assert.deepEqual(opened, []);
  underFinger = "another-rat";
  hold();
  assert.deepEqual(opened, ["moving-rat"]);
});

test("touch native context menus do not trigger a second selection", () => {
  const { calls, listeners } = setup();
  listeners.get("contextmenu")({ pointerType: "touch", preventDefault() {} });
  assert.deepEqual(calls, []);
});

test("blocked touches do not capture targets or open a menu", () => {
  const { input, hold } = setup();
  const calls = [];
  input.registerPointerCallbacks({
    onTouchStart: () => calls.push("capture"),
    onLongPress: () => calls.push("open"),
  });
  input.setPointerBlocked(true);
  input.onPointerDown(pointer("touch"));
  input.setPointerBlocked(false);
  hold();
  assert.deepEqual(calls, []);
});
