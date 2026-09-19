import assert from "node:assert/strict";
import test from "node:test";
import RenderTiming from "../game/renderTiming.ts";

function context(supported = true) {
  const ext = { TIME_ELAPSED_EXT: 1, GPU_DISJOINT_EXT: 2 };
  const gl = {
    QUERY_RESULT_AVAILABLE: 3,
    QUERY_RESULT: 4,
    available: false,
    disjoint: false,
    lost: false,
    nanoseconds: 2500000,
    created: 0,
    deleted: 0,
    ended: 0,
    active: false,
    getExtension: () => supported ? ext : null,
    isContextLost() { return this.lost; },
    getParameter() { return this.disjoint; },
    createQuery() { this.created++; return {}; },
    beginQuery() { assert.equal(this.active, false); this.active = true; },
    endQuery() { assert.equal(this.active, true); this.active = false; this.ended++; },
    deleteQuery() { this.deleted++; },
    getQueryParameter(_query, parameter) {
      if (parameter === this.QUERY_RESULT_AVAILABLE) return this.available;
      assert.equal(this.available, true, "must not read a pending GPU result");
      return this.nanoseconds;
    },
  };
  return gl;
}

test("CPU averages measure draw work, excluding time between frames", (t) => {
  let now = 0;
  t.mock.method(performance, "now", () => now);
  const timing = new RenderTiming(context(false));
  timing.begin(); now = 2; timing.end();
  now = 16.67;
  timing.begin(); now += 4; timing.end();
  assert.deepEqual(timing.takeSample(), { cpuMs: 3, gpuMs: null, gpuSupported: false });
  assert.equal(timing.takeSample().cpuMs, null);
});

test("GPU queries remain bounded and results are collected asynchronously in milliseconds", () => {
  const gl = context();
  const timing = new RenderTiming(gl);
  for (let i = 0; i < 10; i++) { timing.begin(); timing.end(); }
  assert.equal(gl.created, 1);
  assert.equal(gl.ended, 1);
  assert.equal(timing.takeSample().gpuMs, null);
  gl.available = true;
  timing.begin(); timing.end();
  gl.nanoseconds = 4500000;
  timing.begin(); timing.end();
  assert.equal(timing.takeSample().gpuMs, 3.5);
  assert.equal(gl.deleted, 2);
  timing.reset();
  assert.equal(gl.deleted, gl.created);
});

test("GPU clock interruptions discard pending and accumulated timings", () => {
  const gl = context();
  const timing = new RenderTiming(gl);
  timing.begin(); timing.end();
  gl.available = true;
  timing.begin(); timing.end();
  gl.disjoint = true;
  timing.begin(); timing.end();
  assert.equal(timing.takeSample().gpuMs, null);
  assert.equal(gl.deleted, gl.created);
  gl.disjoint = false;
  timing.begin(); timing.end();
  assert.equal(gl.created, 3);
  timing.reset();
});

test("reset releases queries and clears measurements before sampling resumes", () => {
  const gl = context();
  const timing = new RenderTiming(gl);
  timing.begin(); timing.end();
  timing.reset();
  assert.equal(gl.deleted, 1);
  assert.deepEqual(timing.takeSample(), { cpuMs: null, gpuMs: null, gpuSupported: true });
  timing.begin(); timing.end();
  assert.equal(gl.created, 2);
  timing.reset();
});

test("context loss clears stale queries and resumes after restoration", () => {
  const gl = context();
  const timing = new RenderTiming(gl);
  timing.begin(); timing.end();
  gl.lost = true;
  timing.begin(); timing.end();
  assert.equal(gl.deleted, 1);
  assert.equal(timing.takeSample().cpuMs, null);
  gl.lost = false;
  timing.begin(); timing.end();
  assert.equal(gl.created, 2);
  timing.reset();
});
