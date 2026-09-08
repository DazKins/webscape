# Webscape

![Webscape Screenshot](https://dazkins.com/_astro/screenshot.DwpayBVs_2r7kIj.webp)

Welcome to Webscape!

Webscape is a low tick rate, tile-based browser MMO highly inspired by [Runescape](https://oldschool.runescape.com/).

The game is currently in active development. You can access and play it here: https://webscape.dazkins.com/

## Client/server communication

The Go server is authoritative, and the playable client communicates with it through `/ws`. Gameplay traffic follows three deliberately separate patterns:

| Flow | Purpose | Examples | Delivery semantics |
| --- | --- | --- | --- |
| Client → server commands | Express player intent for the server to validate and execute | `move`, `chat`, `interact`, `equip` | A request, never authoritative state or proof that an outcome occurred |
| Server → client state | Replicate the current authoritative truth | `world`, `chunkUpdate`, `gameUpdate` component deltas | Per-client and interest-filtered; reconnect and interest re-entry reconstruct current state |
| Server → client events | Notify eligible clients that something happened once | `chatMessage`, `combatResolved` | Stateless and not replayed; the client owns presentation and expiry |

Commands use `{ "type": ..., "data": ... }`. Server messages use `{ "metadata": { "type": ..., "time": ... }, "data": ... }`.

On each 500 ms game tick, the server runs its ECS systems and then synchronizes every client's visible chunks and entities. `gameUpdate` contains only changed serializable components; a `null` component value removes it. If a fact must survive reconnect or interest re-entry—position, health, inventory, quest progress, whether a door is open—it belongs in replicated state.

One-shot occurrences use the server's typed domain-event dispatcher instead. Registered subscribers can independently advance quests or project a safe event DTO to interested clients. Client events are sent after that tick's state deltas, so rendering reacts to the latest authoritative state. Chat bubbles, hit splats, and transient combat animations are client-owned effects. Ongoing fishing and woodcutting animations are selected from their replicated phase and phase-start tick.

Internal domain events and WebSocket messages are intentionally different contracts: the server explicitly chooses recipients and maps approved fields into a wire DTO. Events are not replayed after a disconnect. Specialized `conversation` and `questCompleted` notifications are also non-state messages, although they currently use their own targeted flows. Session-control responses such as `registered` and `registrationFailed` handle connection setup rather than gameplay state.

## Development

Build the playable client before starting the server:

```sh
cd client
pnpm install --frozen-lockfile
pnpm run build
cd ..
go run .
```

The editor runs separately from `editor/` with `pnpm run dev`.

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

## Published Docker image

The `Publish Docker image` GitHub Actions workflow builds the root Dockerfile on
pushes to `master`, including merged pull requests, and publishes a Linux AMD64
image to GitHub Container Registry:

- `ghcr.io/dazkins/webscape:latest` tracks the latest successful publication.
- `ghcr.io/dazkins/webscape:sha-<full-commit-sha>` identifies a specific revision.

The workflow passes the source SHA into the client build identifier and image
metadata. It authenticates with the automatic `GITHUB_TOKEN` using
`contents: read` and `packages: write`; no additional registry secret is needed.
Publication does not restart the running game server.

Pull and run a published image with:

```sh
docker pull ghcr.io/dazkins/webscape:latest
docker run --rm -p 8080:8080 ghcr.io/dazkins/webscape:latest
```

For a specific revision, replace `latest` with its `sha-<full-commit-sha>` tag.
Package visibility is managed in GitHub Packages settings; make the package public
if anonymous pulls are required, or authenticate to GHCR before pulling a private
package.

## PR preview images

The `Build PR preview image` workflow builds each new PR head revision when a PR
opens, receives commits, or reopens. Same-repository PRs publish a testing-only
prerelease image at
`ghcr.io/dazkins/webscape:pr-<number>-sha-<full-head-sha>`. Fork and Dependabot PRs
validate the Docker build without publishing an image.

Preview images carry the PR head revision, a testing-only description, and
`io.webscape.prerelease=true`. GHCR does not have a GitHub Release-style prerelease
flag; the explicit tags and metadata distinguish previews. They never receive
`latest`, production `sha-...` tags, or release version tags. The production image
workflow remains unchanged.

The versioned [webscape-next-task skill](.agents/skills/webscape-next-task/SKILL.md)
instructs the local agent to pull the matching image, run an isolated container,
expose it through Tailscale Serve, and post a verified preview URL and revision on
the PR. Preview links require tailnet access. Each revision gets its own ports;
the agent replaces the link after validation and stops the old preview. After
feature acceptance and a successful merge, it removes only that PR's preview.

This repository copy is the reviewable source for the skill. If using an installed
copy at `~/.codex/skills/webscape-next-task`, synchronize its `SKILL.md` and
`references/pr-previews.md` from the accepted default-branch version after merging
skill changes.

## License

Copyright © 2025–2026 David Atkins.

Webscape is free software licensed under the [GNU Affero General Public
License version 3 only](LICENSE) (`AGPL-3.0-only`). You may use, modify, and
distribute it under that license. It is provided without warranty.

If you run a modified version for users over a network, the AGPL requires you
to offer those users the complete corresponding source code for that version.
Update the in-game source link if your deployment's source is hosted somewhere
other than this repository.

Third-party components retain their own licenses; see
[THIRD_PARTY_LICENSES.txt](THIRD_PARTY_LICENSES.txt).

## Server tick timing

`config.json` sets `server.tickIntervalMs` to **500** for two authoritative ticks per second (also the default when omitted). The server advertises this interval in the `world` message for the client clock. Game action durations are expressed in ticks, so changing the interval changes gameplay speed.

The loop advances a monotonic deadline by one interval per completed tick. If work or scheduling delays leave ticks overdue, it executes full recovery ticks immediately and sequentially, retaining all accumulated debt. Each recovery tick synchronizes state before flushing events. Shutdown is checked between steps. Sustained CPU overload cannot be cured by catch-up alone; the backlog remains visible rather than silently skipping simulation time.

Logs use these searchable event names:

- `tick_loop_started`: configured interval and target ticks per second.
- `tick_overrun`: every update exceeding the interval, with tick number, total duration, start lateness, game-lock wait, systems, synchronization, events, and the slowest system.
- `tick_recovery_started`: overdue time and number of pending ticks, including delays caused by host scheduling.
- `tick_recovery_progress`: backlog and recovery count every five seconds during sustained recovery.
- `tick_recovery_finished`: recovery tick count and elapsed recovery time.

Pathfinding uses one A* search to reach any valid tile within interaction range, then validates and follows cached steps. Tile blocker counts track doors and entity footprints. Each search is limited to 16,384 expanded nodes; routes exceeding this work limit are rejected rather than allowing an unbounded search to stall a tick.
