# Repository Guidelines

## Project Structure & Module Organization

Webscape is now a content-driven Go game server with two Vite/React/TypeScript frontends. The root entry point is `main.go`; it embeds or serves `client/dist`, requires a `-game-folder` containing `game.json`, loads authored game content, and starts the HTTP/WebSocket server on port `8080`.

- `server/`: Go runtime. HTTP/WebSocket setup lives in `server`, client commands in `server/command`, outgoing messages in `server/message`, ECS game state in `server/game`, components in `server/game/component`, systems in `server/game/system`, authored entity conversion in `server/game/entity`, content loading in `server/game/world`, model types in `server/game/model`, math in `server/math`, and helpers in `server/util`.
- `client/`: playable browser client. Startup is `client/main.ts`, WebSocket transport is `client/ws.ts`, commands are in `client/command`, Three.js game/rendering code is under `client/game`, UI events are in `client/events`, and React HUD components are in `client/ui/components`.
- `editor/`: standalone browser-based game project editor. The main UI is `editor/src/App.tsx`; project/file APIs are in `editor/src/fileSystem.ts`; JSON format helpers are in `editor/src/gameProject.ts`, `worldFormat.ts`, `conversationFormat.ts`, and `questFormat.ts`.
- `game-project/`: checked-in sample/authored game content loaded by `go run . -dev -game-folder game-project` and copied into the production Docker image.
- `schema/`: OpenAPI 3.1 schemas and examples for `game.json`, maps/worlds, conversations, and quests. Keep schemas, editor validators, server loaders, and sample content in sync when changing a format.

Generated output in `client/dist` and `editor/dist` must not be edited by hand.

## Build, Test, and Development Commands

- `cd client && npm ci`: install playable client dependencies from `client/package-lock.json`.
- `cd editor && npm ci`: install editor dependencies from `editor/package-lock.json`.
- `cd client && npm run dev`: run the playable client Vite dev server. The Go server still serves `/ws` on `8080`; check wiring before relying on Vite alone.
- `cd editor && npm run dev`: run the editor Vite dev server.
- `cd client && npm run build`: type-check the playable client and build `client/dist`.
- `cd editor && npm run build`: type-check the editor and build `editor/dist`.
- `go run . -dev -game-folder game-project`: start the Go server with live files from `client/dist`; run the client build first.
- `go run . -game-folder game-project`: start with embedded `client/dist`; this requires `client/dist` to exist at compile time.
- `go build ./...`: compile all Go packages.
- `go test ./...`: run Go tests.
- `docker build -t webscape .`: build the multi-stage production image. The Dockerfile builds only the playable client, compiles the Go server, copies `game-project/`, and runs `./main -game-folder /app/game-project`.

Use Node `v24.11.0` for the client when following `client/.nvmrc`; the editor currently has no separate `.nvmrc`.

## Runtime And Content Contracts

The server is authoritative. Clients send JSON commands over `/ws`; the server mutates ECS state and broadcasts typed messages. When adding a command, update both `server/command/command.go` and client command call sites, then route it in `server/commandhandler.go`.

Game content starts at `game.json` and lists map, conversation, and quest files. The runtime currently loads the first map from `files.maps`. Project file paths must be relative, slash-separated, and valid for `fs.ValidPath`; do not introduce absolute paths or `..` traversal.

Map tile arrays are row-major: index `y * size.x + x`. Map entities are authored as JSON component bags, then converted by `server/game/entity/CreateAuthoredEntity`. Authored entities need a `position` component; `metadata.blocksMovement` contributes to object blockers, and `metadata.width`/`height` must stay in bounds. A `playerSpawn` component marks spawn position and is not loaded as a normal entity.

Conversations and quests are registry-driven. Entities with a `conversation.conversationId` must reference a loaded conversation. Conversation node events use ids such as `conversation:node:<conversationId>:<nodeId>`, and quests advance from generic game events. Keep event ids stable across content, tests, and UI.

The editor uses the browser File System Access API (`showDirectoryPicker`) to open and save project folders. Preserve its normalization and validation behavior when changing schemas; editor-side validation should reject the same practical invalid content that the Go loader rejects.

## Coding Style & Naming Conventions

Use `gofmt` for Go files and keep package names short and lowercase. Respect the ECS style in `server/AGENTS.md`: component ids are lowercase wire identifiers, component structs use `C...`, component constants use `ComponentId...`, systems use `...System`, and system logic runs from `Update`.

TypeScript is strict in both frontends (`noUnusedLocals`, `noUnusedParameters`, `noFallthroughCasesInSwitch`). Use camelCase for functions, variables, and files. React components use `.tsx`; CSS Modules use `*.module.css` beside the component. Prefer the existing Three.js renderer pattern in `client/game/renderer` when adding renderable entity types.

## Testing Guidelines

There is no dedicated frontend test runner configured. For frontend changes, run the relevant build at minimum: `cd client && npm run build`, `cd editor && npm run build`, or both.

For Go changes, add focused `_test.go` files near the package under test and run `go test ./...`. Existing coverage focuses on content loading, world validation, conversations, quests, loot, and combat event emission. Prioritize new tests for game systems, component serialization, authored content parsing, command handling, WebSocket message payloads, and schema/runtime compatibility.

For content or schema changes, validate by running `go test ./...` and then `go run . -dev -game-folder game-project` after rebuilding `client/dist`. If the editor format helpers changed, also run `cd editor && npm run build`.

## Commit & Pull Request Guidelines

Recent commits use short, imperative, lowercase subjects such as `add baseline browser mapping dependency` and `components and interactions refactor`. Keep commits focused on one behavioral change. Pull requests should include a concise summary, validation commands, linked issues when applicable, and screenshots or recordings for UI/editor changes.

## Agent-Specific Instructions

Respect the server ECS design in `server/AGENTS.md`. Avoid broad refactors unless they are necessary for the requested change. When touching JSON formats, update all four surfaces together when applicable: `schema/`, `editor/src/*Format.ts`, `server/game/world`, and `game-project/` examples/content. This development environment may require proxy environment variables for network access; when changing proxy settings, scope the change to the smallest command possible.
