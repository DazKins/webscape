# Client/server communication

The Go server is authoritative, and the playable client communicates with it through `/ws`. Gameplay traffic follows three deliberately separate patterns:

| Flow | Purpose | Examples | Delivery semantics |
| --- | --- | --- | --- |
| Client → server commands | Express player intent for the server to validate and execute | `move`, `chat`, `interact`, `equip` | A request, never authoritative state or proof that an outcome occurred |
| Server → client state | Replicate the current authoritative truth | `world`, `chunkUpdate`, `gameUpdate` component deltas | Per-client and interest-filtered; reconnect and interest re-entry reconstruct current state |
| Server → client events | Notify eligible clients that something happened once | `chatMessage`, `combatResolved` | Stateless and not replayed; the client owns presentation and expiry |

Commands use `{ "type": ..., "data": ... }`. Server messages use `{ "metadata": { "type": ..., "time": ... }, "data": ... }`.

On each 500 ms game tick, the server runs its ECS systems and then synchronizes every client's visible chunks and entities. Every client receives `gameUpdate` with `serverTick`, even when `entities` is an empty array. The entity array contains only changed serializable components; a `null` component value removes it. If a fact must survive reconnect or interest re-entry—position, health, inventory, quest progress, whether a door is open—it belongs in replicated state.

The initial `world` metadata includes `dayCycle: { dayTicks: 1200, cycleTicks: 2400 }` and `tickIntervalMs`. Clients derive the global day/night phase and countdown from confirmed `gameUpdate.serverTick` values. Sky, lighting, and the minimap dial interpolate only within the current tick and stop before the next unconfirmed tick if updates stall. No client clock can advance authoritative game time.

One-shot occurrences use the server's typed domain-event dispatcher instead. Registered subscribers can independently advance quests or project a safe event DTO to interested clients. Client events are sent after that tick's state deltas, so rendering reacts to the latest authoritative state. Chat bubbles, hit splats, and transient combat animations are client-owned effects. Ongoing fishing and woodcutting animations are selected from their replicated phase and phase-start tick.

Internal domain events and WebSocket messages are intentionally different contracts: the server explicitly chooses recipients and maps approved fields into a wire DTO. Events are not replayed after a disconnect. Specialized `conversation` and `questCompleted` notifications are also non-state messages, although they currently use their own targeted flows. Session-control responses such as `registered` and `registrationFailed` handle connection setup rather than gameplay state.
