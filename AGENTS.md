# Repository Guidelines

## Project Structure & Module Organization

Webscape is a Go backend with a Vite/React/TypeScript client. The entry point is `main.go`, which serves the built client and starts the WebSocket game server on port `8080`. Backend code lives under `server/`: commands in `server/command`, HTTP/WebSocket setup in `server`, ECS game state in `server/game`, components in `server/game/component`, systems in `server/game/system`, and helpers in `server/util`. Client code lives under `client/`: startup in `client/main.ts`, rendering and world state in `client/game`, events in `client/events`, WebSocket code in `client/ws.ts`, and React UI in `client/ui/components`. Generated output in `client/dist` should not be edited by hand.

## Build, Test, and Development Commands

- `cd client && npm ci`: install client dependencies from `package-lock.json`.
- `cd client && npm run dev`: run the Vite development server.
- `cd client && npm run build`: type-check with `tsc --noEmit` and build production assets.
- `go run . -dev`: start the Go server using live files from `client/dist`; run `npm run build` first.
- `go build ./...`: compile all Go packages.
- `go test ./...`: run Go tests when present.
- `docker build -t webscape .`: build the multi-stage production image.

## Coding Style & Naming Conventions

Use `gofmt` for Go files and keep package names short and lowercase. Follow the ECS naming style: components use `Component...`, systems use `System...`, and entity constructors live in `server/game/entity`. TypeScript is strict (`noUnusedLocals`, `noUnusedParameters`, `noFallthroughCasesInSwitch`); use camelCase for functions, variables, and files. React components use `.tsx`; CSS Modules use `*.module.css` beside the component.

## Testing Guidelines

There is no dedicated frontend test runner configured yet. For client changes, run `cd client && npm run build` at minimum. For Go changes, add focused `_test.go` files near the package under test and run `go test ./...`. Prioritize tests for game systems, component serialization, command parsing, and WebSocket message handling.

## Commit & Pull Request Guidelines

Recent commits use short, imperative, lowercase subjects such as `remove unused tick counter` and `respect combat range in pathing`. Keep commits focused on one behavioral change. Pull requests should include a concise summary, validation commands, linked issues when applicable, and screenshots or recordings for UI changes.

## Agent-Specific Instructions

Respect the server ECS design in `server/AGENTS.md`. Avoid broad refactors unless they are necessary for the requested change. This development environment may require proxy environment variables for network access; when changing proxy settings, scope the change to the smallest command possible.
