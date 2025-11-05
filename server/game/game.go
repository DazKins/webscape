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

	entities *util.IdMap[*entity.Entity, model.EntityId]
	systems  []system.System

	prevSerialisedComponents map[model.EntityId]map[component.ComponentId]util.Json
}

var NAMES = []string{"Bob", "Alice", "Charlie", "David", "Eve", "Frank", "George", "Hannah", "Isaac", "Jack"}

func NewGame() *Game {
	world := world.NewWorld(10, 10)

	game := &Game{
		clientIdToEntityId: util.NewBiMap[string, model.EntityId](),
		world:              world,
		done:               make(chan bool),

		entities: util.NewIdMap[*entity.Entity](),
		systems:  []system.System{},

		prevSerialisedComponents: make(map[model.EntityId]map[component.ComponentId]util.Json),
	}

	for i := range 3 {
		position := math.Vec2{X: rand.Intn(10) - 5, Y: rand.Intn(10) - 5}
		for world.GetWall(position.X, position.Y) {
			position = math.Vec2{X: rand.Intn(10) - 5, Y: rand.Intn(10) - 5}
		}

		game.AddEntity(entity.CreateDudeEntity(
			NAMES[i],
			position,
		))
	}

	game.RegisterSystem(&system.PathingSystem{})
	game.RegisterSystem(&system.RandomWalkSystem{
		World: world,
	})
	game.RegisterSystem(&system.ChatMessageSystem{})

	return game
}

func (g *Game) RegisterSystem(system system.System) {
	system.SetEntityGetter(g)
	g.systems = append(g.systems, system)
}

func (g *Game) AddEntity(entity *entity.Entity) {
	g.entities.Put(entity)
}

func (g *Game) GetEntity(entityId model.EntityId) (*entity.Entity, bool) {
	return g.entities.GetById(entityId)
}

func (g *Game) RemoveEntity(entityId model.EntityId) {
	g.entities.DeleteById(entityId)
	g.broadcastMessage(message.NewEntityRemoveMessage(entityId))
}

func (g *Game) GetEntities() []*entity.Entity {
	return g.entities.Values()
}

func (g *Game) GetEntitiesWithComponents(componentIds ...component.ComponentId) []*entity.Entity {
	entities := make([]*entity.Entity, 0)
	for _, entity := range g.entities.Values() {
		match := true
		for _, componentId := range componentIds {
			if entity.GetComponent(componentId) == nil {
				match = false
			}
		}
		if match {
			entities = append(entities, entity)
		}
	}
	return entities
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
	for _, entity := range g.entities.Values() {
		if len(entity.GetComponents().Values()) == 0 {
			g.RemoveEntity(entity.GetId())
		}
	}

	for _, system := range g.systems {
		system.Update()
	}

	updatedComponents := make(map[model.EntityId]map[component.ComponentId]util.Json)

	for _, entity := range g.entities.Values() {
		entitySerialisedComponents := g.prevSerialisedComponents[entity.GetId()]
		if entitySerialisedComponents == nil {
			entitySerialisedComponents = make(map[component.ComponentId]util.Json)
			g.prevSerialisedComponents[entity.GetId()] = entitySerialisedComponents
		}

		for _, comp := range entity.GetComponents().Values() {
			serializeableComp, ok := comp.(component.SerializeableComponent)
			if !ok {
				continue
			}

			serialized := serializeableComp.Serialize()

			if util.JsonEqual(entitySerialisedComponents[comp.GetId()], serialized) {
				continue
			}

			entitySerialisedComponents[comp.GetId()] = serialized

			if updatedComponents[entity.GetId()] == nil {
				updatedComponents[entity.GetId()] = make(map[component.ComponentId]util.Json)
			}
			updatedComponents[entity.GetId()][comp.GetId()] = serialized
		}
	}

	if len(updatedComponents) > 0 {
		g.broadcastMessage(message.NewGameUpdateMessage(updatedComponents))
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

	playerEntity := entity.CreatePlayerEntity(id, name)
	g.AddEntity(playerEntity)

	g.clientIdToEntityId.Put(clientID, id)

	g.sendMessage(clientID, message.NewJoinedMessage(id.String()))
	g.sendMessage(clientID, message.NewWorldMessage(g.world))

	// This seems hacky. Could be race conditions if we try
	// to send the new client the state of all entities while
	// a game tick is in progress. Works for now.
	g.sendMessage(clientID, message.NewGameUpdateMessage(g.prevSerialisedComponents))
}

func (g *Game) HandleMove(clientID string, x int, y int) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		panic("client ID not found in clientIdToEntityId")
	}

	entity, ok := g.GetEntity(entityId)
	if !ok {
		panic("entity not found")
	}

	positionComponent := entity.GetComponent(component.ComponentIdPosition).(*component.CPosition)
	if positionComponent == nil {
		panic("position component not found")
	}

	path, err := g.getPath(positionComponent.Position, math.Vec2{X: x, Y: y})
	if err != nil {
		log.Printf("failed to get path: %v\n", err)
		return
	}

	pathingComponent := &component.CPathing{}
	pathingComponent.Path = &path
	entity.SetComponent(pathingComponent)
}

func (g *Game) HandleLeave(clientID string) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	g.clientIdToEntityId.Delete(clientID)
	g.RemoveEntity(entityId)
}

func (g *Game) HandleChat(clientID string, chatMessage string) {
	entityId, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		log.Println("Client ID not found in clientIdToEntityId")
		return
	}

	chatMessageEntity := entity.CreateChatMessageEntity(entityId, chatMessage)
	g.AddEntity(chatMessageEntity)
}

func (g *Game) HandleInteract(clientID string, entityId model.EntityId, option component.InteractionOption) {
	entity, ok := g.GetEntity(entityId)
	if !ok {
		panic("entity not found")
	}

	interactableComponent := entity.GetComponent(component.ComponentIdInteractable).(*component.CInteractable)
	if interactableComponent == nil {
		panic("interactable component not found")
	}

	for _, interactionOption := range interactableComponent.InteractionOptions {
		if interactionOption == option {
			// TODO
		}
	}

	log.Println("Invalid interaction option")
}
