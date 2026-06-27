# Testing Patterns

**Analysis Date:** 2026-06-27

## Test Framework

**Runner:**
- Go uses the standard `testing` package with adjacent `*_test.go` files under `server/`.
- There is no dedicated frontend test runner configured in `client/package.json` or `editor/package.json`.
- Frontend validation currently relies on TypeScript compilation plus Vite build.

**Assertion Library:**
- Go tests use standard `testing.T` assertions with `t.Fatal` and `t.Fatalf`.
- No third-party Go assertion library is present in `go.mod`.
- No Jest/Vitest/Playwright dependencies are present in the frontend `package.json` files.

**Run Commands:**
```bash
go test ./...                         # Run all Go tests
go test ./server/game/world           # Run one Go package
go test ./server/game/world -run TestLoadFromGameFSLoadsFirstMap
go build ./...                        # Compile all Go packages
cd client && npm run build            # Type-check and build playable client
cd editor && npm run build            # Type-check and build editor
```

## Test File Organization

**Location:**
- Go tests are placed beside the package under test: `server/game/world/world_test.go`, `server/game/system/systempathing_test.go`, `server/message/world_test.go`.
- No `tests/` tree is used.
- No frontend test files are present in the current repo.

**Naming:**
- Go unit/integration-style tests use `Test...` names that describe behavior, for example `TestLoadFromGameFSRejectsInvalidMapPath`, `TestInventoryAddItemEnforcesCapacity`, and `TestCombatSystemEmitsKillEventsForPlayerKill`.
- Helper functions use lower camelCase and call `t.Helper()`, as in `loadPathingTestWorld`, `assertPathNotFoundChatMessage`, and `setupConversationTestGame`.
- Test doubles use descriptive local types, such as `recordingEventEmitter` in `server/game/system/combat_event_test.go`.

**Structure:**
```text
server/
  game/
    conversation_test.go
    quest_test.go
    world/
      world.go
      world_test.go
    component/
      componentinventory.go
      componentinventory_test.go
    system/
      systempathing.go
      systempathing_test.go
      combat_event_test.go
  message/
    world.go
    world_test.go
client/
  package.json        # build is the validation path
editor/
  package.json        # build is the validation path
```

## Test Structure

**Suite Organization:**
```go
func TestInventoryAddItemEnforcesCapacity(t *testing.T) {
	inventory := NewCInventory()

	for i := 0; i < InventoryCapacity; i++ {
		if !inventory.AddItem(model.NewItem(fmt.Sprintf("Item %d", i), "test")) {
			t.Fatalf("AddItem returned false before capacity at item %d", i)
		}
	}

	if inventory.GetItemCount() != InventoryCapacity {
		t.Fatalf("inventory count = %d, want %d", inventory.GetItemCount(), InventoryCapacity)
	}
}
```

**Patterns:**
- Use explicit arrange/act/assert blocks through clear spacing rather than comments.
- Prefer `t.Fatalf` when subsequent assertions depend on setup or decoded data.
- Use `t.Fatal` for boolean invariants where formatted detail is unnecessary.
- Use `t.Helper()` in shared assertion/setup helpers.
- Keep tests focused on public package behavior unless the test is in the same package and intentionally covers an internal helper or system subroutine.

## Mocking

**Framework:**
- No mocking framework is used.
- Go tests use small hand-written fakes and closures.

**Patterns:**
```go
type recordingEventEmitter struct {
	events []gameevent.Event
}

func (r *recordingEventEmitter) EmitGameEvent(event gameevent.Event) {
	r.events = append(r.events, event)
}
```

**What to Mock:**
- Use `testing/fstest.MapFS` for authored game content and schema/runtime loader scenarios, as in `server/game/world/world_test.go` and `server/message/world_test.go`.
- Use closure-based senders/broadcasters to capture outbound messages in game tests, as in `server/game/conversation_test.go`.
- Use small interface fakes for event emitters or boundaries, as in `server/game/system/combat_event_test.go`.

**What NOT to Mock:**
- Do not mock pure validation, math, component, or format logic.
- Prefer real `ComponentManager`, `World`, and ECS components for system tests so interactions match runtime behavior.
- Do not add frontend test doubles unless a frontend test runner is introduced.

## Fixtures and Factories

**Test Data:**
```go
gameFS := fstest.MapFS{
	"game.json": {
		Data: []byte(`{
			"formatVersion": 1,
			"id": "test_game",
			"files": {
				"maps": ["maps/test.json"],
				"conversations": [],
				"quests": []
			}
		}`),
	},
	"maps/test.json": {
		Data: []byte(`{
			"formatVersion": 1,
			"id": "test",
			"size": { "x": 1, "y": 1 },
			"terrain": ["grass"]
		}`),
	},
}
```

**Location:**
- Fixtures are inline in each test file using JSON strings and `fstest.MapFS`.
- Reusable setup remains local to the test file unless multiple packages need it.
- Domain object creation uses real constructors such as `component.NewCInventory`, `model.NewItem`, `model.NewEntityId`, and `world.LoadFromGameFS`.

## Coverage

**Requirements:**
- No numeric coverage threshold is configured.
- Coverage emphasis is behavioral: content loading, path validation, conversations, quests, loot, spawn/openable interactions, combat event emission, collision, pathing, and component serialization.
- New server behavior should usually include focused Go tests near the changed package.

**Configuration:**
- No coverage config files are present.
- Use the Go toolchain if coverage is needed ad hoc.

**View Coverage:**
```bash
go test ./... -cover
go test ./server/game/... -coverprofile=/tmp/webscape-game.cover
go tool cover -func=/tmp/webscape-game.cover
```

## Test Types

**Unit Tests:**
- Component and utility tests exercise isolated behavior with real constructors, such as `server/game/component/componentinventory_test.go` and `server/game/collision/collision_test.go`.
- Message tests build real runtime objects and assert marshaled payload shape, as in `server/message/world_test.go`.
- System tests instantiate `ComponentManager` and the target system directly, then inspect components/events after `Update()` or helper calls.

**Integration Tests:**
- Loader and game-flow tests combine in-memory content, `World`, `Game`, ECS components, and outbound message capture.
- `server/game/conversation_test.go` is the current pattern for command-to-system-to-message flow without starting HTTP/WebSocket infrastructure.
- Content/schema-related changes should test invalid and valid authored JSON through `world.LoadFromGameFS`.

**E2E Tests:**
- No browser or WebSocket E2E framework is configured.
- Validate UI/client changes with builds at minimum; for runtime-content changes, rebuild the client and run `go run . -dev -game-folder game-project` for manual verification when appropriate.

## Common Patterns

**Async Testing:**
```go
func TestLoadFromGameFSLoadsFirstMap(t *testing.T) {
	world, err := LoadFromGameFS(gameFS)
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}

	if world.GetSizeX() != 2 || world.GetSizeY() != 1 {
		t.Fatalf("world size = (%d, %d), want (2, 1)", world.GetSizeX(), world.GetSizeY())
	}
}
```
- Current automated tests are synchronous Go tests.
- Avoid sleeping or starting real network servers for new tests unless the behavior cannot be isolated.

**Error Testing:**
```go
func TestLoadFromGameFSRejectsInvalidMapPath(t *testing.T) {
	if _, err := LoadFromGameFS(gameFS); err == nil {
		t.Fatal("LoadFromGameFS returned nil error for invalid map path")
	}
}
```
- Assert only that an error exists unless the exact message is part of the contract.
- Use specific fatal messages that state the observed behavior and expected behavior.

**Snapshot Testing:**
- Snapshot testing is not used.
- For JSON messages, unmarshal into a small local struct and assert the semantic fields, as in `server/message/world_test.go` and `server/game/conversation_test.go`.

---

*Testing analysis: 2026-06-27*
*Update when test patterns change*
