# Webscape Terrain Elevation

## What This Is

Webscape is a content-driven browser game server with a playable Three.js client and a standalone browser editor for authoring game projects. The next milestone adds per-tile terrain height so authored maps can include visible elevation in the game and direct 3D height editing in the editor.

The first elevation version is visual-only for gameplay: height changes how terrain is rendered and authored, but it does not change movement, interactions, combat, line-of-sight, or entity placement rules yet.

## Core Value

Authors can create and inspect clearly stepped 3D terrain heights without making existing maps or gameplay rules harder to maintain.

## Requirements

### Validated

- ✓ Content-driven game loading from `game-project/game.json`, map files, conversations, and quests — existing
- ✓ Server-authoritative Go ECS runtime with JSON commands over `/ws` — existing
- ✓ Playable Three.js client renders world state and entity components — existing
- ✓ Standalone React/TypeScript editor opens and saves project folders through the browser File System Access API — existing
- ✓ Map, project, conversation, and quest formats are represented across `schema/`, `server/game/world/`, `editor/src/*Format.ts`, and `game-project/` examples — existing

### Active

- [ ] Add per-tile terrain height to authored map data with a fixed v1 range of `0-10`.
- [ ] Default or migrate all existing maps so every tile behaves as height `0`.
- [ ] Render height in the playable client as clearly stepped terrain.
- [ ] Add real 3D editor support so authors can paint/select tile heights directly in a 3D editing surface.
- [ ] Keep height visual-only in v1; movement, interactions, combat, placement, and line-of-sight continue to use existing 2D rules.
- [ ] Keep schemas, server loaders, editor validators, and sample content in sync for the terrain-height format.

### Out of Scope

- Gameplay elevation effects — height does not affect movement, blocking, combat, pathing, line-of-sight, or interaction reach in v1.
- Smooth slopes or blended terrain — v1 uses clear stepped terrain rather than slope interpolation.
- Per-map configurable height ranges — v1 uses a fixed `0-10` range for straightforward validation and UI.
- Full 3D game physics — terrain height is authored and rendered, not a new physics model.

## Context

The current runtime loads a single authored map from `game-project/game.json` and treats map tile arrays as row-major data. Format changes must preserve the repository contract: update `schema/`, editor format helpers in `editor/src/`, server loading and validation in `server/game/world/`, and checked-in examples in `game-project/` together.

The playable client already uses Three.js under `client/game/`, so stepped terrain should extend the existing rendering surface rather than introduce a second rendering stack. The editor is currently a standalone React app under `editor/src/App.tsx`; this milestone changes the editor from JSON/2D-oriented map authoring toward a true 3D editing surface for terrain height.

Existing authored maps must keep working. Missing or pre-existing tile height data should be treated as `0` during migration/defaulting so old content loads and renders flat until authors edit heights.

## Constraints

- **Height model**: Tile height is an integer from `0` through `10` — keeps validation, rendering, and editing controls bounded for v1.
- **Gameplay semantics**: Height is visual-only in v1 — avoids coupling terrain rendering to pathing, blockers, interactions, combat, and entity placement in the first milestone.
- **Rendering style**: Use stepped terrain — authors and players need to read height clearly, not infer it from subtle smoothing.
- **Editor experience**: The editor needs real 3D editing support — authors should paint/select tile heights directly in a 3D view, not only inspect JSON fields.
- **Backward compatibility**: Existing maps default to height `0` — prevents content migrations from breaking the sample project or older authored maps.
- **Format synchronization**: JSON schema, editor validators, Go loaders, and sample content must change together — prevents mismatched editor/runtime behavior.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Terrain height is visual-only for v1 | Keeps the first milestone focused on authoring and rendering without changing server gameplay semantics | - Pending |
| v1 height range is fixed at `0-10` | Gives enough visible range while keeping UI and validation simple | - Pending |
| Existing maps default to height `0` | Keeps current game content compatible and flat until edited | - Pending |
| Client renders height as stepped terrain | Stepped geometry makes height obvious and avoids slope complexity | - Pending |
| Editor needs direct 3D editing support | The authoring workflow should let map creators paint/select height where they see it | - Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `$gsd-transition`):
1. Requirements invalidated? -> Move to Out of Scope with reason
2. Requirements validated? -> Move to Validated with phase reference
3. New requirements emerged? -> Add to Active
4. Decisions to log? -> Add to Key Decisions
5. "What This Is" still accurate? -> Update if drifted

**After each milestone** (via `$gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-06-27 after initialization*
