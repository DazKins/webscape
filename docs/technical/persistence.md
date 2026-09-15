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

The database/user must already exist. On startup the adapter creates
`webscape_worlds` and `webscape_components` if needed, so the user needs schema
CREATE permission and SELECT/INSERT/UPDATE/DELETE permissions on those tables.
Use a direct PostgreSQL connection
or a session-pooling endpoint: a session advisory lock prevents two game servers
from writing the same `worldKey`. Transaction-pooling endpoints are unsupported.
Use different keys/databases for independent worlds and previews. A rolling
replacement must stop the old server before the new server acquires its key.

The game exposes storage-independent snapshot/restore methods. Each component's
`component*_save.go` owns its DTO, versioned decoder registration, `Save` method,
and invariant checks. Shared helpers dispatch through interfaces; restore orchestration
decodes and validates all entities before installing any. The read-only validation
context supplies authored registries and cross-entity item ownership checks without
importing `Game` into components. `server/snapshot` defines only the neutral envelope;
`server/persistence` owns PostgreSQL, schema migrations, delta calculation and atomic
writes. Game systems and command handlers have no database dependency. New components
provide their own save codec or explicitly mark themselves transient.

Each snapshot captures entity IDs, durable components (including server-only fields),
the simulation tick, and offline players. PostgreSQL stores world metadata in
`webscape_worlds` and one row per `(world_key, entity_id, component_id)` in
`webscape_components`, with versioned JSON payloads. The adapter compares against
its last successfully committed baseline and writes only added/changed components
and deletions. All row changes and world metadata commit in one transaction, so an
item transfer cannot be partially saved. Loads read a consistent transaction too.

Saves run after ticks, registered-player disconnects, and graceful shutdown.
Commands never directly trigger persistence: accepted changes are captured by the
next tick or disconnect, and rejected/unknown/unregistered traffic cannot amplify
checkpoint work. Viewer disconnects do not save. Database I/O happens outside the
game mutex; captures and writes are serialized to prevent older snapshots overwriting
newer ones. Startup/load/save failures stop the server without resetting progress or
switching to `none`. Abrupt termination restores the last committed checkpoint.

The smaller rows avoid rewriting unchanged component payloads and allow targeted
inspection/migrations. The tradeoff is more rows/indexes, explicit deletion handling,
and transactional assembly on load. Snapshot capture and comparison still scan the
in-memory world; incremental dirty tracking is a separate future optimization.

Restored players stay outside the active ECS until they reconnect using their
verified OIDC account (issuer and subject). Their name, appearance, items, equipment,
health and quest progress survive; no starter inventory is granted again. Clearing
browser storage requires signing in again but does not lose account ownership.
Legacy anonymous saves remain intact and require an explicit administrative
migration to associate with an account. See [Authentication](authentication.md).
Movement, combat, fishing,
woodcutting, facing targets, conversations and trading sessions resume idle rather
than replaying old actions or events. Resource/spawn countdowns pause while the
server is stopped. Authored terrain and registries still load from `game-project`.

Snapshots and component payloads are versioned and validated before restoration.
A fingerprint of `game.json` and its referenced content files rejects restores
against changed authored content. Back up and explicitly migrate a saved world
when changing that content, or choose a new `worldKey` for a fresh world. The
existing save is never discarded automatically. PostgreSQL backups must include both
`webscape_worlds` and `webscape_components` in the same consistent backup.

SQL schema v2 automatically imports a v1 `webscape_snapshots` row for the selected
world key on first load. The import is transactional and preserves the entity and
component save formats. The original blob remains as a migration backup, marked v2
so an older binary refuses to load stale progress. Subsequent saves use the new tables;
the legacy blob is not a current backup. Existing v1 installations also need SELECT
and UPDATE permission on the legacy table during migration. Do not downgrade to the
old binary after migration.

Run the isolated storage integration tests with:

```sh
WEBSCAPE_TEST_DATABASE_URL='postgres://USER:PASSWORD@127.0.0.1:5432/TEST_DB?sslmode=disable' go test ./... -count=1
```

Integration tests create unique world keys and remove their rows afterward. Use an
isolated test database: rollback tests also install and remove a temporary test trigger.
Without that variable, database integration tests are skipped; snapshot/restore,
configuration and coordinator tests still run under `go test ./...`.

For provider-free testing, see [no-auth testing mode](authentication.md#no-auth-testing-mode) (`auth.mode: "none"` with `server.devMode: true`).
