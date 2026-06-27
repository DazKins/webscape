# External Integrations

**Analysis Date:** 2026-06-27

## APIs & External Services

**Payment Processing:**
- None currently.

**Email/SMS:**
- None currently.

**External APIs:**
- None currently. Application code does not call remote HTTP APIs via `fetch`, Go HTTP clients, SDK clients, or provider-specific packages.
- npm registry - Build-time dependency source for `client/` and `editor/`, recorded in `client/package-lock.json` and `editor/package-lock.json`; this is not an application runtime integration.
- Go module proxy/source hosting - Build-time module source for `github.com/google/uuid` and `github.com/gorilla/websocket` through `go.mod` and `go.sum`; this is not an application runtime integration.

**Browser Platform APIs:**
- WebSocket API - Playable client connects to the same-origin server endpoint `/ws` from `client/ws.ts`.
  - SDK/Client: Native browser `WebSocket`.
  - Auth: No token or cookie authentication in current code; `client/main.ts` sends a persisted `myPlayerId` from `window.localStorage` in the `join` command after connection.
  - Endpoints used: `ws://<host>/ws` or `wss://<host>/ws` depending on page protocol.
- File System Access API - Editor opens and saves local game project folders from `editor/src/fileSystem.ts`.
  - SDK/Client: Native `window.showDirectoryPicker`.
  - Auth: Browser permission prompt for local folder access.
  - Files used: `game.json`, map files, conversation files, and quest files listed by the project manifest.
- Web Storage API - Playable client persists `myPlayerId` in `window.localStorage` from `client/main.ts`.
  - Auth: None.
  - Data stored: Client-generated UUID only.

## Data Storage

**Databases:**
- None currently. Runtime state is in-memory Go ECS state under `server/game/`, and authored content is loaded from JSON files.

**File Storage:**
- Local game project folder - Server loads authored content from the `-game-folder` path in `main.go` and `server/game/world/world.go`.
  - SDK/Client: Go `os.DirFS`, `fs.ReadFile`, and `fs.ValidPath`.
  - Auth: Operating-system filesystem permissions.
  - Files: `game.json`, maps, conversations, and quests.
- Embedded static assets - Production embeds `client/dist` through `//go:embed all:client/dist` in `main.go`.
  - SDK/Client: Go `embed.FS` and `http.FileServer`.
  - Auth: None.
- Editor local folder writes - Browser writes project JSON through `FileSystemFileHandle.createWritable()` in `editor/src/fileSystem.ts`.
  - SDK/Client: File System Access API.
  - Auth: Browser-mediated local folder permission.

**Caching:**
- None currently. No Redis, CDN client SDK, service worker cache, or explicit application cache integration was found.

## Authentication & Identity

**Auth Provider:**
- None currently. There is no external auth provider, JWT validation, session middleware, OAuth client, or user database.

**OAuth Integrations:**
- None currently.

**Local Identity:**
- Client-generated player id - `client/main.ts` creates `crypto.randomUUID()`, stores it in `window.localStorage`, and sends it in the `join` command.
  - Implementation: Browser `crypto.randomUUID()` and `localStorage`.
  - Token storage: Not a security token; persisted as `myPlayerId`.
  - Session management: Server-side WebSocket connection ids are generated with `github.com/google/uuid` in `server/websocket.go`.

## Monitoring & Observability

**Error Tracking:**
- None currently. No Sentry, Datadog, OpenTelemetry, or hosted error-tracking SDK appears in manifests or code.

**Analytics:**
- None currently.

**Logs:**
- stdout/stderr logging only.
  - Integration: Go `log` package in `main.go`, `server/server.go`, and `server/websocket.go`; browser `console.*` in client transport and startup code.
  - Hosting/runtime log collection depends on the environment running the binary or Docker container.

## CI/CD & Deployment

**Hosting:**
- Docker container - `Dockerfile` defines the production build and runtime image.
  - Deployment: Manual image build is documented with `docker build -t webscape .`; no deployment provider config is present.
  - Environment vars: None required by current application code.
  - Runtime command: `./main -game-folder /app/game-project`.

**CI Pipeline:**
- None currently. No `.github/workflows` directory was found in this mapper pass.
  - Workflows: None.
  - Secrets: None detected in repository config. Forbidden secret-like files were not read; matching files were not present in the scanned repo depth.

## Environment Configuration

**Development:**
- Required env vars: None currently.
- Required flags: `-game-folder game-project` for the Go server; optional `-dev` switches static file serving from embedded FS to live `client/dist`.
- Secrets location: No repository secret files detected; do not introduce secrets into tracked files.
- Mock/stub services: Not applicable; there are no external runtime services to mock.

**Staging:**
- No staging-specific configuration is present.
- Future staging work should define deployment, content source, origin/CORS policy, and log collection explicitly before adding external services.

**Production:**
- Secrets management: None required by current application code.
- Data: JSON content is copied into the Docker image from `game-project/`; updates require rebuilding or changing how `-game-folder` is provided.
- Failover/redundancy: None configured in repo; the Go server holds active game state in process memory.

## Webhooks & Callbacks

**Incoming:**
- None currently. The server exposes HTTP static file serving and WebSocket endpoint `/ws`, but no third-party webhook routes.

**Outgoing:**
- None currently.

## Internal Network Contracts

**Playable Client to Server:**
- Same-origin WebSocket `/ws` - `client/ws.ts` opens the connection and `server/server.go` registers the handler.
  - Protocol: JSON text messages.
  - Incoming client commands: Defined in `server/command/command.go` and emitted from `client/command/command.ts`.
  - Outgoing server messages: Defined under `server/message/` and consumed in `client/main.ts`.
  - Origin policy: `server/websocket.go` currently returns `true` from `CheckOrigin`; tighten this before exposing a public multi-origin deployment.

**Editor to Local Project Folder:**
- Browser-local file IO - `editor/src/fileSystem.ts` reads/writes the same JSON formats loaded by the Go runtime.
  - Contract files: `schema/game-format.openapi.yaml`, `schema/world-format.openapi.yaml`, `schema/conversation-format.openapi.yaml`, and `schema/quest-format.openapi.yaml`.
  - Runtime loader: `server/game/world/world.go`, `server/game/world/conversation.go`, and `server/game/world/quest.go`.
  - Planning rule: When changing a content format, update schemas, editor format helpers, server loaders, and `game-project/` examples together.

---

*Integration audit: 2026-06-27*
*Update when adding/removing external services*
