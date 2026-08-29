# Webscape Server Coding Agent Instructions

The server is built in Golang and is responsible for:
- Owning the state of the game
- Processing commands from clients
- Sending updates to clients

## ECS Pattern

The core game architecture is based on an ECS (Entity Component System) pattern.

These component and system concepts are managed by the `Game` and `ComponentManager` constructs.

### Components

Data describing the state of an entity. These can be found in `game/component`.

Each component is a struct that implements the `Component` interface.

There is a `SerializeableComponent` interface that can be implemented by components that need to be serialized and sent to clients. The `Serialize` method should return a JSON object that can be sent to the client. If a component does not implement this interface, it will not be serialized and sent to clients. The `Game` will keep track of the previous serialised state of each component and only send updates to clients if the component has changed since the last update.

Serializable components must represent durable, reconstructable game state. If the client only needs to react once to an occurrence, use a domain event and client-event projection instead of a temporary entity, component, or TTL. A useful test is: reconnecting clients need the current state value, but they should not replay an old stateless event.

### Systems

Logic that operates on entities and components. These can be found in `game/system`.

Each system is a struct that implements the `System` interface.

The `Update` method is called once per game tick.

### Game

The `Game` struct is responsible for managing the state of the game.

## Commands, State, And Events

Keep the three WebSocket communication paths distinct:

1. Client commands carry intent into `server/command` and `server/commandhandler.go`; handlers validate and mutate authoritative state.
2. Server state replication comes from per-client `world`, `chunkUpdate`, and delta-only `gameUpdate` messages produced by `Game.syncClient` after each tick.
3. Server stateless events begin in `gameevent.Dispatcher`. Register event-specific subscribers, project only explicit safe DTOs from `server/message`, select recipients by server-side interest, and flush client events after state updates.

Do not serialize an internal `gameevent.Event` directly. Domain events may contain server-only details and may have subscribers other than WebSocket delivery. Current projected event examples are chat, combat resolution, and woodcutting swings. The browser owns their visual lifetime; the ECS does not.
