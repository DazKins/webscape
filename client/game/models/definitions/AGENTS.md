# Code-Native Model Authoring

- Work in world units: one map tile is `1 x 1`. Put the model floor at local `y = 0`, center it around local `x = 0, z = 0`, and make forward local `+z`. The human is approximately `1.35` units tall.
- Build articulated models from named `THREE.Group` joints. Joint origins are pivots; child geometry should be positioned relative to that pivot. Use stable camelCase joint names such as `leftShoulder` and `doorHinge`.
- Use the helpers in `../primitives.ts` and favor boxes, low-segment cylinders/cones/spheres, toruses, and polyhedra. Keep geometry and materials instance-owned so `ModelInstance.dispose()` can release them.
- Animation samplers receive normalized time in `[0, 1]`. Poses are additive to the captured bind pose: position values are offsets, rotation values are XYZ Euler offsets in radians, and scale values are multipliers. Looping samplers must meet cleanly at phases `0` and `1`.
- Add model factories to `../registry.ts` without changing existing wire identifiers. Entity-derived options belong in `ModelOptions`; renderers should only position the model, pass options/state, and select or seek animations.
- Iterate with `npm run build`, then capture with `npm run model:screenshot -- --model human --animation run --phase 0.25`. Inspect `.model-previews/`, adjust the definition or sampler, rebuild, and recapture. Use `--all` for the registry gallery and never commit generated previews.
