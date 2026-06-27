# Coding Conventions

**Analysis Date:** 2026-06-27

## Naming Patterns

**Files:**
- Go files use lowercase package-oriented names, with tests as adjacent `*_test.go` files such as `server/game/world/world_test.go` and `server/game/component/componentinventory_test.go`.
- Server ECS files use descriptive prefixes by role: components in `server/game/component/componentinventory.go`, systems in `server/game/system/systemcombat.go`, and entity constructors in `server/game/entity/entityworldobject.go`.
- Client TypeScript files use camelCase filenames such as `client/game/referenceGeometry.ts`, `client/events/interactionMenu.ts`, and `client/ui/components/entityHealthBar.tsx`.
- React components use `.tsx` in `client/ui/components/`; CSS Modules live beside components as `*.module.css`.
- Editor source files are camelCase TypeScript modules under `editor/src/`, including `editor/src/worldFormat.ts`, `editor/src/gameProject.ts`, and `editor/src/fileSystem.ts`.

**Functions:**
- Go exported functions and methods use PascalCase (`LoadFromGameFS`, `HandleJoin`, `Serialize`); unexported helpers use camelCase (`loadWorldFromBytes`, `makeBlockerGrid`, `entityPosition`).
- Go constructors usually use `New...` (`NewCInventory`, `NewWorld`, `NewWsServer`) and return pointers for mutable runtime objects.
- ECS systems expose `Update()` and keep private helper methods for phases of work, as in `server/game/system/systemcombat.go`.
- TypeScript functions use camelCase (`normalizeWorld`, `validateWorld`, `openGameProjectFolder`, `sendMessage`).
- React event handlers use `handle...` names inside components (`handleOpen`, `handleSave` in `editor/src/App.tsx`).

**Variables:**
- Go local variables use short camelCase names; package constants use PascalCase when exported (`ComponentIdInventory`, `InventoryCapacity`) and lower camelCase when private (`combatTextTtlTicks`).
- TypeScript locals and state values use camelCase. State setters follow React `set...` naming in `editor/src/App.tsx`.
- TypeScript constants use `UPPER_SNAKE_CASE` for fixed values in UI/rendering modules (`SERVER_TICK_SECONDS`, `HUMAN_MODEL_URL`, `INVENTORY_SLOT_COUNT`) and lower camelCase for object maps (`itemIcons`, `itemIconSrcCache`).
- Private TypeScript class fields use the `private` keyword, not underscore prefixes, as in `client/game/renderer/rendererHuman.ts`.

**Types:**
- Go component structs use the `C...` prefix (`CInventory`, `CCombatState`) and implement `Component` or `SerializeableComponent` from `server/game/component/component.go`.
- Go component id constants use the `ComponentId...` prefix with lowercase wire identifiers (`ComponentIdInventory = ComponentId("inventory")`).
- Go systems use `...System` names (`CombatSystem`, `PathingSystem`) and embed or initialize `SystemBase`.
- TypeScript type aliases use PascalCase (`WorldFormat`, `ProjectDirectoryHandle`, `InventoryItem`). The codebase currently favors `type` aliases over `interface`.
- String-union types are used for closed UI/domain sets such as `Tool`, `EditorTab`, `ItemIconKind`, and selection variants in `editor/src/App.tsx` and `client/ui/components/inventory.tsx`.

## Code Style

**Formatting:**
- Go must be formatted with `gofmt`. Keep imports grouped and sorted by the Go toolchain.
- TypeScript is compiled by `tsc --noEmit` through `client/package.json` and `editor/package.json`; there is no checked-in Prettier config.
- Existing TypeScript uses two-space indentation, double quotes, semicolons, and trailing commas in multiline calls/objects.
- Vite config files use ESM JavaScript with double quotes and compact `defineConfig` exports in `client/vite.config.js` and `editor/vite.config.js`.
- Do not hand-edit generated output in `client/dist` or `editor/dist`.

**Linting:**
- No ESLint configuration is present in the current repo.
- TypeScript strictness is enforced through `client/tsconfig.json` and `editor/tsconfig.json`: `strict`, `noUnusedLocals`, `noUnusedParameters`, and `noFallthroughCasesInSwitch` are enabled.
- The practical lint command for frontend code is the relevant build: `cd client && npm run build` or `cd editor && npm run build`.

## Import Organization

**Order:**
1. Go standard library imports (`encoding/json`, `fmt`, `testing`).
2. Internal Go module imports under `webscape/...`.
3. External Go dependencies such as `github.com/google/uuid` and `github.com/gorilla/websocket`.
4. TypeScript local imports and external imports are not strictly separated in all files, so match nearby style when editing.

**Grouping:**
- Go import grouping is left to `gofmt`.
- TypeScript generally places imports at the top with named imports grouped by source; type-only imports are included with value imports using `type` specifiers, as in `editor/src/App.tsx` and `editor/src/fileSystem.ts`.
- Preserve explicit `.ts` extensions in editor relative imports such as `./gameProject.ts` and `./worldFormat.ts`.

**Path Aliases:**
- No TypeScript path aliases are configured in `client/tsconfig.json` or `editor/tsconfig.json`.
- Use relative imports in both frontends.
- Go imports use the module path `webscape/...` from `go.mod`.

## Error Handling

**Patterns:**
- Go loaders and serializers return `error` with contextual wrapping via `fmt.Errorf("context: %w", err)`, as in `server/game/world/world.go`.
- Go command handlers and WebSocket handlers log malformed client input and return rather than panicking, as in `server/commandhandler.go` and `server/websocket.go`.
- Go runtime code often uses guard clauses for nil or invalid state (`if targetHealth == nil { ... continue }`) in systems such as `server/game/system/systemcombat.go`.
- TypeScript editor normalization/validation functions throw `Error` with user-facing messages for invalid documents (`editor/src/worldFormat.ts`, `editor/src/fileSystem.ts`).
- TypeScript UI handlers catch unknown errors and convert them to status text with `errorMessage(...)` in `editor/src/App.tsx`.

**Error Types:**
- Custom error classes are not currently used.
- For expected validation failures, return structured validation results in editor format helpers (`ValidationResult` from `editor/src/formatUtils.ts`) and aggregate error strings.
- For Go content-loading changes, return errors from loader/validator boundaries rather than logging internally; tests assert error/no-error behavior.
- For client transport/rendering failures, existing code uses `console.error` and graceful fallback where possible, such as `client/ws.ts` and model loading in `client/game/renderer/rendererHuman.ts`.

## Logging

**Framework:**
- Go uses the standard `log` package in server boundaries such as `server/commandhandler.go` and `server/websocket.go`.
- TypeScript uses `console.error` for browser-side transport and rendering errors in `client/ws.ts` and `client/game/renderer/rendererHuman.ts`.
- No structured logging library is configured.

**Patterns:**
- Log at network/client-command boundaries, not in low-level pure helpers.
- Include enough context to diagnose malformed input (`Invalid UUID`, command type/data).
- Keep game/content validation functions deterministic and testable by returning errors instead of logging.

## Comments

**When to Comment:**
- Use comments sparingly for non-obvious behavior or public boundary behavior, as in `server/websocket.go` comments for `Broadcast` and `SendToClient`.
- Existing comments are short and explanatory; avoid repeating obvious code.
- Comments are appropriate when returning defensive copies or documenting development-only tradeoffs, such as `GetAllItems` in `server/game/component/componentinventory.go` and `CheckOrigin` in `server/websocket.go`.

**JSDoc/TSDoc:**
- No JSDoc/TSDoc requirement is established.
- Prefer clear exported type/function names over broad documentation blocks unless adding a public editor/client API with non-obvious constraints.

**TODO Comments:**
- No established TODO ownership pattern is present.
- If adding TODOs, include an actionable reason and preferably a tracking issue or phase reference.

## Function Design

**Size:**
- Keep Go logic split by behavior: public command/update entry points delegate to small helpers (`HandleCommand` to `handle...Command`, `Update` to private system methods).
- TypeScript editor functions can be larger in `editor/src/App.tsx`; when adding new behavior, prefer extracting pure format/path/validation logic to modules such as `editor/src/worldFormat.ts` or `editor/src/gameProject.ts`.
- Renderer classes can hold stateful Three.js behavior; keep asset loading, normalization, animation, and update helpers separated as in `client/game/renderer/rendererHuman.ts`.

**Parameters:**
- Go code accepts explicit parameters for domain primitives and uses structs for bundled data (`PathingTarget`, `WorldEntity`, `WorldWall`).
- TypeScript uses object types for component props and options (`WebSocketClientOptions`, React `Props`) and explicit domain types for JSON formats.
- Avoid passing raw `any` through new TypeScript code where a domain type can be defined; current `any` usage is mostly at dynamic WebSocket/JSON boundaries in `client/ws.ts`.

**Return Values:**
- Go getters often return defensive copies for slices (`GetTerrain`, `GetEntities`, `GetAllItems`).
- Go mutators that can fail without exceptional control flow return `bool`, as in inventory methods in `server/game/component/componentinventory.go`.
- Editor validators return `{ valid, errors }` and normalizers throw only after validation fails.
- React handlers update local status/dirty state rather than returning values.

## Module Design

**Exports:**
- Go exports cross-package constructors, ids, model types, and server entry points; unexport helpers that are local implementation details.
- TypeScript uses named exports for format helpers and types in `editor/src/*Format.ts`.
- React components in `client/ui/components/` generally default-export the component.
- Renderer implementations default-export renderer classes such as `RendererHuman`.

**Barrel Files:**
- No broad barrel-file convention is present.
- Import directly from the module that owns the type/function.

---

*Convention analysis: 2026-06-27*
*Update when patterns change*
