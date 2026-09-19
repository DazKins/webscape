# Game persistence

Persistence is independent of `server.devMode`. The checked-in `config.json` uses
`"persistence": {"driver": "none"}`, which retains offline characters in memory for
rejoining and device switching, but loses progress when the server restarts.
For restart testing, use PostgreSQL, the same backend as deployment.
There is no embedded SQLite database.

Deploy PostgreSQL separately and set these options in your mounted `config.json`:

```json
"persistence": {
  "driver": "postgres",
  "worldKey": "webscape-production",
  "timeoutSeconds": 10,
  "postgres": {
    "connectionStringEnv": "WEBSCAPE_DATABASE_URL"
  }
}
```

Set `WEBSCAPE_DATABASE_URL` in the game container's environment to a PostgreSQL
connection URL, for example
`postgres://webscape:PASSWORD@postgres.example.net:5432/webscape?sslmode=verify-full`.
Percent-encode special characters in URL credentials. The value is read at startup;
a missing/empty variable fails startup. Connection strings and passwords are never
logged. PostgreSQL's keyword connection-string syntax is also supported.

Alternatively, configure individual fields:

```json
"persistence": {
  "driver": "postgres",
  "worldKey": "webscape-production",
  "timeoutSeconds": 10,
  "postgres": {
    "host": "postgres.example.net",
    "port": 5432,
    "database": "webscape",
    "user": "webscape",
    "passwordEnv": "WEBSCAPE_DATABASE_PASSWORD",
    "sslMode": "verify-full",
    "sslRootCert": "/certs/postgres-ca.pem"
  }
}
```

`port` defaults to 5432, `sslMode` to `verify-full`, `worldKey` to `default`, and
`timeoutSeconds` to 10. `sslRootCert` is optional when the issuing CA is already
trusted. Optional `sslCert` and `sslKey` support client certificates; certificate
paths must be readable inside the game container. `password` supports an inline
password, but cannot be combined with `passwordEnv`. `connectionStringEnv`, when
set, overrides the individual connection fields, including TLS settings. Local
isolated database tests can explicitly use `sslMode: "disable"`.

The database/user must already exist. The adapter creates `webscape_player_saves`
and `webscape_players` at startup, requiring schema CREATE and table
SELECT/INSERT/UPDATE/DELETE permissions. Use a direct connection or session-pooling
endpoint: a session advisory lock prevents two servers writing the same `worldKey`.
Transaction-pooling endpoints are unsupported. Use separate keys for previews.
Stop the old server before its replacement acquires the same key.

Only players persist. `webscape_players` contains one row per `(world_key, player_id)`
with versioned durable components in a `components` JSONB object and an `updated_at`
timestamp. Inventory, bank, equipment, identity, stats, position, and quest progress
stay together. `webscape_player_saves` stores only the envelope version and last
checkpoint timestamp for each key; it stores no world state or simulation clock.

Every restart builds the world from current `game-project` content. NPCs, doors,
chests, resources, and spawns return to authored defaults. Ground drops disappear.
The simulation clock resets to zero and day/night starts at daybreak. Map additions
appear immediately after restart without reconciling saved world entities.

Restored players stay outside the active ECS until they reconnect using their
verified account. Offline players are also captured in saves. Absence from a
checkpoint does not delete an existing player row. Starter inventory is not granted
again. If a saved position is missing or blocked in the current map, the player is
moved to spawn. Movement, combat, banking, trading, and other transient activities
reset. Clearing browser storage does not remove account progress. Legacy anonymous
characters still require explicit account association; see [Authentication](authentication.md).

The game captures player components under its mutex. Component codecs own versioned
payloads and validation; all players are validated before any are installed, including
cross-player item ownership checks. Only components attached to players need save
support. Game systems and commands have no database dependency.

The adapter compares player records against its last successful baseline and writes
only changed rows. Each checkpoint is a transaction, preserving atomic changes across
players too. Loads use a consistent transaction. Saves follow ticks, registered-player
disconnects, and graceful shutdown. Captures and writes are serialized; database I/O
runs outside the game mutex. Failures stop the server instead of resetting progress.
Abrupt termination restores the latest committed checkpoint.

Only the current player-save format (version 2) is supported. Legacy world snapshots
and their migration readers have been retired after all environments migrated.
Older deployments must upgrade through the player-persistence migration release
(commit `1326e0e`) before using this version.

The legacy `webscape_components`, `webscape_worlds`, and `webscape_snapshots` tables
are no longer read or written. They may be removed separately after confirming the
migration for every world key sharing the database and retaining any desired backup.
The runtime does not drop these tables automatically.

Back up before upgrading. Current backups must include `webscape_player_saves` and
`webscape_players` together. Map edits no longer require a world-save migration;
changes to item definitions or quest contracts still need compatible player data.

Run the isolated storage integration tests with:

```sh
WEBSCAPE_TEST_DATABASE_URL='postgres://USER:PASSWORD@127.0.0.1:5432/TEST_DB?sslmode=disable' go test ./... -count=1
```

Integration tests create unique world keys and remove their rows afterward. Use an
isolated test database: rollback tests also install and remove a temporary test constraint.
Without that variable, database integration tests are skipped; snapshot/restore,
configuration and coordinator tests still run under `go test ./...`.

For provider-free testing, see [no-auth testing mode](authentication.md#no-auth-testing-mode) (`auth.mode: "none"` with `server.devMode: true`).
