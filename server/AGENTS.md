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

### Systems

Logic that operates on entities and components. These can be found in `game/system`.

Each system is a struct that implements the `System` interface.

The `Update` method is called once per game tick.

### Game

The `Game` struct is responsible for managing the state of the game.
