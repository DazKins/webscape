# Codebase Concerns

**Analysis Date:** 2026-06-27

## Tech Debt

**WebSocket command decoding is untyped and panic-prone:**
- Issue: Client command payloads are decoded into `map[string]any`, then most handlers use direct type assertions such as `cmd.Data["id"].(string)` and `cmd.Data["x"].(float64)`.
- Files: `server/command/command.go`, `server/commandhandler.go`
- Why: The current command surface grew from a small trusted client contract.
- Impact: Any malformed, missing, or wrong-typed WebSocket payload can panic the process before reaching game logic. This is both reliability debt and a security boundary problem.
- Fix approach: Replace `map[string]any` dispatch with per-command structs and safe decode/validate helpers; add table tests for invalid payloads in `server/command` and `server`.

**Game state is mutated directly from WebSocket goroutines and the tick loop:**
- Issue: `server/websocket.go` invokes the incoming message handler from each client `readPump`, while `server/game/game.go` runs `update()` on a ticker and all paths share `ComponentManager` maps without locks.
- Files: `server/websocket.go`, `server/server.go`, `server/game/game.go`, `server/game/component/manager.go`
- Why: Direct method calls are simple for early single-user development.
- Impact: Live clients can race the update loop, leading to data races, `concurrent map iteration and map write` panics, or inconsistent snapshots. The comment in `server/game/game.go` already calls out a join snapshot race.
- Fix approach: Funnel all client commands into the game loop through a channel, process them serially before/after systems, and keep outbound sends as queued side effects. Use `go test -race ./...` once this is in place.

**Editor is a large single component:**
- Issue: `editor/src/App.tsx` is about 2,167 lines and owns project state, file operations, tab state, validation display, map editing, conversation editing, and quest editing.
- Files: `editor/src/App.tsx`
- Why: The standalone editor has accumulated feature slices in one component.
- Impact: Format changes are harder to review safely; state coupling makes regressions likely when adding entities, schema-backed forms, or richer validation.
- Fix approach: Split into feature panels and state/update helpers by domain: project shell, map editor, conversation editor, quest editor, and validation/status components.

**Runtime content validation is duplicated across server, editor, and schemas:**
- Issue: Validation rules exist separately in Go loaders, TypeScript helpers, and OpenAPI files.
- Files: `server/game/world/world.go`, `server/game/world/conversation.go`, `server/game/world/quest.go`, `editor/src/worldFormat.ts`, `editor/src/conversationFormat.ts`, `editor/src/questFormat.ts`, `schema/*.openapi.yaml`
- Why: Each surface validates locally and there is no generated/shared schema harness.
- Impact: Changes to JSON formats can pass one surface and fail another, especially for authored entity component bags.
- Fix approach: Add fixture-driven compatibility tests that load `schema/examples/*` and `game-project/*` through server validators and editor validators; consider generating editor/server validation fixtures from the schema examples.

## Known Bugs

**Malformed commands can crash the server:**
- Symptoms: A WebSocket message such as `{"type":"move","data":{}}` or `{"type":"join","data":{"id":1}}` can panic on a direct type assertion.
- Trigger: Send a command with missing or wrong-typed `data` fields to `/ws`.
- Files: `server/commandhandler.go`
- Workaround: Only the bundled client sends well-formed commands during normal play.
- Root cause: Direct assertions at `server/commandhandler.go:48`, `server/commandhandler.go:64`, `server/commandhandler.go:71`, `server/commandhandler.go:77`, `server/commandhandler.go:78`, and `server/commandhandler.go:90`.
- Blocked by: Needs typed command decoding and negative tests.

**Unauthenticated move before join can crash the server:**
- Symptoms: A client can send `move` before `join`; `HandleMove` panics when `clientID` is not in `clientIdToEntityId`.
- Trigger: Connect to `/ws` and send a valid `move` command before a successful `join`.
- Files: `server/game/game.go`
- Workaround: The current client sends `join` on WebSocket open.
- Root cause: `server/game/game.go:539` to `server/game/game.go:548` uses panics for expected client-state errors.

**Editor accepts some maps that the server rejects:**
- Symptoms: A project can appear saveable in the editor but fail at server startup.
- Trigger: Create an entity with `metadata.width` or `metadata.height` that extends past map bounds, or a spawn child template containing a `position` component.
- Files: `editor/src/worldFormat.ts`, `server/game/world/world.go`
- Workaround: Run `go test ./...` or start the server after content edits.
- Root cause: Editor validation checks only the entity origin in `editor/src/worldFormat.ts:126` to `editor/src/worldFormat.ts:134`; server validation checks footprint bounds and spawn child template rules in `server/game/world/world.go:254` to `server/game/world/world.go:260`.

**WebSocket disconnect handling can panic if no disconnect handler is set:**
- Symptoms: `w.onDisconnect(client.id)` is called unconditionally on unregister.
- Trigger: A `wsServer` created without `SetDisconnectHandler` receives a disconnect.
- Files: `server/websocket.go`
- Workaround: Production startup in `server/server.go` does set the handler.
- Root cause: `server/websocket.go:87` assumes the callback is always installed.

## Security Considerations

**WebSocket accepts all origins:**
- Risk: Any website can open a WebSocket to the server from a user's browser and send game commands.
- Files: `server/websocket.go`
- Current mitigation: None in code; `CheckOrigin` returns `true` at `server/websocket.go:150` to `server/websocket.go:152`.
- Recommendations: Make allowed origins configurable, default to same-origin, and keep permissive behavior behind a development flag.

**No authentication or authorization on player identity:**
- Risk: Player identity comes from client-local `localStorage` and the client sends it in the `join` command. Another client can claim any UUID if known or guessed.
- Files: `client/main.ts`, `server/commandhandler.go`, `server/game/game.go`
- Current mitigation: UUID entropy reduces accidental collision; the server rejects only duplicate sessions for the same client connection mapping.
- Recommendations: Add server-issued sessions or authentication before persistent player state matters; treat client-supplied IDs as requests, not authority.

**Command input lacks size, rate, and content limits:**
- Risk: Large chat messages, rapid commands, or oversized JSON can consume memory/CPU and flood game update traffic.
- Files: `server/websocket.go`, `server/commandhandler.go`, `server/game/game.go`
- Current mitigation: WebSocket buffers are 1024 bytes, but no read limit, rate limit, chat length, or per-client command throttle is enforced.
- Recommendations: Set a WebSocket read limit, cap chat length, validate coordinate ranges, and rate-limit high-frequency command types per client.

**Game folder path is trusted process input:**
- Risk: `-game-folder` can point to any readable directory available to the process.
- Files: `main.go`, `server/game/world/world.go`
- Current mitigation: Project-internal manifest paths are checked with `fs.ValidPath` in `server/game/world/world.go:277` to `server/game/world/world.go:289`.
- Recommendations: This is acceptable for local/server operator input; if exposed through UI or deployment config, restrict it to an approved content root.

## Performance Bottlenecks

**Pathfinding recomputes full paths every tick:**
- Problem: `PathingSystem.Update` recalculates a path on every 500 ms tick for every pathing entity.
- Files: `server/game/system/systempathing.go`, `server/game/collision/collision.go`
- Measurement: No benchmark exists; code path is visible at `server/game/system/systempathing.go:95` to `server/game/system/systempathing.go:96`.
- Cause: The system discards the cached path and calls `GetPath` each update.
- Improvement path: Reuse paths until target or blockers change; add benchmarks for `collision.Checker.GetPath` across representative map sizes and entity counts.

**Collision checks scan all positioned entities:**
- Problem: Each `IsBlocked` call loops over every entity with a `position` component.
- Files: `server/game/collision/collision.go`
- Measurement: No benchmark exists; the linear scan is at `server/game/collision/collision.go:28` to `server/game/collision/collision.go:37`.
- Cause: No spatial index or occupancy grid for dynamic blockers.
- Improvement path: Maintain a dynamic blocker grid or spatial hash updated when position, metadata, or openable state changes.

**Game update diff walks all serializable components every tick:**
- Problem: Every tick serializes and compares all serializable components, then separately detects removals.
- Files: `server/game/game.go`
- Measurement: No benchmark exists; the scan starts at `server/game/game.go:168` to `server/game/game.go:208`.
- Cause: There is no dirty-component tracking in `ComponentManager`.
- Improvement path: Add dirty flags or a mutation journal in `ComponentManager`, then emit game updates from changed component IDs only.

**World/client snapshots are sent as one payload on join:**
- Problem: New clients receive full world data plus a full component snapshot immediately after join.
- Files: `server/game/game.go`, `server/message/world.go`, `client/game/game.ts`
- Measurement: No payload-size measurement exists; sample map content is already sizable in `game-project/maps/starter.json`.
- Cause: The protocol has no chunking, compression awareness, or lazy world/entity streaming.
- Improvement path: Add payload-size logging in development, then chunk or cache static world data if maps grow.

## Fragile Areas

**Authoritative game loop and ECS mutation boundary:**
- Why fragile: Systems, command handlers, quest/event handling, and WebSocket join/leave all mutate the same component maps directly.
- Files: `server/game/game.go`, `server/game/component/manager.go`, `server/game/system/*.go`, `server/websocket.go`
- Common failures: Races during join snapshots, panics while maps are iterated, dropped updates, stale interactions.
- Safe modification: Introduce a command queue and make the tick loop the only writer before adding new commands or systems.
- Test coverage: Go tests cover many game behaviors, but not concurrent WebSocket/tick interaction or race detection.

**Content format compatibility:**
- Why fragile: `game.json`, world files, conversations, quests, editor validators, Go loaders, and OpenAPI schemas must change together.
- Files: `schema/`, `editor/src/*Format.ts`, `server/game/world`, `game-project/`
- Common failures: Editor saves content that server rejects; schema examples drift from loader behavior; runtime accepts values not documented for authors.
- Safe modification: For each format change, update all four surfaces and add a fixture that is validated by both Go and TypeScript tooling.
- Test coverage: Server loader tests exist in `server/game/world/world_test.go`; there are no frontend/editor tests that exercise validators.

**Authored entity component bags:**
- Why fragile: The schema leaves `components` mostly open-ended while runtime conversion silently ignores invalid component shapes.
- Files: `schema/world-format.openapi.yaml`, `server/game/entity/entityworldobject.go`, `editor/src/worldFormat.ts`
- Common failures: Component typo silently drops behavior; invalid loot or combat stats default unexpectedly; missing conversation references skip entities.
- Safe modification: Add explicit component schemas for every supported authored component and server validation that fails instead of silently dropping critical malformed components.
- Test coverage: `server/game/entity/entityworldobject_test.go` covers some conversion paths, but not the full authored component matrix.

**Conversation and quest event ids:**
- Why fragile: Quest advancement depends on exact string ids emitted by conversation/combat/loot systems.
- Files: `server/game/game.go`, `server/game/world/quest.go`, `server/game/world/conversation.go`, `game-project/quests/first_errand.json`
- Common failures: Renaming a conversation/node/metadata token breaks quest progression without compile-time feedback.
- Safe modification: Add validation that quest event ids referencing known content resolve where possible, especially `conversation:node:<conversationId>:<nodeId>` and known kill/collect prefixes.
- Test coverage: Quest tests cover runtime advancement, but no static cross-reference validation exists for all authored content.

## Scaling Limits

**Single-process authoritative server:**
- Current capacity: Not measured.
- Limit: All clients, ECS state, WebSocket connections, and map content live in one Go process on port `8080`.
- Symptoms at limit: Increased tick latency, slow joins, dropped WebSocket sends when per-client queues fill, and process-wide failure on panic.
- Scaling path: Add tick duration/client count metrics, isolate panics, then introduce rooms/shards only after the single-loop command queue is stable.

**Single loaded map:**
- Current capacity: Runtime loads the first map from `files.maps`.
- Limit: Multi-map projects can be authored but only the first map is active at runtime.
- Symptoms at limit: Additional maps in `game.json` are ignored by the server during play.
- Scaling path: Define map selection/transition semantics and update `server/game/world/world.go`, client world registration, editor affordances, and schemas together.

**Browser editor memory and UX scale:**
- Current capacity: Not measured; current sample content loads fully into React state.
- Limit: Large maps or many conversations/quests will make `editor/src/App.tsx` harder to render and maintain.
- Symptoms at limit: Slow editor updates, hard-to-navigate panels, expensive whole-document validation.
- Scaling path: Split editor state by document, memoize derived summaries carefully, and add focused editor tests before larger authoring features.

## Dependencies at Risk

**Gorilla WebSocket origin/auth behavior is entirely application-owned:**
- Risk: `github.com/gorilla/websocket` is stable, but it does not provide application-level auth, origin policy, message validation, or rate limiting.
- Files: `go.mod`, `server/websocket.go`
- Impact: Security posture depends on local code that is currently permissive.
- Migration plan: Keep the dependency, but wrap upgrade/read handling in local policy middleware and tests.

**Vite major versions differ between client and editor manifests:**
- Risk: `client/package.json` pins `vite` to `7.1.9`, while `editor/package.json` allows `^7.3.3`.
- Files: `client/package.json`, `editor/package.json`
- Impact: Build behavior can diverge across the two frontends despite similar stacks.
- Migration plan: Align Vite and React tooling versions intentionally, then run both `cd client && npm run build` and `cd editor && npm run build`.

**Go toolchain version is very new:**
- Risk: `go.mod` declares `go 1.25.2`.
- Files: `go.mod`
- Impact: Contributors and CI images need a matching modern Go toolchain; older default Go installations will fail.
- Migration plan: Document the required Go version in setup docs/CI or lower only if the codebase does not require 1.25 features.

## Missing Critical Features

**No command protocol tests:**
- Problem: The WebSocket command surface is a trust boundary but has no invalid-input test coverage.
- Files: `server/command/command.go`, `server/commandhandler.go`
- Current workaround: Manual use through the browser client.
- Blocks: Safe expansion of command types and hardening against malformed clients.
- Implementation complexity: Low to medium; add typed decode helpers and table tests.

**No frontend/editor automated tests:**
- Problem: Both TypeScript apps rely on `tsc`/Vite builds for validation.
- Files: `client/`, `editor/`
- Current workaround: Manual browser testing plus `npm run build`.
- Blocks: Safe refactors of editor validation, HUD state, WebSocket message handling, and renderer behavior.
- Implementation complexity: Medium; start with Vitest unit tests for format helpers and WebSocket message reducers before adding browser tests.

**No schema validation command in CI/tooling:**
- Problem: OpenAPI schemas and examples are present, but there is no repo command that validates examples or game content against schemas.
- Files: `schema/`, `game-project/`
- Current workaround: Server loader tests cover selected fixtures.
- Blocks: Confident format evolution across editor/server/schema.
- Implementation complexity: Medium; add a schema validation script and include it in documented verification.

**No observability for runtime health:**
- Problem: There are no metrics for tick duration, connected clients, command rates, WebSocket queue drops, payload sizes, or panic recovery.
- Files: `server/game/game.go`, `server/websocket.go`, `server/server.go`
- Current workaround: Logs only.
- Blocks: Diagnosing performance and stability issues under real play.
- Implementation complexity: Medium; start with structured log counters or `/debug` metrics guarded for development.

## Test Coverage Gaps

**Command handling invalid inputs:**
- What's not tested: Missing fields, wrong JSON types, unknown command types, command-before-join, coordinate bounds, and chat length.
- Files: `server/commandhandler.go`, `server/game/game.go`
- Risk: Remote clients can crash or stress the server.
- Priority: High.
- Difficulty to test: Low after command decoding is made typed and error-returning.

**Concurrent game mutation:**
- What's not tested: WebSocket read goroutines mutating ECS while the tick loop updates systems and broadcasts diffs.
- Files: `server/websocket.go`, `server/game/game.go`, `server/game/component/manager.go`
- Risk: Race conditions and map panics under multiple clients.
- Priority: High.
- Difficulty to test: Medium; requires a command queue design or race-focused integration harness.

**Editor validation parity:**
- What's not tested: TypeScript validators against the same fixtures used by Go world/conversation/quest loaders.
- Files: `editor/src/worldFormat.ts`, `editor/src/conversationFormat.ts`, `editor/src/questFormat.ts`, `server/game/world`
- Risk: Authors can save projects that fail at runtime.
- Priority: High.
- Difficulty to test: Medium because no frontend test runner is configured yet.

**WebSocket message payload contracts:**
- What's not tested: Client handling of `gameUpdate`, `world`, `conversation`, `questCompleted`, and malformed/unknown server messages.
- Files: `server/message/`, `client/game/game.ts`, `client/main.ts`
- Risk: Server/client payload drift breaks UI state silently.
- Priority: Medium.
- Difficulty to test: Medium; add JSON fixture tests shared between Go message tests and TypeScript handlers.

**Renderer and UI behavior:**
- What's not tested: Three.js renderer creation, entity removal cleanup, interaction menu behavior, inventory/equipment UI, conversation panel, and quest overlay.
- Files: `client/game/renderer/`, `client/ui/components/`
- Risk: Visual regressions or stale entities after gameplay changes.
- Priority: Medium.
- Difficulty to test: Medium to high; start with component-level tests and use Playwright only for end-to-end smoke coverage.

**Schema examples and sample content:**
- What's not tested: All `schema/examples/*.json` examples and checked-in `game-project/` content against both schema and runtime validators in one command.
- Files: `schema/examples/`, `game-project/`, `server/game/world`
- Risk: Documentation/examples drift from executable content.
- Priority: Medium.
- Difficulty to test: Low to medium; add fixture discovery tests and schema validation tooling.

---

*Concerns audit: 2026-06-27*
*Update as issues are fixed or new ones discovered*
