# Codebase Structure

**Analysis Date:** 2026-06-27

## Directory Layout

```text
webscape/
├── main.go                 # Go binary entry; embeds/serves playable client and starts server
├── go.mod                  # Go module and server dependencies
├── Dockerfile              # Production image build for client and Go server
├── AGENTS.md               # Repository-wide agent instructions
├── server/                 # Go HTTP/WebSocket and authoritative ECS runtime
│   ├── command/            # Incoming client command DTOs
│   ├── game/               # ECS game state, content runtime, systems, models
│   ├── message/            # Outgoing WebSocket message factories
│   ├── math/               # Server math primitives
│   └── util/               # Generic server helpers
├── client/                 # Playable Vite/React/TypeScript/Three.js browser client
│   ├── command/            # Outgoing command helper
│   ├── events/             # Custom UI/game events
│   ├── game/               # Client-side mirror state, camera, world, renderers
│   ├── ui/                 # React HUD root and components
│   ├── public/             # Static client assets
│   └── dist/               # Generated playable client build output
├── editor/                 # Standalone Vite/React/TypeScript game project editor
│   ├── src/                # Editor app, file APIs, and format helpers
│   └── dist/               # Generated editor build output
├── game-project/           # Checked-in sample/authored game content
│   ├── game.json           # Project manifest
│   ├── maps/               # World/map JSON documents
│   ├── conversations/      # Conversation JSON documents
│   └── quests/             # Quest JSON documents
├── schema/                 # OpenAPI 3.1 schemas and example documents
│   └── examples/           # Schema examples
└── .planning/              # GSD planning artifacts
    └── codebase/           # Generated codebase mapping documents
```

## Directory Purposes

**`server/`:**
- Purpose: Go server process and authoritative runtime.
- Contains: HTTP/WebSocket setup, command routing, ECS game loop, game systems, content loading, outgoing messages.
- Key files: `server/server.go`, `server/websocket.go`, `server/commandhandler.go`, `server/AGENTS.md`.
- Subdirectories: `server/command/`, `server/game/`, `server/message/`, `server/math/`, `server/util/`.

**`server/command/`:**
- Purpose: Define incoming client command wire shape.
- Contains: Go command type constants and JSON unmarshal helper.
- Key files: `server/command/command.go`.
- Subdirectories: None.

**`server/game/`:**
- Purpose: Own gameplay state and domain behavior.
- Contains: `Game`, tests for conversations/quests/loot/spawn/openable behavior, ECS subpackages.
- Key files: `server/game/game.go`, `server/game/conversation_test.go`, `server/game/quest_test.go`, `server/game/loot_test.go`, `server/game/spawn_test.go`.
- Subdirectories: `component/`, `system/`, `entity/`, `world/`, `model/`, `collision/`, `gameevent/`.

**`server/game/component/`:**
- Purpose: ECS component definitions and component storage.
- Contains: component interfaces, component id constants, serializable component structs, inventory/equipment/combat helpers.
- Key files: `server/game/component/component.go`, `server/game/component/manager.go`, `server/game/component/componentposition.go`, `server/game/component/componentinventory.go`, `server/game/component/componentquestlog.go`.
- Subdirectories: None.

**`server/game/system/`:**
- Purpose: Per-tick ECS behavior.
- Contains: system interface/base and systems for pathing, interaction, combat, random walk, TTL, health, and spawning.
- Key files: `server/game/system/system.go`, `server/game/system/systempathing.go`, `server/game/system/systeminteraction.go`, `server/game/system/systemcombat.go`, `server/game/system/systemspawn.go`.
- Subdirectories: None.

**`server/game/entity/`:**
- Purpose: Entity/component factory helpers.
- Contains: player/chat/world object creation and authored JSON component conversion.
- Key files: `server/game/entity/entityworldobject.go`, `server/game/entity/entityplayer.go`, `server/game/entity/entitychatmessage.go`, `server/game/entity/entitydude.go`.
- Subdirectories: None.

**`server/game/world/`:**
- Purpose: Runtime content formats, validation, and registries.
- Contains: game manifest loader, world/map loader, conversation registry, quest registry.
- Key files: `server/game/world/world.go`, `server/game/world/conversation.go`, `server/game/world/quest.go`, `server/game/world/world_test.go`.
- Subdirectories: None.

**`server/game/model/`:**
- Purpose: Domain model types shared by components and systems.
- Contains: entity ids, item ids, equipment slots, item definitions.
- Key files: `server/game/model/entity.go`, `server/game/model/item.go`, `server/game/model/items.go`.
- Subdirectories: None.

**`server/message/`:**
- Purpose: Outgoing server-to-client WebSocket messages.
- Contains: message envelope, message type constants, and payload factories.
- Key files: `server/message/message.go`, `server/message/type.go`, `server/message/gameupdate.go`, `server/message/world.go`, `server/message/conversation.go`, `server/message/questcompleted.go`.
- Subdirectories: None.

**`client/`:**
- Purpose: Playable browser client source and Vite project.
- Contains: WebSocket client, command helper, Three.js game runtime, React HUD, CSS Modules, static assets.
- Key files: `client/main.ts`, `client/ws.ts`, `client/input.ts`, `client/package.json`, `client/vite.config.js`, `client/tsconfig.json`.
- Subdirectories: `client/game/`, `client/ui/`, `client/events/`, `client/command/`, `client/math/`, `client/util/`, `client/public/`.

**`client/game/`:**
- Purpose: Client-side game shell and rendering model.
- Contains: `Game` event target, local `Entity`, camera, world geometry, entity render system, renderer classes.
- Key files: `client/game/game.ts`, `client/game/entityRenderSystem.ts`, `client/game/camera.ts`, `client/game/world/world.ts`.
- Subdirectories: `client/game/entity/`, `client/game/renderer/`, `client/game/world/`.

**`client/game/renderer/`:**
- Purpose: Three.js renderers for server `renderable.type` values.
- Contains: renderer base classes and concrete renderers for humans, trees, doors, chests, rocks, buildings, reward drops, chat messages, combat text, walls, and errors.
- Key files: `client/game/renderer/renderer.ts`, `client/game/renderer/positionedEntityRenderer.ts`, `client/game/renderer/rendererHuman.ts`, `client/game/renderer/rendererWall.ts`.
- Subdirectories: None.

**`client/ui/`:**
- Purpose: React HUD root and reusable UI components.
- Contains: `UiRoot`, CSS Modules, and panels for chat, combat log, inventory/equipment, quests, interactions, conversations, overlays, and name/health decorations.
- Key files: `client/ui/uiRoot.tsx`, `client/ui/components/chatBox.tsx`, `client/ui/components/inventory.tsx`, `client/ui/components/conversationPanel.tsx`, `client/ui/components/questPanel.tsx`.
- Subdirectories: `client/ui/components/`.

**`client/events/`:**
- Purpose: Typed custom browser events between `client/game/game.ts` and React components.
- Contains: event classes for chat, combat log, conversation, interaction menu, inventory, quest log, and quest completion.
- Key files: `client/events/conversation.ts`, `client/events/inventory.ts`, `client/events/questCompleted.ts`.
- Subdirectories: None.

**`editor/`:**
- Purpose: Standalone game project editor Vite app.
- Contains: React entry, app shell, browser file API wrappers, and JSON format helpers.
- Key files: `editor/src/App.tsx`, `editor/src/main.tsx`, `editor/src/fileSystem.ts`, `editor/package.json`, `editor/vite.config.js`.
- Subdirectories: `editor/src/`, `editor/dist/`.

**`editor/src/`:**
- Purpose: Editor implementation.
- Contains: app state/UI, format normalizers/validators/serializers, file system access helpers, CSS.
- Key files: `editor/src/App.tsx`, `editor/src/fileSystem.ts`, `editor/src/gameProject.ts`, `editor/src/worldFormat.ts`, `editor/src/conversationFormat.ts`, `editor/src/questFormat.ts`, `editor/src/formatUtils.ts`.
- Subdirectories: None.

**`game-project/`:**
- Purpose: Sample/authored content loaded by the Go server and used as format examples.
- Contains: `game.json`, maps, conversations, quests.
- Key files: `game-project/game.json`, `game-project/maps/starter.json`, `game-project/conversations/new_conversation.json`, `game-project/quests/first_errand.json`.
- Subdirectories: `game-project/maps/`, `game-project/conversations/`, `game-project/quests/`.

**`schema/`:**
- Purpose: OpenAPI schemas and examples for authored JSON formats.
- Contains: schema files for game manifest, world/map, conversation, and quest formats.
- Key files: `schema/game-format.openapi.yaml`, `schema/world-format.openapi.yaml`, `schema/conversation-format.openapi.yaml`, `schema/quest-format.openapi.yaml`.
- Subdirectories: `schema/examples/`.

## Key File Locations

**Entry Points:**
- `main.go`: Go binary entry point and content/static asset bootstrap.
- `server/server.go`: HTTP server and game/WebSocket wiring.
- `server/websocket.go`: WebSocket connection entry point at `/ws`.
- `client/main.ts`: playable client browser entry and animation loop.
- `editor/src/main.tsx`: editor browser entry.
- `editor/src/App.tsx`: editor app shell and document state owner.

**Configuration:**
- `go.mod`: Go module, Go version, and server dependencies.
- `client/package.json`: playable client scripts and dependencies.
- `client/tsconfig.json`: playable client TypeScript strictness.
- `client/vite.config.js`: playable client Vite config.
- `editor/package.json`: editor scripts and dependencies.
- `editor/tsconfig.json`: editor TypeScript strictness.
- `editor/vite.config.js`: editor Vite config.
- `Dockerfile`: production build and runtime image.
- `AGENTS.md`: repo-wide development instructions.
- `server/AGENTS.md`: server ECS-specific instructions.

**Core Server Logic:**
- `server/commandhandler.go`: incoming command dispatch.
- `server/command/command.go`: incoming command wire type definitions.
- `server/game/game.go`: authoritative game runtime, systems registration, update loop, commands, quests, conversations, inventory, equipment.
- `server/game/component/manager.go`: component storage and entity component operations.
- `server/game/component/component.go`: component interfaces.
- `server/game/system/system.go`: system interface.
- `server/game/world/world.go`: game manifest and world/map loading/validation.
- `server/game/entity/entityworldobject.go`: authored entity component conversion.
- `server/message/message.go`: outgoing message envelope and JSON marshaling.
- `server/message/gameupdate.go`: component delta message creation.

**Core Client Logic:**
- `client/ws.ts`: browser WebSocket wrapper and reconnect behavior.
- `client/command/command.ts`: outgoing command factory.
- `client/game/game.ts`: local game state, input commands, message handlers, UI event emission.
- `client/game/entity/entity.ts`: local client entity component bag.
- `client/game/entityRenderSystem.ts`: renderer lifecycle and renderable type dispatch.
- `client/game/world/world.ts`: map geometry, terrain, wall geometry hookup, hovered tile detection.
- `client/input.ts`: keyboard/mouse tracking and active receiver.
- `client/ui/uiRoot.tsx`: HUD composition.

**Core Editor Logic:**
- `editor/src/App.tsx`: project/editor UI state and editing operations.
- `editor/src/fileSystem.ts`: File System Access API open/save/delete flows.
- `editor/src/gameProject.ts`: `game.json` format helper.
- `editor/src/worldFormat.ts`: world/map format helper.
- `editor/src/conversationFormat.ts`: conversation format helper.
- `editor/src/questFormat.ts`: quest format helper.
- `editor/src/formatUtils.ts`: common format utilities.

**Content and Schema:**
- `game-project/game.json`: checked-in game manifest.
- `game-project/maps/starter.json`: checked-in map content.
- `game-project/conversations/new_conversation.json`: checked-in conversation content.
- `game-project/quests/first_errand.json`: checked-in quest content.
- `schema/game-format.openapi.yaml`: game manifest schema.
- `schema/world-format.openapi.yaml`: world/map schema.
- `schema/conversation-format.openapi.yaml`: conversation schema.
- `schema/quest-format.openapi.yaml`: quest schema.
- `schema/examples/game.json`: game manifest example.
- `schema/examples/starter-conversations.json`: conversation example.
- `schema/examples/tutorial-quests.json`: quest example.

**Testing:**
- `server/game/world/world_test.go`: world/content loader validation tests.
- `server/game/conversation_test.go`: conversation behavior tests.
- `server/game/quest_test.go`: quest behavior tests.
- `server/game/loot_test.go`: loot behavior tests.
- `server/game/openable_test.go`: openable behavior tests.
- `server/game/spawn_test.go`: spawn behavior tests.
- `server/game/component/componentinventory_test.go`: inventory component tests.
- `server/game/entity/entityworldobject_test.go`: authored entity conversion tests.
- `server/game/system/systempathing_test.go`: pathing system tests.
- `server/game/system/systemrandomwalk_test.go`: random walk system tests.
- `server/game/system/combat_event_test.go`: combat event tests.
- `server/message/world_test.go`: world message tests.

## Naming Conventions

**Files:**
- Lowercase Go package files: `server/game/system/systemcombat.go`, `server/game/component/componenthealth.go`.
- Go tests colocated with packages: `server/game/world/world_test.go`, `server/game/system/systempathing_test.go`.
- TypeScript modules use camelCase/lowercase filenames: `client/game/entityRenderSystem.ts`, `editor/src/gameProject.ts`.
- React components use `.tsx` with camelCase filenames: `client/ui/components/conversationPanel.tsx`, `editor/src/App.tsx`.
- CSS Modules sit beside client React components as `*.module.css`: `client/ui/components/chatBox.module.css`.
- JSON content filenames are lowercase with underscores where needed: `game-project/conversations/new_conversation.json`.
- Schemas use kebab-case OpenAPI filenames: `schema/world-format.openapi.yaml`.
- Generated build directories are named `dist`: `client/dist`, `editor/dist`.

**Directories:**
- Server packages are short lowercase nouns: `server/message`, `server/util`, `server/game/world`.
- Server ECS subdirectories are domain layers: `server/game/component`, `server/game/system`, `server/game/entity`.
- Client feature directories are lowercase nouns: `client/events`, `client/game`, `client/ui`.
- Content directories are plural collections: `game-project/maps`, `game-project/conversations`, `game-project/quests`.
- Schema examples live under `schema/examples`.

**Special Patterns:**
- Server component structs use `C...`, component ids use lowercase wire strings, and constants use `ComponentId...` in `server/game/component/`.
- Server systems use `...System` structs and implement `Update()` in `server/game/system/`.
- Outgoing message factories use `New...Message` in `server/message/`.
- Client event classes use `...Event` in `client/events/`.
- Client renderer classes use `Renderer...` and are selected in `client/game/entityRenderSystem.ts`.
- Editor format helpers export `normalize...`, `validate...`, `serialize...`, and `createBlank...` functions in `editor/src/*Format.ts`.

## Where to Add New Code

**New Server Command:**
- Definition: `server/command/command.go`
- Handler: `server/commandhandler.go`
- Game behavior: `server/game/game.go` or a focused system under `server/game/system/`
- Client caller: `client/command/command.ts` and command call site in `client/game/game.ts` or `client/ui/components/`
- Tests: colocate under `server/` package touched, usually `server/game/` or `server/game/system/`

**New Outgoing Message:**
- Type constant: `server/message/type.go`
- Message payload/factory: new file under `server/message/`
- Server sender: `server/game/game.go` or relevant server package using `g.sendMessage`/`g.broadcastMessage`
- Client handler: `client/main.ts`
- Client state/UI event: `client/game/game.ts` and optionally `client/events/`
- Tests: `server/message/*_test.go` for serialization shape when needed.

**New ECS Component:**
- Component implementation: new `server/game/component/component{name}.go`
- Component id constant: same component file or related component file under `server/game/component/`
- Serialization: implement `Serialize()` only if the client needs it.
- Authored content conversion: `server/game/entity/entityworldobject.go` if the component can come from map JSON.
- Client consumption: `client/game/game.ts`, `client/game/entity/entity.ts`, UI, or renderer code depending on visibility.
- Tests: `server/game/component/` for component behavior and `server/game/entity/entityworldobject_test.go` for authored conversion.

**New ECS System:**
- Implementation: `server/game/system/system{name}.go`
- Registration order: `server/game/game.go` in `NewGameWithWorld`
- Tests: `server/game/system/system{name}_test.go`
- Shared collision/world helpers: use `server/game/collision/` or `server/game/world/` rather than duplicating grid logic.

**New Authored Entity Type:**
- Runtime conversion: extend `server/game/entity/entityworldobject.go`
- Renderable string support: `client/game/entityRenderSystem.ts`
- Renderer implementation: `client/game/renderer/renderer{Name}.ts`
- Content example: `game-project/maps/`
- Schema/editor support: `schema/world-format.openapi.yaml` and `editor/src/worldFormat.ts` if the format surface changes.

**New Content Format Field:**
- Server loader/validator: `server/game/world/`
- Editor normalizer/validator/serializer: `editor/src/gameProject.ts`, `editor/src/worldFormat.ts`, `editor/src/conversationFormat.ts`, or `editor/src/questFormat.ts`
- Schema: matching file under `schema/`
- Sample content: `game-project/` and relevant `schema/examples/`
- Tests: `server/game/world/world_test.go` or focused loader tests; editor build at minimum.

**New Client UI Panel or HUD Component:**
- Component: `client/ui/components/{name}.tsx`
- Styles: `client/ui/components/{name}.module.css`
- Root placement: `client/ui/uiRoot.tsx`
- Game events/state: `client/events/` and `client/game/game.ts`
- Commands: use `client/command/command.ts` through methods on `client/game/game.ts`.

**New Client Renderer:**
- Renderer class: `client/game/renderer/renderer{Name}.ts`
- Base class: extend existing patterns in `client/game/renderer/renderer.ts` or `client/game/renderer/positionedEntityRenderer.ts`
- Registration: add `renderable.type` branch in `client/game/entityRenderSystem.ts`
- Server/content source: ensure server emits `renderable` with the same type from `server/game/entity/entityworldobject.go` or entity factories.

**New Editor Capability:**
- UI and state: `editor/src/App.tsx`
- File persistence: `editor/src/fileSystem.ts`
- Format logic: relevant `editor/src/*Format.ts`
- Shared validation utilities: `editor/src/formatUtils.ts`
- Runtime parity: update `server/game/world/` and `schema/` when saved JSON changes.

**Utilities:**
- Server helpers: `server/util/` for generic data structures or JSON/path helpers.
- Server math: `server/math/` for coordinate/vector math.
- Client helpers: `client/util/` for browser/client-specific utilities.
- Editor helpers: keep format-specific helpers in `editor/src/*Format.ts`; put cross-format helpers in `editor/src/formatUtils.ts`.

## Special Directories

**`client/dist/`:**
- Purpose: Generated playable client build output served or embedded by the Go server.
- Source: `cd client && npm run build`.
- Generated: Yes.
- Committed: Present in repository structure but must not be edited by hand.

**`editor/dist/`:**
- Purpose: Generated standalone editor build output.
- Source: `cd editor && npm run build`.
- Generated: Yes.
- Committed: Present in repository structure but must not be edited by hand.

**`client/public/`:**
- Purpose: Static assets copied by Vite for the playable client.
- Source: authored assets such as `client/public/models/man.glb`.
- Generated: No.
- Committed: Yes.

**`game-project/`:**
- Purpose: Sample authored game content loaded by development and Docker runtime commands.
- Source: hand-authored/editor-authored JSON.
- Generated: No.
- Committed: Yes.

**`schema/`:**
- Purpose: OpenAPI schemas and examples for authored content contracts.
- Source: hand-maintained contract files.
- Generated: No.
- Committed: Yes.

**`.planning/codebase/`:**
- Purpose: GSD codebase mapping documents for future planning and execution.
- Source: mapper-generated Markdown.
- Generated: Yes.
- Committed: Orchestrator decides; mapper must not commit.

---

*Structure analysis: 2026-06-27*
