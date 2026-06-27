<!-- refreshed: 2026-06-27 -->
# Architecture

**Analysis Date:** 2026-06-27

## System Overview

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│                         Browser Applications                                │
├──────────────────────────────────────┬───────────────────────────────────────┤
│ Playable client                      │ Standalone project editor             │
│ `client/main.ts`                     │ `editor/src/App.tsx`                  │
│ `client/game/game.ts`                │ `editor/src/fileSystem.ts`            │
└───────────────────┬──────────────────┴───────────────────┬──────────────────┘
                    │ WebSocket JSON commands               │ File System Access API
                    ▼                                        ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                              Go HTTP/WebSocket Server                        │
│ `main.go` -> `server/server.go` -> `server/websocket.go`                    │
│ command routing in `server/commandhandler.go`                               │
└────────────────────────────────────────────┬─────────────────────────────────┘
                                             ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                         Authoritative ECS Runtime                            │
│ `server/game/game.go` owns state, systems, commands, and broadcasts          │
│ components: `server/game/component/`                                         │
│ systems: `server/game/system/`                                               │
└────────────────────────────────────────────┬─────────────────────────────────┘
                                             ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                            Authored Content                                  │
│ `game-project/game.json` -> `game-project/maps/`, conversations, quests      │
│ schemas in `schema/`, editor validators in `editor/src/*Format.ts`           │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| CLI entry | Parse `-dev` and `-game-folder`, choose embedded or live `client/dist`, load authored content | `main.go` |
| HTTP/WebSocket setup | Serve static client files, create `Game`, wire command handling and broadcast/sender callbacks | `server/server.go` |
| WebSocket transport | Manage client UUIDs, read/write pumps, broadcasts, direct sends, connect/disconnect lifecycle | `server/websocket.go` |
| Command boundary | Decode client JSON commands and route to game handlers | `server/command/command.go`, `server/commandhandler.go` |
| ECS runtime | Own authoritative world state, component manager, systems, player sessions, quest/conversation events, game update deltas | `server/game/game.go` |
| Component store | Store components by component id and entity id | `server/game/component/manager.go` |
| Systems | Mutate ECS state each tick for pathing, interaction, combat, random walk, TTL, health, and spawn behavior | `server/game/system/` |
| Authored entity conversion | Convert JSON component bags into concrete Go components | `server/game/entity/entityworldobject.go` |
| Content loading | Load `game.json`, map, conversation, and quest registries from an `fs.FS` | `server/game/world/world.go`, `server/game/world/conversation.go`, `server/game/world/quest.go` |
| Outgoing messages | Wrap typed server payloads with message metadata | `server/message/` |
| Playable client shell | Create Three.js scene, React HUD, WebSocket client, and animation loop | `client/main.ts` |
| Client game model | Mirror server component updates into local entities, drive rendering, input, UI events, and commands | `client/game/game.ts` |
| Entity rendering | Select renderers by `renderable.type` and keep Three.js objects in sync with local entities | `client/game/entityRenderSystem.ts`, `client/game/renderer/` |
| React HUD | Subscribe to `Game` events and render panels for chat, combat, inventory, equipment, quests, conversations, and overlays | `client/ui/uiRoot.tsx`, `client/ui/components/` |
| Editor app | Edit project, map, conversation, and quest JSON documents in browser state | `editor/src/App.tsx` |
| Editor file I/O | Open/save project folders through the browser File System Access API | `editor/src/fileSystem.ts` |
| Format contracts | Define schemas, runtime validators, editor validators, and sample authored data | `schema/`, `server/game/world/`, `editor/src/*Format.ts`, `game-project/` |

## Pattern Overview

**Overall:** Content-driven Go authoritative game server with two Vite/React browser apps.

**Key Characteristics:**
- Server-authoritative ECS: clients send commands; `server/game/game.go` mutates state and broadcasts component deltas.
- Content-driven runtime: `game-project/game.json` lists maps, conversations, and quests; loaders in `server/game/world/` validate and register them.
- JSON wire contracts: commands live in `server/command/command.go` and `client/command/command.ts`; outgoing messages live in `server/message/` and are consumed by `client/main.ts`.
- Separate browser experiences: `client/` is the playable runtime; `editor/` is a standalone project editor and does not run through the Go server.
- Format changes are multi-surface: update `schema/`, `server/game/world/`, `editor/src/*Format.ts`, and `game-project/` together.

## Layers

**Startup and Hosting Layer:**
- Purpose: Start the application, load content, and expose HTTP/WebSocket endpoints.
- Location: `main.go`, `server/server.go`
- Contains: flag parsing, embedded/live static file selection, `world.LoadFromGameFolder`, HTTP handlers.
- Depends on: `server/game/world`, `server`, `io/fs`, `net/http`.
- Used by: local development commands, production binary, Docker image.

**Transport Layer:**
- Purpose: Move JSON commands and messages between browser clients and the authoritative server.
- Location: `server/websocket.go`, `client/ws.ts`
- Contains: WebSocket connection setup, reconnects, read/write pumps, send queues, broadcast/direct-send callbacks.
- Depends on: Gorilla WebSocket on the server, browser `WebSocket` on the client.
- Used by: `server/server.go`, `client/main.ts`, `client/game/game.ts`.

**Command Layer:**
- Purpose: Convert client intent into typed game operations.
- Location: `server/command/command.go`, `server/commandhandler.go`, `client/command/command.ts`
- Contains: command type constants, generic command payloads, route handlers for join, move, chat, interact, equip, unequip, and conversation options.
- Depends on: `server/game`, `server/game/model`, `server/game/component`.
- Used by: WebSocket message handler in `server/server.go`; UI/input methods in `client/game/game.ts`.

**Authoritative ECS Layer:**
- Purpose: Own all mutable game state and decide what clients may observe.
- Location: `server/game/`
- Contains: `Game`, `ComponentManager`, component structs, systems, game event emission, quest progress, conversation state, inventory/equipment, entity factories.
- Depends on: loaded `world.World`, component manager, systems, message factories.
- Used by: `server/commandhandler.go` and per-tick update loop.

**Content and Format Layer:**
- Purpose: Load, validate, and expose authored project data.
- Location: `server/game/world/`, `game-project/`, `schema/`, `editor/src/*Format.ts`
- Contains: game manifest format, world/map format, conversation registry, quest registry, editor normalizers/validators, OpenAPI schemas and examples.
- Depends on: JSON decoding, `fs.ValidPath`, editor file system helpers.
- Used by: `main.go`, `server/game/game.go`, editor save/open flows, future content/schema changes.

**Message Serialization Layer:**
- Purpose: Build typed outgoing JSON messages from server state.
- Location: `server/message/`
- Contains: metadata wrapper, world snapshot, game updates, entity removal, join/join failure, conversation, quest completed messages.
- Depends on: component serialization and world registry data.
- Used by: `server/game/game.go` and `server/websocket.go`.

**Playable Client Runtime Layer:**
- Purpose: Render the game and translate input/UI interactions into commands.
- Location: `client/`
- Contains: `Game` event target, local `Entity` component bags, Three.js scene/camera/renderers, world geometry, input router, React HUD components.
- Depends on: browser APIs, Three.js, React, server message shapes.
- Used by: browser entry `client/main.ts`.

**Editor Layer:**
- Purpose: Author and validate project JSON files in a browser.
- Location: `editor/src/`
- Contains: React app state, map editing tools, conversation/quest editors, project file APIs, JSON serialization and validation.
- Depends on: browser File System Access API and local format helper modules.
- Used by: standalone editor Vite app.

## Data Flow

### Server Startup Path

1. Process starts at `main.go`, parses `-dev` and `-game-folder`.
2. Static file system is selected: live `client/dist` in dev mode or embedded `client/dist` via `embed.FS` in production (`main.go`).
3. `loadGameWorld` requires `-game-folder` and calls `world.LoadFromGameFolder` (`main.go`).
4. `world.LoadFromGameFS` reads `game.json`, validates paths, loads conversation and quest registries, then loads the first map from `files.maps` (`server/game/world/world.go`).
5. `server.Start` serves the static file system, creates `game.NewGameWithWorld`, starts the update loop, wires WebSocket handlers, and listens on `:8080` (`server/server.go`).

### Client Command Path

1. `client/main.ts` creates a persistent player id in `localStorage`, starts `WebSocketClient`, and sends `join` on connect.
2. Client input and UI methods call `createCommand` and `wsClient.sendMessage` for move, chat, interact, equip, unequip, and conversation choices (`client/game/game.ts`).
3. `server/websocket.go` reads text frames and passes raw JSON to the incoming message handler.
4. `server/command.Unmarshal` decodes into `Command{Type, Data}` (`server/command/command.go`).
5. `ClientCommandHandler.HandleCommand` routes by command type and calls `Game` handlers such as `HandleMove`, `HandleInteract`, or `HandleConversationOption` (`server/commandhandler.go`).
6. `server/game/game.go` mutates ECS components; systems process effects on the next tick.

### Game Tick and Broadcast Path

1. `Game.StartUpdateLoop` runs `update` every 500 ms (`server/game/game.go`).
2. Registered systems run in order: pathing, interaction, combat, random walk, TTL, health, spawn (`server/game/game.go`).
3. Serializable components are compared against `prevSerialisedComponents` and only changed component payloads are collected (`server/game/game.go`).
4. Removed components and completely removed entities are detected from previous serialized state (`server/game/game.go`).
5. `message.NewGameUpdateMessage` serializes per-entity component updates plus available interactions (`server/message/gameupdate.go`).
6. `wsServer.Broadcast` sends marshaled message JSON to all connected clients (`server/websocket.go`).
7. `client/main.ts` dispatches `gameUpdate` to `Game.handleGameUpdate`, which updates local entities and emits UI events (`client/game/game.ts`).
8. `EntityRenderSystem.update` creates/updates/removes renderers for entities with a `renderable` component (`client/game/entityRenderSystem.ts`).

### Authored Content Flow

1. `game-project/game.json` lists map, conversation, and quest file paths.
2. The server enforces project path validity with `fs.ValidPath` and rejects empty or missing required structures (`server/game/world/world.go`).
3. Maps decode into `worldFormat`, including row-major terrain/blockers, walls, and authored entity component bags (`server/game/world/world.go`).
4. `Game.loadWorldEntities` skips entities with `playerSpawn`, validates referenced conversations, and passes authored component bags to `entity.CreateAuthoredEntity` (`server/game/game.go`).
5. `CreateAuthoredEntity` maps known JSON components into concrete server components such as position, metadata, renderable, openable, lootable, conversation, randomwalk, health, basestats, equipped, and combatstats (`server/game/entity/entityworldobject.go`).
6. The editor normalizes and validates the same practical file formats before saving through `editor/src/gameProject.ts`, `editor/src/worldFormat.ts`, `editor/src/conversationFormat.ts`, and `editor/src/questFormat.ts`.

### Conversation and Quest Flow

1. Entities with `conversation.conversationId` are converted into `CConversation` components (`server/game/entity/entityworldobject.go`).
2. Interacting with a talk target sets pathing/interacting components; `InteractionSystem` starts the conversation when in range (`server/game/system/systeminteraction.go`).
3. `Game.StartConversationFor` creates `CActiveConversation`, sends the start node, and emits `conversation:node:<conversationId>:<nodeId>` events (`server/game/game.go`).
4. `conversationOption` commands advance the active node and can end the conversation (`server/game/game.go`).
5. Quest start and step progress are driven by generic game event ids in `Game.EmitGameEvent` (`server/game/game.go`).
6. Quest completion updates `CQuestLog`, delivers rewards to inventory or dropped entities, sends a `questCompleted` message, and emits `quest:completed:<questId>` (`server/game/game.go`).

**State Management:**
- Server state is authoritative and stored in-memory in `Game.componentManager` plus `clientIdToEntityId` and loaded `world.World` registries (`server/game/game.go`).
- Client state is a local mirror in `client/game/entity/entity.ts` and `client/game/game.ts`; it must be treated as presentation state only.
- Editor state is React component state in `editor/src/App.tsx`; persistence happens only through explicit save flows in `editor/src/fileSystem.ts`.

## Key Abstractions

**Game:**
- Purpose: Coordinate ECS state, systems, player sessions, commands, world registries, and outgoing messages.
- Examples: `server/game/game.go`
- Pattern: Central authoritative service with injected sender/broadcaster callbacks.

**Component:**
- Purpose: Represent entity state by lowercase wire component id.
- Examples: `server/game/component/componentposition.go`, `server/game/component/componentinventory.go`, `server/game/component/componentquestlog.go`
- Pattern: Go structs implement `Component`; client receives serialized component bags keyed by `ComponentId`.

**SerializeableComponent:**
- Purpose: Mark components that should be sent to clients.
- Examples: `server/game/component/component.go`, component files with `Serialize()` implementations.
- Pattern: Components without `Serialize()` remain server-only.

**System:**
- Purpose: Apply per-tick logic to entities with relevant components.
- Examples: `server/game/system/systempathing.go`, `server/game/system/systeminteraction.go`, `server/game/system/systemcombat.go`
- Pattern: `Update()` methods read and write via `ComponentManager`.

**World Registry:**
- Purpose: Hold loaded content that is not represented as mutable ECS components.
- Examples: `server/game/world/world.go`, `server/game/world/conversation.go`, `server/game/world/quest.go`
- Pattern: JSON files load into registries and are referenced by component ids or game event ids.

**Message:**
- Purpose: Standardize outgoing WebSocket payloads with metadata and data.
- Examples: `server/message/message.go`, `server/message/type.go`, `server/message/gameupdate.go`
- Pattern: Message factory functions create typed payloads, then `Message.Marshal()` JSON encodes them.

**Client Game:**
- Purpose: Bridge wire messages, input, rendering, and UI events.
- Examples: `client/game/game.ts`
- Pattern: `EventTarget` subclass used by React components and the Three.js renderer.

**Editor Format Helper:**
- Purpose: Normalize, validate, and serialize JSON documents for authoring.
- Examples: `editor/src/gameProject.ts`, `editor/src/worldFormat.ts`, `editor/src/conversationFormat.ts`, `editor/src/questFormat.ts`
- Pattern: Pure TypeScript functions paired with editor React state and server loader expectations.

## Entry Points

**Server Binary:**
- Location: `main.go`
- Triggers: `go run . -dev -game-folder game-project`, `go run . -game-folder game-project`, or Docker command.
- Responsibilities: parse flags, choose static file source, load game content, start server.

**HTTP/WebSocket Server:**
- Location: `server/server.go`
- Triggers: `main.go` calls `server.Start`.
- Responsibilities: serve `/`, serve `/ws`, create and wire `Game`, start update loop.

**WebSocket Connection Handler:**
- Location: `server/websocket.go`
- Triggers: browser connects to `/ws`.
- Responsibilities: assign client id, register/unregister connection, run read/write pumps.

**Command Handler:**
- Location: `server/commandhandler.go`
- Triggers: WebSocket incoming message handler after `command.Unmarshal`.
- Responsibilities: validate/coerce command fields and call game methods.

**Game Loop:**
- Location: `server/game/game.go`
- Triggers: `Game.StartUpdateLoop`.
- Responsibilities: update systems, compute deltas, broadcast updates and removals.

**Playable Client:**
- Location: `client/main.ts`
- Triggers: browser loads built Vite app from Go server or Vite dev server.
- Responsibilities: create `Game`, mount React HUD, connect WebSocket, route incoming messages, run animation loop.

**Editor Client:**
- Location: `editor/src/main.tsx`, `editor/src/App.tsx`
- Triggers: browser loads editor Vite app.
- Responsibilities: mount editor UI, manage project documents, open/save through directory handles.

## Architectural Constraints

- **Threading:** Server update loop and WebSocket read/write pumps run in goroutines; `ComponentManager` and `Game` state do not expose general locking, so new cross-goroutine mutations should route through existing game methods with care.
- **Global state:** `http.Handle` uses the default HTTP mux in `server/server.go`; WebSocket `upgrader` is package-level in `server/websocket.go`.
- **Persistence:** Runtime game state is in-memory only. Authored content is loaded from `-game-folder`; the editor writes project JSON through browser file handles.
- **Static assets:** `client/dist` is embedded by `main.go` for production and served from disk in `-dev` mode; do not manually edit generated output.
- **Content path safety:** Project file paths must be relative, slash-separated, and valid for `fs.ValidPath` on the server and `isValidProjectPath` in the editor (`server/game/world/world.go`, `editor/src/gameProject.ts`).
- **Map indexing:** Authored tile arrays are row-major (`y * size.x + x`) in both server blocker conversion and editor helpers (`server/game/world/world.go`, `editor/src/worldFormat.ts`).
- **Wire compatibility:** Component ids, command types, message types, renderable types, and event ids are string contracts across Go, TypeScript, schema, and content.
- **Single-map runtime:** The server currently loads the first entry in `files.maps` (`server/game/world/world.go`); editor supports multiple map files in project state.

## Anti-Patterns

### Client-Authoritative Game Logic

**What happens:** Adding rules only in `client/game/game.ts` or React components makes the browser decide outcomes.
**Why it's wrong:** The server owns state and broadcasts authoritative component updates.
**Do this instead:** Put gameplay rules in `server/game/game.go` or a focused system under `server/game/system/`, then expose the result through serializable components or messages in `server/message/`.

### Format Changes in One Surface Only

**What happens:** Updating only `schema/` or only `editor/src/*Format.ts` creates content the runtime cannot load, or runtime content the editor rejects.
**Why it's wrong:** Project files cross four surfaces: schemas, editor validators, server loaders, and sample content.
**Do this instead:** Update `schema/`, `editor/src/*Format.ts`, `server/game/world/`, and `game-project/` together for format changes.

### Ad Hoc Wire Strings

**What happens:** New command, message, component, renderable, or event strings are added in one language without matching consumers.
**Why it's wrong:** The Go server and TypeScript client communicate through stringly typed JSON contracts.
**Do this instead:** Add command types in `server/command/command.go` and `client/command/command.ts`; add outgoing message types in `server/message/type.go` and handle them in `client/main.ts`; add renderer support in `client/game/entityRenderSystem.ts` for new `renderable.type` values.

### Direct Edits to Generated Builds

**What happens:** Files under `client/dist` or `editor/dist` are edited by hand.
**Why it's wrong:** Build output is generated and overwritten by Vite builds.
**Do this instead:** Edit source under `client/` or `editor/src/`, then run the relevant build command.

## Error Handling

**Strategy:** Fail fast during startup/content loading; log and ignore malformed runtime commands where practical; panic only in some unexpected connected-player paths.

**Patterns:**
- Startup and content errors return wrapped errors from loaders and terminate through `log.Fatal` in `main.go`.
- Format loaders return validation errors with content path context (`server/game/world/world.go`, `server/game/world/conversation.go`, `server/game/world/quest.go`).
- Command handler field extraction is mostly unchecked type assertions for known client payloads; new commands should validate type assertions before use in `server/commandhandler.go`.
- WebSocket parse errors are logged and dropped in `server/server.go`.
- Client JSON parse and WebSocket errors are logged in `client/ws.ts`.
- Editor open/save errors are caught and surfaced as status text in `editor/src/App.tsx`.

## Cross-Cutting Concerns

**Logging:** Server uses `log.Printf`/`log.Println` in `main.go`, `server/server.go`, `server/websocket.go`, `server/commandhandler.go`, and `server/game/game.go`. Client uses `console.error`/`console.warn` in transport and renderer selection.

**Validation:** Server validates game, world, conversation, and quest documents in `server/game/world/`. Editor validates before saving in `editor/src/gameProject.ts`, `editor/src/worldFormat.ts`, `editor/src/conversationFormat.ts`, and `editor/src/questFormat.ts`. Schemas live in `schema/`.

**Authentication:** Not detected. WebSocket clients self-identify player entity id through the `join` command, and `server/websocket.go` allows all origins.

**Serialization:** Server components serialize via `SerializeableComponent`; outgoing messages wrap payloads in `{metadata, data}` through `server/message/message.go`. Editor serialization uses `serializeJson` through format helpers.

**IDs:** Entity ids use UUID-backed `model.EntityId` in the server for runtime entities; authored content ids and event tokens are string ids validated by format helpers and schema. Game event tokens are normalized by `server/game/gameevent/event.go`.

---

*Architecture analysis: 2026-06-27*
