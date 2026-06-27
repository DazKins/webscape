# Technology Stack

**Analysis Date:** 2026-06-27

## Languages

**Primary:**
- Go 1.25.2 - Authoritative game server, HTTP/WebSocket runtime, ECS game logic, content loading, and production binary in `main.go`, `server/`, and `go.mod`.
- TypeScript 5.9.3 - Playable client and standalone editor source in `client/` and `editor/`.

**Secondary:**
- JavaScript - Vite configuration files in `client/vite.config.js` and `editor/vite.config.js`.
- JSON/YAML - Authored game content in `game-project/` and OpenAPI schemas/examples in `schema/`.
- CSS / CSS Modules - Client HUD and editor styling in `client/ui/components/*.module.css`, `client/ui/uiRoot.module.css`, and `editor/src/App.css`.

## Runtime

**Environment:**
- Go 1.25.2 - Required by `go.mod`; the Docker build uses `golang:1.25.2-alpine`.
- Node.js v24.11.0 - Declared for the playable client in `client/.nvmrc`; Docker uses `node:24-alpine` for the client build.
- Browser runtime - The playable client runs in a browser and connects to `/ws`; the editor runs as a browser app and requires File System Access API support for project folders.
- HTTP port `8080` - `server/server.go` starts `http.ListenAndServe(":8080", nil)`.

**Package Manager:**
- npm - Both frontends use npm scripts and checked-in lockfiles.
- Lockfiles: `client/package-lock.json` and `editor/package-lock.json` are present.
- Go modules - `go.mod` and `go.sum` are present.

## Frameworks

**Core:**
- Go standard library `net/http`, `embed`, and `io/fs` - Static file serving, embedded `client/dist`, live dev filesystem serving, flags, and HTTP server setup in `main.go` and `server/server.go`.
- Gorilla WebSocket v1.5.3 - WebSocket upgrade and message transport in `server/websocket.go`.
- React 19.2 - Client HUD and editor UI in `client/ui/` and `editor/src/`.
- Three.js 0.180.0 - Playable client scene, camera, entity rendering, and GLB loading under `client/game/`.
- Vite 7.x - Dev/build tooling for both browser apps via `client/vite.config.js` and `editor/vite.config.js`.

**Testing:**
- Go `testing` package - Unit tests live beside packages as `_test.go` files under `server/`.
- No dedicated frontend test runner is configured in `client/package.json` or `editor/package.json`; frontend validation currently depends on TypeScript plus Vite builds.

**Build/Dev:**
- TypeScript compiler 5.9.3 - `npm run build` runs `tsc --noEmit` before Vite in both frontends.
- Vite React plugin 5.x - React transform plugin in both Vite configs.
- Docker multi-stage build - `Dockerfile` builds the playable client, compiles the Go server, copies `game-project/`, exposes `8080`, and runs `./main -game-folder /app/game-project`.

## Key Dependencies

**Critical:**
- `github.com/gorilla/websocket` v1.5.3 - Server-side WebSocket transport for `/ws`.
- `github.com/google/uuid` v1.6.0 - Server-generated connection ids in `server/websocket.go`.
- `react` / `react-dom` 19.2 - UI runtime for both browser apps.
- `three` 0.180.0 - 3D rendering and client-side scene composition in the playable client.
- Browser WebSocket API - Client transport in `client/ws.ts`.
- Browser File System Access API - Editor project open/save flow in `editor/src/fileSystem.ts`.

**Infrastructure:**
- Go standard library filesystem APIs - `os.DirFS`, `fs.ReadFile`, `fs.ValidPath`, and embedded FS drive static assets and content loading in `main.go` and `server/game/world/`.
- OpenAPI 3.1 schemas - Content contract documentation and validation references in `schema/*.openapi.yaml`.
- npm registry packages - Build-time dependency source recorded in `client/package-lock.json` and `editor/package-lock.json`.

## Configuration

**Environment:**
- Runtime configuration is via CLI flags, not environment variables: `main.go` accepts `-dev` and required `-game-folder`.
- No application `.env`, `.env.*`, `*.env`, `.npmrc`, or `.netrc` files were found during this mapper pass.
- The server requires a game folder containing `game.json`; content paths are validated with `fs.ValidPath` in `server/game/world/world.go`.

**Build:**
- `go.mod` / `go.sum` - Go module and dependency lock state.
- `client/package.json`, `client/package-lock.json`, `client/tsconfig.json`, `client/vite.config.js`, `client/.nvmrc` - Playable client build and runtime expectations.
- `editor/package.json`, `editor/package-lock.json`, `editor/tsconfig.json`, `editor/vite.config.js` - Editor build expectations.
- `Dockerfile` - Production image build and runtime command.
- `schema/*.openapi.yaml` and `game-project/*.json` - Authored content format contracts and sample runtime content.

## Platform Requirements

**Development:**
- Any platform with Go 1.25.2 and Node.js v24.11.0 for client work.
- Run `cd client && npm ci` and `cd editor && npm ci` before frontend builds.
- Build `client/dist` before `go run . -dev -game-folder game-project` or `go run . -game-folder game-project`; `main.go` embeds or serves `client/dist`.
- Use a browser with WebSocket support for the playable client and File System Access API support for the editor folder workflow.

**Production:**
- Docker target is a Linux Alpine container based on `alpine:3.22.1`.
- Production binary is statically built with `CGO_ENABLED=0 GOOS=linux` in `Dockerfile`.
- The image includes `game-project/`, serves the embedded playable client, exposes port `8080`, and defaults to `./main -game-folder /app/game-project`.
- No external database, cache, object storage, or managed platform SDK is currently required by application code.

---

*Stack analysis: 2026-06-27*
*Update after major dependency changes*
