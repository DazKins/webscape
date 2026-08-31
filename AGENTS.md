# Repository Guidelines

## Project Structure & Module Organization

Webscape is now a content-driven Go game server with two Vite/React/TypeScript frontends. The root entry point is `main.go`; it loads `config.json`, serves the configured client folder, loads authored game content from the configured game folder, and starts the HTTP/WebSocket server at the configured address.

- `server/`: Go runtime. HTTP/WebSocket setup lives in `server`, client commands in `server/command`, outgoing messages in `server/message`, ECS game state in `server/game`, components in `server/game/component`, systems in `server/game/system`, authored entity conversion in `server/game/entity`, content loading in `server/game/world`, model types in `server/game/model`, math in `server/math`, and helpers in `server/util`.
- `client/`: playable browser client. Startup is `client/main.ts`, WebSocket transport is `client/ws.ts`, commands are in `client/command`, Three.js game/rendering code is under `client/game`, UI events are in `client/events`, and React HUD components are in `client/ui/components`.
- `editor/`: standalone browser-based game project editor. The main UI is `editor/src/App.tsx`; project/file APIs are in `editor/src/fileSystem.ts`; JSON format helpers are in `editor/src/gameProject.ts`, `worldFormat.ts`, `conversationFormat.ts`, and `questFormat.ts`.
- `game-project/`: checked-in sample/authored game content selected by `config.json` and copied into the production Docker image.
- `schema/`: OpenAPI 3.1 schemas and examples for `game.json`, world chunks, conversations, and quests. Keep schemas, editor validators, server loaders, and sample content in sync when changing a format.

Generated output in `client/dist` and `editor/dist` must not be edited by hand.

## Build, Test, and Development Commands

- `cd client && pnpm install --frozen-lockfile`: install playable client dependencies from `client/pnpm-lock.yaml`.
- `cd editor && pnpm install --frozen-lockfile`: install editor dependencies from `editor/pnpm-lock.yaml`.
- `cd client && pnpm run dev`: run the playable client Vite dev server. The Go server still serves `/ws` on `8080`; check wiring before relying on Vite alone.
- `cd editor && pnpm run dev`: run the editor Vite dev server.
- `cd client && pnpm run build`: type-check the playable client and build `client/dist`.
- `cd editor && pnpm run build`: type-check the editor and build `editor/dist`.
- `go run .`: load `config.json` and start the Go server; build `client/dist` first.
- `go build ./...`: compile all Go packages.
- `go test ./...`: run Go tests.
- `docker build -t webscape .`: build the multi-stage production image. The Dockerfile builds the playable client, compiles the Go server, and copies `config.json`, `client/dist`, and `game-project/` into the final image.

Use Node `v24.11.0` for the client when following `client/.nvmrc`; the editor currently has no separate `.nvmrc`.

## Runtime And Content Contracts

The server is authoritative. All live gameplay communication uses `/ws` and falls into three architectural categories. Choose the category from its semantics, not from whichever implementation is most convenient.

### 1. Client-to-server commands

Commands express player intent. Their wire shape is `{ "type": <commandType>, "data": <payload> }`. The client may request an action, but it never sends authoritative state or claims that an outcome occurred. The server validates the connection, visibility, payload, and game rules before mutating ECS state or emitting domain events.

When adding a command, update the command type and decoding in `server/command/command.go`, route it in `server/commandhandler.go`, and update the call site under `client/command` or `client/game`. Commands are not state replication and should not be reused as client acknowledgements or server notifications.

### 2. Server-to-client state replication

State replication communicates authoritative facts that remain true until superseded. A reconnecting or newly interested client needs the current value, not every transition that produced it. Examples include positions, health, inventory, equipment, quest progress, open/depleted state, metadata, and renderable identity.

The server runs systems on a 500 ms tick and then synchronizes each client's interest set in `Game.syncClient`. `world` supplies static registries and world metadata, `chunkUpdate` loads and unloads terrain for that client's current interest, and `gameUpdate` carries per-client ECS component deltas. Only `SerializeableComponent` values visible to that client are considered; unchanged values are omitted, removed components use `data: null`, and no `gameUpdate` is sent when there is no delta. These messages are per-client even when several clients happen to receive equivalent data.

Use a component/state update when late join, reconnect, or interest re-entry must reconstruct the current truth. Do not create short-lived entities, TTL components, or synthetic component changes merely to make the client react once.

### 3. Server-to-client stateless events

Stateless events communicate that something happened once. Examples currently projected to the wire are `chatMessage` and `combatResolved`. They are not authoritative snapshots, are not replayed to late joiners, and may be lost across disconnect/reconnect. The client reacts to them and owns presentation lifetime such as chat bubbles, hit splats, and animations. Ongoing actions such as fishing and woodcutting are replicated state so late joiners can reconstruct their current phase.

Occurrence-driven gameplay begins as a typed domain event in `server/game/gameevent`. Emit it once from authoritative command or system logic. Registered dispatcher subscribers independently handle server concerns such as quest progression and safe client projection. The client projection must choose recipients using server-side interest/visibility, translate the internal payload into an explicit DTO in `server/message`, and must never serialize an internal event wholesale. Projected events are queued and flushed after state synchronization so a client sees the tick's authoritative state before reacting to its events.

New event ids are stable contracts. Subscribe quest handling only to ids referenced by authored quest definitions, and subscribe wire projection only to explicitly supported client events. Do not put client-only presentation fields into ECS state. Existing `conversation` and `questCompleted` messages are also targeted, non-state notifications, though they currently use specialized direct flows rather than the general client-event projector.

All server-to-client messages use `{ "metadata": { "type": ..., "time": ... }, "data": ... }`. Keep wire message types and DTOs in `server/message`, dispatch them in `client/main.ts`, and keep domain-event payloads separate from transport DTOs. Session-control messages such as `registered` and `registrationFailed` support the connection lifecycle; they do not introduce a fourth gameplay communication model.

Startup and deployment settings live in root `config.json`; add runtime settings there instead of introducing command-line flags. Game content starts at project format v2 `game.json`, which defines a uniform `world.chunkSize` and lists chunk, conversation, and quest files. The runtime loads every sparse chunk in `files.chunks`; authored positions are chunk-local and become global server coordinates. Missing chunks are impassable. Project file paths must be relative, slash-separated, and valid for `fs.ValidPath`; do not introduce absolute paths or `..` traversal.

Chunk tile arrays are row-major: index `y * chunkSize.x + x`. Chunk entities are authored as JSON component bags, then converted by `server/game/entity/CreateAuthoredEntity`. Authored entities need a local `position` component; `metadata.blocksMovement` contributes to object blockers, and their `metadata.width`/`height` footprints must remain inside the source chunk. Exactly one `playerSpawn` is required across the project and is not loaded as a normal entity.

Conversations and quests are registry-driven. Entities with a `conversation.conversationId` must reference a loaded conversation. Conversation node events use ids such as `conversation:node:<conversationId>:<nodeId>`, and quests advance from generic game events. Keep event ids stable across content, tests, and UI.

The editor uses the browser File System Access API (`showDirectoryPicker`) to open and save project folders. Preserve its normalization and validation behavior when changing schemas; editor-side validation should reject the same practical invalid content that the Go loader rejects.

## Coding Style & Naming Conventions

Use `gofmt` for Go files and keep package names short and lowercase. Respect the ECS style in `server/AGENTS.md`: component ids are lowercase wire identifiers, component structs use `C...`, component constants use `ComponentId...`, systems use `...System`, and system logic runs from `Update`.

TypeScript is strict in both frontends (`noUnusedLocals`, `noUnusedParameters`, `noFallthroughCasesInSwitch`). Use camelCase for functions, variables, and files. React components use `.tsx`; CSS Modules use `*.module.css` beside the component. Prefer the existing Three.js renderer pattern in `client/game/renderer` when adding renderable entity types.

## Testing Guidelines

There is no dedicated frontend test runner configured. For frontend changes, run the relevant build at minimum: `cd client && pnpm run build`, `cd editor && pnpm run build`, or both.

For Go changes, add focused `_test.go` files near the package under test and run `go test ./...`. Existing coverage focuses on content loading, world validation, conversations, quests, loot, and combat event emission. Prioritize new tests for game systems, component serialization, authored content parsing, command handling, event subscriber routing, client interest filtering, state-before-event ordering, WebSocket message payloads, and schema/runtime compatibility.

For content or schema changes, validate by running `go test ./...` and then `go run .` after rebuilding `client/dist`. If the editor format helpers changed, also run `cd editor && pnpm run build`.

## Commit & Pull Request Guidelines

Recent commits use short, imperative, lowercase subjects such as `add baseline browser mapping dependency` and `components and interactions refactor`. Keep commits focused on one behavioral change. Pull requests should include a concise summary, validation commands, linked issues when applicable, and screenshots or recordings for UI/editor changes.

## Agent-Specific Instructions

Respect the server ECS design in `server/AGENTS.md`. Avoid broad refactors unless they are necessary for the requested change. When touching JSON formats, update all four surfaces together when applicable: `schema/`, `editor/src/*Format.ts`, `server/game/world`, and `game-project/` examples/content. This development environment may require proxy environment variables for network access; when changing proxy settings, scope the change to the smallest command possible.
