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

Armour uses `{ equipped: true }` to select a fitted variant; the default remains
the inventory/ground model. `equipment.ts` maps its root and named parts to human
joint sockets. Parts use joint-local coordinates and move with the host's pose.
The attachment controller owns those detached parts and restores covered clothing
on removal, including when several items cover the same body part. Hide clothing
meshes/groups, never the joints carrying the armour. The model lab uses this same
controller and accepts comma-separated equipment names to preview a complete set.

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

Terrain appearances live in `../world/terrainAppearance.ts`. `grass` and
`grassShort` use short tufts; `grassLong` uses taller, denser tufts (4–12 per tile);
`stone` adds tiny pebbles. Other types retain bare surfaces. Tile colours and
scatter use deterministic global-coordinate/type seeds. Colours vary only in
brightness (up to 4.5% for grass), and geometry does not change on reload.
Details follow the rendered terrain triangles and are cosmetic: they do not
participate in collision, picking, or server state. The editor exposes these
identifiers with matching base-colour swatches; it does not preview the scatter.

The chunk worker produces one extra opaque vertex-coloured mesh, with no
animation or shadow casting. Maximum detail triangles per tile: short grass 12,
long grass 36, stone 15. Long grass averages 24 detail triangles per tile;
short grass and stone average roughly half their caps. Chunk culling
and disposal apply to the whole detail mesh. No per-tuft objects are retained.
The east meadow has a short/long grass sample beside its northern stone court.
