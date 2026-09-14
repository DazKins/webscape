# Shared model assets

`modelAssets.instantiate(name, options)` (also exported as `createModel`) returns
an independent model instance. Its meshes and joints have their own transforms,
visibility, sockets and animation playback. Geometry and materials are immutable
shared assets.

Trees, scenery, equipment and other repeated models use a bounded cache of 64
model templates. Instantiation clones the hierarchy and reconnects joints and
sockets to the clone. Tree damage stage belongs to the instance, not the cache
key. Humans build their own hierarchy from shared primitives and palette
materials, avoiding a cache of every appearance combination.

`primitives.ts` caches geometry by shape parameters and materials by appearance.
Special geometry (tree skirts, embedded rocks, apples, chest lids) uses explicit
variant keys and is modified only inside its `sharedGeometry` builder. Never
modify cached geometry or materials after publishing them. For a new shape,
clone the primitive inside a separately keyed builder. Transforming a mesh does
not modify its geometry.

`assetCache.ts` gives each cache entry and each live model a resource reference.
Disposing an instance releases its references. LRU eviction releases the cache's
reference, and the last release disposes the resource. Geometry and material
caches retain at most 1024 and 256 entries respectively; live instances can keep
evicted resources alive. `modelAssets.clear()` and `clearSharedAssets()` release
cached ownership without invalidating live instances. Equipment owns its own
model and must still be disposed separately from its host.

# Background construction

`../assets/construction.worker.ts` prepares common procedural model geometry at
startup, including all human hair styles. Buffers are transferred to the main
thread and installed in the geometry cache. Authored building dimensions are
prepared on demand before their renderer is constructed. Main-thread factories
then assemble meshes/materials/joints using those cached geometries. Direct
factory use (for example the model lab), cache eviction, or worker failure can
still fall back to synchronous geometry construction.

Chunk jobs receive authored terrain and a one-tile height border, and generate
terrain, water and wall geometry in the worker. Wall geometry is shared by shape
and slope; textures and materials are cached on the main thread. Cached worker
buffers are copied before transfer so later builds can still reuse them.

Each world coalesces dirty chunk requests and runs one at a time. Revision and
object-identity checks reject outdated results after replacement, neighbor
changes, unloading or world disposal. Existing surfaces remain until replacements
are ready. At most one completed chunk is attached per frame. Entity construction
and its first update have a 3 ms frame budget, with the local player first; one
individual model can exceed that budget. Rendering and GPU uploads remain on
the main thread. This reduces CPU stalls but does not guarantee a frame-time cap.

# Validation

From `client/`:

- `pnpm run build`
- `pnpm run model:check`: registry geometry, animations, disposal and equipment.
- `pnpm run assets:check`: sharing, appearance/pose isolation, live-resource
  ownership, actual worker execution, terrain/water/wall parity, transferable
  reuse, stale results, neighboring borders and worker failure fallback.
- `node scripts/model-screenshot.mjs --model human --animation run --phase 0.25`
