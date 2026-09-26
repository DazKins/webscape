import assert from "node:assert/strict";
import test from "node:test";
import * as THREE from "three";
import { pickEntityAtScreenPoint } from "../game/entityPicking.ts";

const viewport = { width: 200, height: 200 };
const camera = new THREE.OrthographicCamera(-100, 100, 100, -100, 0.1, 100);
camera.position.z = 10;
camera.updateMatrixWorld();

function entity(id, x = 0, z = 0, size = 20) {
  const mesh = new THREE.Mesh(new THREE.PlaneGeometry(size, size), new THREE.MeshBasicMaterial());
  mesh.position.set(x, 0, z);
  mesh.userData.entityId = id;
  mesh.updateMatrixWorld();
  return mesh;
}
const pick = (x, y, objects) => pickEntityAtScreenPoint(x, y, camera, viewport, objects);

test("padding selects just outside the model but leaves distant ground clear", () => {
  const target = entity("rat");
  assert.equal(pick(118, 100, [target]), "rat");
  assert.equal(pick(122.1, 100, [target]), null);
  assert.equal(pick(118, 118, [target]), "rat");
});

test("a direct hit wins over a nearby padded target", () => {
  assert.equal(pick(100, 100, [entity("nearby", 12, 2), entity("direct")]), "direct");
});

test("overlapping direct hits select the front entity", () => {
  assert.equal(pick(100, 100, [entity("back"), entity("front", 0, 2)]), "front");
});

test("nested meshes resolve the owning entity", () => {
  const group = new THREE.Group();
  group.userData.entityId = "rat";
  const mesh = entity("unused");
  delete mesh.userData.entityId;
  group.add(mesh);
  group.updateMatrixWorld();
  assert.equal(pick(100, 100, [group]), "rat");
});
