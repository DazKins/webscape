# Webscape

Webscape is a content-driven Go game server with playable and editor
frontends built with Vite, React, and TypeScript.

## Development

Build the playable client before starting the server:

```sh
cd client
npm ci
npm run build
cd ..
go run .
```

The editor runs separately from `editor/` with `npm run dev`.

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
