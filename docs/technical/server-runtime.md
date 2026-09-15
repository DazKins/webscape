# Server runtime

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

## Connection health and inactivity
`server.connections` in `config.json` controls these timeouts (also their defaults when omitted):

```json
{
  "pingIntervalSeconds": 20,
  "pongTimeoutSeconds": 60,
  "idleTimeoutSeconds": 900,
  "idleWarningSeconds": 60
}
```

The server sends WebSocket pings and removes connections that stop answering with pongs.
Registered players are also disconnected after 15 minutes without input, with a warning
and a **Stay connected** button during the final minute. Clicks, key presses, scrolling,
and gameplay commands count as activity; heartbeats and automatic game actions do not.
The server owns the inactivity deadline. A disconnected player can explicitly **Rejoin game**
without signing in again, provided their account session is still valid.

When a character is already active, **Continue here** explicitly replaces its previous
connection. The old page displays **Your game was opened elsewhere** and stops reconnecting.
Inactivity and replacement never revoke account authentication. Their WebSocket close codes
are 4002 and 4003; account-session expiry/revocation uses 4001. The `inactivity` message and
`activity` command carry connection lifecycle information, not ECS state or gameplay events.
