# Development

## Local setup

Run commands from the repository root unless stated otherwise. Use Node `v24.11.0` (see `client/.nvmrc`), pnpm, and Go.

Configure an OIDC provider and the `auth` settings in the configuration file you will run (`config.dev.json` in the example below), as described in
[Authentication](authentication.md), including the client-secret environment
variable. For provider-free local testing, configure the explicit [no-auth testing mode](authentication.md#no-auth-testing-mode). Startup fails closed until auth is configured. Then build the playable client before starting the server:

```sh
cd client
pnpm install --frozen-lockfile
pnpm run build
cd ..
WEBSCAPE_CONFIG=config.dev.json go run .
```

This runs a temporary dev server at `http://localhost:8080` with frontend caching
and persistence disabled. Stop it with Ctrl+C. Without `WEBSCAPE_CONFIG`, the
server loads `config.json`.

The editor runs separately:

```sh
cd editor
pnpm install --frozen-lockfile
pnpm run dev
```

## Build and validation

Run Go checks from the repository root:

```sh
go build ./...
go test ./...
```

For frontend changes, run `pnpm run build` in `client/`, `editor/`, or both as applicable. These commands type-check and build the frontends. Generated `dist/` files must not be edited by hand.

See [persistence](persistence.md) for PostgreSQL integration tests and [authentication](authentication.md#validation) for authentication checks.

## Client build identifier
The top-right build label identifies the loaded client assets by the first eight
characters of their Git commit SHA. Hover over it for the full revision. The
identifier is fixed when Vite starts or builds; rebuild and reload the client to
pick up a different revision.

Local builds detect Git HEAD automatically, including in Git worktrees, and append
`-dirty` when the repository has uncommitted changes (including untracked files).
Set `WEBSCAPE_BUILD_REVISION` to an explicit commit SHA to override detection. An
explicit revision is used as supplied, without an automatic dirty suffix. Builds
without Git metadata or an override show `Build unknown`.

Docker excludes Git metadata, so pass the client source revision at build time:

```sh
docker build --build-arg WEBSCAPE_BUILD_REVISION="$(git rev-parse HEAD)" -t webscape .
```

Build deployments from a clean checkout so that this SHA identifies their source.
The label identifies the client build; it does not version a separately deployed
server or game content.
