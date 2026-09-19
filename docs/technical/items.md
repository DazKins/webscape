# Item definitions and instances

[`itemcatalogue.go`](../../server/game/model/itemcatalogue.go) defines all items in
Go, including equipment, materials, currency and quest items. Add a typed definition
here instead of registering a shop-specific factory, then rebuild the server.
The editor accepts item IDs directly and validates reference structure, counts,
slots and prices. The Go loader validates that IDs exist and are compatible with
their use. There is no editor item catalogue or code generation step.
The playable client receives resolved item fields from the server over WebSocket.
IDs such as `ironSword`
are permanent content and save contracts; changing a display name does not change
an item's identity. Removing or renaming an ID requires a save/content migration.

A definition contains the name, gameplay category (`type`), render model,
equipment slot, combat profile and stack rule. Catalogue lookups return copies so
callers cannot change shared definitions accidentally. `model.NewItem(definitionID)`
creates an instance with a fresh UUID, quantity 1 and no individual properties;
unknown IDs are rejected.

An `Item` stores only `id`, `definitionId`, `quantity` and optional `properties`.
`ItemProperties` is a typed extension point for individual traits. It currently
supports a custom display name and a replacement combat profile. Those overrides
are used by presentation and combat, persisted with the item and deep-copied when
preparing inventory transactions. Add future traits such as durability here when
implementing their gameplay, rather than duplicating the definition on each item.

Ordinary items stack only when their definition permits stacking and their
`definitionId` matches. Items with individual properties stay separate and have
quantity 1. Standard shop offers do not buy customised instances at the base price.

## Content references

Loot, gathering yields and quest rewards use the same reference shape:

```json
{"definitionId":"ironSword","count":1}
```

Equipment references retain the compact slot mapping:

```json
{"equipped":{"slots":{"weapon":"ironSword","offhand":"woodenShield"}}}
```

Shop offers contain a definition ID and prices, without an inventory instance:

```json
{"definitionId":"ironSword","buyPrice":40,"sellPrice":16}
```

Buying creates a fresh instance from the catalogue. Selling, dropping and equipping
still target the UUID of an item actually owned by the player. The existing trade
command's `itemId` payload means a definition ID for `buy` and an instance UUID for
`sell`; prices and definitions are always resolved by the server.

Inventory/equipment state includes `definitionId`. The server also resolves UI
fields (name, model, stats, etc.) at the transport boundary for existing client
consumers. Shop presentation has a `definitionId` but no fake instance UUID.
Icons and shop matching use definition IDs, independently of display names.

Collection emits `collect:definition:<definitionId>` for new quests. Existing
`collect:item:<category>` and `collect:name:<normalised definition name>` event IDs
remain supported; a custom instance name does not change those events.

## Saves and compatibility

Item-owning components, shops and resource references now write component save
version 2. Version 1 snapshots migrate known name/category pairs to catalogue IDs,
preserve UUIDs and quantities, and retain differing equipment stats as individual
properties. Unknown or inconsistent legacy items fail restore with an error
instead of being replaced. The server can still read known legacy name/type
content references. The editor requires `definitionId` references; old name/type
references must be migrated before editing. Legacy shop `itemId` offers remain
readable; new content and editor saves use `definitionId`.
The content fingerprint includes definition values, independently of Go source
formatting or map iteration order.

Run `go test ./...` and build both frontends after changing item contracts.
`cd client && pnpm run model:check` checks registered models, animations and
equipment attachments. The [world](../../schema/world-format.openapi.yaml) and
[quest](../../schema/quest-format.openapi.yaml) schemas describe content references
using string item IDs; item definitions and instance types live in Go.
