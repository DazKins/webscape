# Technical documentation

Implementation, infrastructure, tooling, and validation notes.

- [Development](development.md): local setup, builds, validation, and the client build identifier. Start here to run the project.
- [Client/server communication](communication.md): commands, state replication, domain events, and message ordering.
- [Server runtime](server-runtime.md): tick timing, recovery, connection health, and inactivity.
- [Authentication](authentication.md): provider configuration, development modes, sessions, and saved characters.
- [Items](items.md): shared definitions, instance properties, content references and save migration.
- [World content](world-content.md): authored equipment and world validation.
- [Persistence](persistence.md): PostgreSQL setup, snapshots, restore behavior, migrations, and integration tests.
- [Deployment](deployment.md): published Docker images and Coolify deployment automation.
- [Licensing](licensing.md): project license, deployment source obligations, and third-party notices.

Paths and shell commands in these guides are relative to the repository root unless stated otherwise. Runtime settings live in root `config.json`; local development can select `config.dev.json` through `WEBSCAPE_CONFIG`.

Put character, world, and story notes in [game](../game/README.md).
