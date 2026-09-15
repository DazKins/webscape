# World content implementation

The authored game content lives in [`game-project/`](../../game-project/game.json). See the [world overview](../game/world.md) for Willowbrook’s districts, characters, and landmarks.

## Authored equipment

Authored `equipped.slots` maps equipment slots to catalog ids, for example:

```json
{"equipped":{"slots":{"weapon":"ironSword","offhand":"woodenShield"}}}
```

The loader creates fresh items with catalog stats; the client receives normal equipment state. Server, schema, and editor reject invalid slot/item combinations.

## Validation

Run from the repository root: `go test ./...`. Run `pnpm run build` in both `client/` and `editor/`, then perform a local runtime/browser check.

`TestWillowbrookLandmarksAreReachable` loads the checked-in project and checks overlap, spawn clearance, and access to interactions, treating usable doors as open.
