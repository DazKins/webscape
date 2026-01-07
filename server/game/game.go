package game

import (
	"log"
	"math/rand"
	"time"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/model"
	"webscape/server/game/system"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
	"webscape/server/util"
)

type MessageBroadcaster func(message message.Message)
type MessageSender func(clientID string, message message.Message)

type Game struct {
	world              *world.World
	clientIdToEntityId *util.BiMap[string, model.EntityId]
	ticker             *time.Ticker
	done               chan bool
	sendMessage        MessageSender
	broadcastMessage   MessageBroadcaster

	systems []system.System

	componentManager         *component.ComponentManager
	prevSerialisedComponents map[component.ComponentId]map[model.EntityId]util.Json
}

var NAMES = []string{"Bob", "Alice", "Charlie", "David", "Eve", "Frank", "George", "Hannah", "Isaac", "Jack"}

func NewGame() *Game {
	world := world.NewWorld(10, 10)

	game := &Game{
		clientIdToEntityId: util.NewBiMap[string, model.EntityId](),
		world:              world,
		done:               make(chan bool),

		componentManager: component.NewComponentManager(),
		systems:          []system.System{},

		prevSerialisedComponents: make(map[component.ComponentId]map[model.EntityId]util.Json),
	}

	for i := range 3 {
		position := math.Vec2{X: rand.Intn(10) - 5, Y: rand.Intn(10) - 5}
		for world.GetWall(position.X, position.Y) {
			position = math.Vec2{X: rand.Intn(10) - 5, Y: rand.Intn(10) - 5}
		}

		components := entity.CreateDudeEntity(
			NAMES[i],
			position,
		)
		game.componentManager.CreateNewEntity(components...)
	}

	game.RegisterSystem(&system.PathingSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		World: world,
	})
	game.RegisterSystem(&system.InteractionSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		ChatMessageSender: game,
	})
	game.RegisterSystem(&system.RandomWalkSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
		World: world,
	})
	game.RegisterSystem(&system.TtlSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
	})
	game.RegisterSystem(&system.HealthSystem{
		SystemBase: system.SystemBase{
			ComponentManager: game.componentManager,
		},
	})

	return game
}

func (g *Game) RegisterSystem(system system.System) {
	g.systems = append(g.systems, system)
}

func (g *Game) GetEntitiesWithComponents(componentIds ...component.ComponentId) []model.EntityId {
	if len(componentIds) == 0 {
		return []model.EntityId{}
	}

	// Count how many of the requested components each entity has
	entityCounts := make(map[model.EntityId]int)
	for _, componentId := range componentIds {
		for entityId := range g.componentManager.GetComponent(componentId) {
			entityCounts[entityId]++
		}
	}

	// Only keep entities that have all requested components
	result := make([]model.EntityId, 0)
	requiredCount := len(componentIds)
	for entityId, count := range entityCounts {
		if count == requiredCount {
			result = append(result, entityId)
		}
	}

	return result
}

func (g *Game) StartUpdateLoop() {
	g.ticker = time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-g.ticker.C:
				g.update()
			case <-g.done:
				g.ticker.Stop()
				return
			}
		}
	}()
}

func (g *Game) update() {
	for _, system := range g.systems {
		system.Update()
	}

	updatedComponents := make(map[component.ComponentId]map[model.EntityId]util.Json)

	for componentId, entities := range g.componentManager.GetAllComponents() {
		prevSerialisedEntities := g.prevSerialisedComponents[componentId]
		if prevSerialisedEntities == nil {
			prevSerialisedEntities = make(map[model.EntityId]util.Json)
			g.prevSerialisedComponents[componentId] = prevSerialisedEntities
		}

		for entityId, comp := range entities {
			serializeableComponent, ok := comp.(component.SerializeableComponent)
			if !ok {
				continue
			}

			serialized := serializeableComponent.Serialize()

			if util.JsonEqual(prevSerialisedEntities[entityId], serialized) {
				continue
			}

			prevSerialisedEntities[entityId] = serialized

			if updatedComponents[componentId] == nil {
				updatedComponents[componentId] = make(map[model.EntityId]util.Json)
			}
			updatedComponents[componentId][entityId] = serialized
		}
	}

	removedComponents := make(map[component.ComponentId][]model.EntityId)
	removedEntities := make(map[model.EntityId]bool)

	for componentId, prevSerialisedEntities := range g.prevSerialisedComponents {
		for entityId := range prevSerialisedEntities {
			if g.componentManager.GetEntityComponent(componentId, entityId) == nil {
				removedComponents[componentId] = append(removedComponents[componentId], entityId)
				delete(prevSerialisedEntities, entityId)
				removedEntities[entityId] = true
			}
		}
	}

	// Check if any entities have been completely removed (no components left)
	completelyRemovedEntities := make([]model.EntityId, 0)
	for entityId := range removedEntities {
		// Check if entity has any components left
		hasComponents := false
		for _, components := range g.componentManager.GetAllComponents() {
			if _, exists := components[entityId]; exists {
				hasComponents = true
				break
			}
		}
		if !hasComponents {
			completelyRemovedEntities = append(completelyRemovedEntities, entityId)
		}
	}

	if len(updatedComponents) > 0 || len(removedComponents) > 0 {
		g.broadcastMessage(message.NewGameUpdateMessage(updatedComponents, removedComponents))
	}

	// Send entity removal messages for completely removed entities
	for _, entityId := range completelyRemovedEntities {
		g.broadcastMessage(message.NewEntityRemoveMessage(entityId))
	}
}

func (g *Game) Stop() {
	g.done <- true
}

func (g *Game) RegisterBroadcaster(messageBroadcaster MessageBroadcaster) {
	g.broadcastMessage = messageBroadcaster
}

func (g *Game) RegisterSender(messageSender MessageSender) {
	g.sendMessage = messageSender
}

func (g *Game) HandleJoin(clientID string, id model.EntityId, name string) {
	rand.Seed(time.Now().UnixNano())

	if _, ok := g.clientIdToEntityId.Get(clientID); ok {
		g.sendMessage(clientID, message.NewJoinFailedMessage("you are already connected in another session"))
		return
	}

	components := entity.CreatePlayerEntity(id, name)
	g.componentManager.SetEntityComponents(id, components...)

	g.clientIdToEntityId.Put(clientID, id)

	g.sendMessage(clientID, message.NewJoinedMessage(id.String()))
	g.sendMessage(clientID, message.NewWorldMessage(g.world))

	// This seems hacky. Could be race conditions if we try
	// to send the new client the state of all entities while
	// a game tick is in progress. Works for now.
	g.sendMessage(clientID, message.NewGameUpdateMessage(
		g.prevSerialisedComponents,
		make(map[component.ComponentId][]model.EntityId),
	))
}

func (g *Game) HandleMove(clientID string, x int, y int) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		panic("client ID not found in clientIdToEntityId")
	}

	positionComponent := g.componentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)
	if !ok || positionComponent == nil {
		panic("position component not found")
	}

	pathingComponent := component.NewCPathing(component.PathingTarget{
		Position: util.OptionalSome(math.Vec2{X: x, Y: y}),
	})
	g.componentManager.SetEntityComponent(entityId, pathingComponent)
}

func (g *Game) HandleLeave(clientID string) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	g.componentManager.RemoveEntity(entityId)
	g.clientIdToEntityId.Delete(clientID)
}

// SendChatMessageEntityFor sends a new chat message entity for the given entity,
// removing any existing chat messages from that entity first to ensure only one
// chat message exists per entity at a time.
func (g *Game) SendChatMessageEntityFor(fromEntityId model.EntityId, message string) model.EntityId {
	// Remove any existing chat messages from this entity
	chatMessageEntities := g.componentManager.GetComponent(component.ComponentIdChatMessage)
	for existingEntityId, comp := range chatMessageEntities {
		chatMessageComp := comp.(*component.CChatMessage)
		if chatMessageComp.GetFromEntityId() == fromEntityId {
			g.componentManager.RemoveEntity(existingEntityId)
		}
	}

	// Create the new chat message entity
	chatMessageComponents := entity.CreateChatMessageEntity(fromEntityId, message)
	return g.componentManager.CreateNewEntity(chatMessageComponents...)
}

func (g *Game) HandleChat(clientID string, chatMessage string) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	g.SendChatMessageEntityFor(entityId, chatMessage)
}

func (g *Game) HandleInteract(clientID string, entityId model.EntityId, option component.InteractionOption) {
	interactableComponent := g.componentManager.GetEntityComponent(component.ComponentIdInteractable, entityId).(*component.CInteractable)
	if interactableComponent == nil {
		panic("interactable component not found")
	}

	// TODO check actual options in the future
	// for _, interactionOption := range interactableComponent.InteractionOptions {
	// 	if interactionOption == option {
	// 	}
	// }

	interactingEntityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	// Set pathing component to path to the target entity
	pathingComponent := component.NewCPathing(component.PathingTarget{
		EntityId: util.OptionalSome(entityId),
	})
	g.componentManager.SetEntityComponent(interactingEntityId, pathingComponent)

	// Set interacting component to track the interaction
	interactingComponent := component.NewCInteracting(entityId, option)
	g.componentManager.SetEntityComponent(interactingEntityId, interactingComponent)
}
